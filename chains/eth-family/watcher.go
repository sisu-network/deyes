package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/groupcache/lru"
	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
	libchain "github.com/sisu-network/lib/chain"
	"github.com/sisu-network/lib/log"
	"go.uber.org/atomic"

	chainstypes "github.com/sisu-network/deyes/chains/types"
)

const (
	minGasPrice      = 10_000_000_000
	TxTrackCacheSize = 1_000
)

type GasPriceGetter func(ctx context.Context) (*big.Int, error)

type BlockHeightExceededError struct {
	ChainHeight uint64
}

func NewBlockHeightExceededError(chainHeight uint64) error {
	return &BlockHeightExceededError{
		ChainHeight: chainHeight,
	}
}

func (e *BlockHeightExceededError) Error() string {
	return fmt.Sprintf("Our block height is higher than chain's height. Chain height = %d", e.ChainHeight)
}

// TODO: Move this to the chains package.
type Watcher struct {
	cfg             config.Chain
	clients         []EthClient
	blockTime       int
	db              database.Database
	txsCh           chan *types.Txs
	txTrackCh       chan *chainstypes.TrackUpdate
	gateway         string
	gasPrice        *atomic.Int64
	gasPriceGetters []GasPriceGetter
	lock            *sync.RWMutex
	txTrackCache    *lru.Cache

	// Block fetcher
	blockCh      chan *etypes.Block
	blockFetcher *defaultBlockFetcher

	// Receipt fetcher
	receiptFetcher    receiptFetcher
	receiptResponseCh chan *txReceiptResponse
}

func NewWatcher(db database.Database, cfg config.Chain, txsCh chan *types.Txs,
	txTrackCh chan *chainstypes.TrackUpdate, clients []EthClient) chains.Watcher {
	blockCh := make(chan *etypes.Block)
	receiptResponseCh := make(chan *txReceiptResponse)

	w := &Watcher{
		receiptResponseCh: receiptResponseCh,
		blockCh:           blockCh,
		blockFetcher:      newBlockFetcher(cfg, blockCh, clients),
		receiptFetcher:    newReceiptFetcher(receiptResponseCh, clients, cfg.Chain),
		db:                db,
		cfg:               cfg,
		txsCh:             txsCh,
		txTrackCh:         txTrackCh,
		blockTime:         cfg.BlockTime,
		gasPrice:          atomic.NewInt64(0),
		clients:           clients,
		lock:              &sync.RWMutex{},
		txTrackCache:      lru.New(TxTrackCacheSize),
	}

	gasPriceGetters := []GasPriceGetter{w.getGasPriceFromNode}
	w.gasPriceGetters = gasPriceGetters
	return w
}

func (w *Watcher) init() {
	var err error
	w.gateway, err = w.db.GetGateway(w.cfg.Chain)
	if err != nil {
		panic(err)
	}

	log.Infof("Saved gateway in the db for chain %s is %s", w.cfg.Chain, w.gateway)
}

func (w *Watcher) SetVault(addr string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	err := w.db.SetVault(w.cfg.Chain, addr)
	if err == nil {
		w.gateway = strings.ToLower(addr)
	} else {
		log.Error("Failed to save gateway")
	}
}

func (w *Watcher) Start() {
	log.Info("Starting Watcher...")

	w.init()
	go w.scanBlocks()
}

func (w *Watcher) scanBlocks() {
	go w.blockFetcher.start()
	go w.receiptFetcher.start()

	go w.waitForBlock()
	go w.waitForReceipt()
}

// waitForBlock waits for new blocks from the block fetcher. It then filters interested txs and
// passes that to receipt fetcher to fetch receipt.
func (w *Watcher) waitForBlock() {
	for {
		block := <-w.blockCh

		// Only update gas price at deterministic block height
		// Ex: updateBlockHeight = startBlockHeight + (n * interval) (n is an integer from 0 ... )
		chainParams := config.ChainParamsMap[w.cfg.Chain]
		if (block.Number().Int64()-chainParams.GasPriceStartBlockHeight)%chainParams.Interval == 0 {
			go func() {
				gasPrice := w.GetGasPrice()
				if gasPrice == 0 {
					w.updateGasPrice(context.Background())
				}
			}()
		}

		// Pass this block to the receipt fetcher
		log.Info(w.cfg.Chain, " Block length = ", len(block.Transactions()))
		txs := w.processBlock(block)
		log.Info(w.cfg.Chain, " Filtered txs = ", len(txs))

		if len(txs) > 0 {
			w.receiptFetcher.fetchReceipts(block.Number().Int64(), txs)
		}
	}
}

// waitForReceipt waits for receipts returned by the fetcher.
func (w *Watcher) waitForReceipt() {
	for {
		response := <-w.receiptResponseCh
		txs := w.extractTxs(response)

		log.Verbose(w.cfg.Chain, ": txs sizes = ", len(txs.Arr))

		if len(txs.Arr) > 0 {
			// Send list of interested txs back to the listener.
			w.txsCh <- txs
		}

		// Save all txs into database for later references.
		w.db.SaveTxs(w.cfg.Chain, response.blockNumber, txs)
	}
}

// extractTxs takes resposne from the receipt fetcher and converts them into deyes transactions.
func (w *Watcher) extractTxs(response *txReceiptResponse) *types.Txs {
	arr := make([]*types.Tx, 0)
	for i, tx := range response.txs {
		receipt := response.receipts[i]
		bz, err := tx.MarshalBinary()
		if err != nil {
			log.Error("Cannot serialize ETH tx, err = ", err)
			continue
		}

		if _, ok := w.txTrackCache.Get(tx.Hash().String()); ok {
			// Get Tx Receipt
			result := chainstypes.TrackResultConfirmed
			if receipt.Status == 0 {
				result = chainstypes.TrackResultFailure
			}

			// This is a transaction that we are tracking. Inform Sisu about this.
			w.txTrackCh <- &chainstypes.TrackUpdate{
				Chain:       w.cfg.Chain,
				Bytes:       bz,
				Hash:        tx.Hash().String(),
				BlockHeight: response.blockNumber,
				Result:      result,
			}

			continue
		}

		var to string
		if tx.To() == nil {
			to = ""
		} else {
			to = tx.To().String()
		}

		from, err := w.getFromAddress(w.cfg.Chain, tx)
		if err != nil {
			log.Errorf("cannot get from address for tx %s on chain %s, err = %v", tx.Hash().String(), w.cfg.Chain, err)
			continue
		}

		arr = append(arr, &types.Tx{
			Hash:       tx.Hash().String(),
			Serialized: bz,
			From:       from.Hex(),
			To:         to,
			Success:    receipt.Status == 1,
		})
	}

	return &types.Txs{
		Chain:     w.cfg.Chain,
		Block:     response.blockNumber,
		BlockHash: response.blockHash,
		Arr:       arr,
	}
}

func (w *Watcher) getSuggestedGasPrice() (*big.Int, error) {
	for _, client := range w.clients {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(w.blockTime)*2*time.Millisecond)
		gasPrice, err := client.SuggestGasPrice(ctx)
		cancel()

		if err == nil {
			return gasPrice, nil
		}
	}

	return nil, fmt.Errorf("Gas price not found")
}

func (w *Watcher) processBlock(block *etypes.Block) []*etypes.Transaction {
	ret := make([]*etypes.Transaction, 0)

	for _, tx := range block.Transactions() {
		if _, ok := w.txTrackCache.Get(tx.Hash().String()); ok {
			ret = append(ret, tx)
			continue
		}

		if w.acceptTx(tx) {
			ret = append(ret, tx)
		}
	}

	return ret
}

func (w *Watcher) acceptTx(tx *etypes.Transaction) bool {
	if tx.To() != nil {
		to := strings.ToLower(tx.To().String())
		if to == w.gateway {
			return true
		}
	}

	return false
}

func (w *Watcher) getFromAddress(chain string, tx *etypes.Transaction) (common.Address, error) {
	signer := libchain.GetEthChainSigner(chain)
	if signer == nil {
		return common.Address{}, fmt.Errorf("cannot find signer for chain %s", chain)
	}

	msg, err := tx.AsMessage(etypes.NewLondonSigner(tx.ChainId()), nil)
	if err != nil {
		return common.Address{}, err
	}

	return msg.From(), nil
}

func (w *Watcher) GetNonce(address string) int64 {
	cAddr := common.HexToAddress(address)
	for _, client := range w.clients {
		nonce, err := client.PendingNonceAt(context.Background(), cAddr)
		if err == nil {
			return int64(nonce)
		} else {
			log.Error("cannot get nonce of chain", w.cfg.Chain, " at", address)
		}
	}

	return 0
}

func (w *Watcher) GetGasPrice() int64 {
	return w.gasPrice.Load()
}

func (w *Watcher) updateGasPrice(ctx context.Context) error {
	potentialGasPriceList := make([]*big.Int, 0)
	for _, getter := range w.gasPriceGetters {
		gasPrice, err := getter(ctx)
		if err != nil {
			return err
		}

		potentialGasPriceList = append(potentialGasPriceList, gasPrice)
	}

	medianGasPrice := utils.GetMedianBigInt(potentialGasPriceList)
	w.gasPrice.Store(medianGasPrice.Int64())

	return nil
}

func (w *Watcher) getGasPriceFromNode(ctx context.Context) (*big.Int, error) {
	gasPrice, err := w.getSuggestedGasPrice()
	if err != nil {
		log.Error("error when getting gas price", err)
		return big.NewInt(0), err
	}

	return gasPrice, nil
}

func (w *Watcher) TrackTx(txHash string) {
	log.Verbose("Tracking tx: ", txHash)
	w.txTrackCache.Add(txHash, true)
}

func (w *Watcher) getTransactionReceipt(txHash common.Hash) (*etypes.Receipt, error) {
	for _, client := range w.clients {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		receipt, err := client.TransactionReceipt(ctx, txHash)
		cancel()

		if err == nil {
			return receipt, nil
		}
	}

	return nil, fmt.Errorf("Cannot find receipt for tx hash: %s", txHash.String())
}
