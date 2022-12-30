package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/groupcache/lru"
	"github.com/sisu-network/deyes/chains"
	deyesethtypes "github.com/sisu-network/deyes/chains/eth/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	libchain "github.com/sisu-network/lib/chain"
	"github.com/sisu-network/lib/log"

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
	cfg          config.Chain
	client       EthClient
	blockTime    int
	db           database.Database
	txsCh        chan *types.Txs
	txTrackCh    chan *chainstypes.TrackUpdate
	vault        string
	lock         *sync.RWMutex
	txTrackCache *lru.Cache
	gasCal       *gasCalculator

	// Block fetcher
	blockCh      chan *ethtypes.Block
	blockFetcher *defaultBlockFetcher

	// Receipt fetcher
	receiptFetcher    receiptFetcher
	receiptResponseCh chan *txReceiptResponse
}

func NewWatcher(db database.Database, cfg config.Chain, txsCh chan *types.Txs,
	txTrackCh chan *chainstypes.TrackUpdate, client EthClient) chains.Watcher {
	blockCh := make(chan *ethtypes.Block)
	receiptResponseCh := make(chan *txReceiptResponse)

	w := &Watcher{
		receiptResponseCh: receiptResponseCh,
		blockCh:           blockCh,
		blockFetcher:      newBlockFetcher(cfg, blockCh, client),
		receiptFetcher:    newReceiptFetcher(receiptResponseCh, client, cfg.Chain),
		db:                db,
		cfg:               cfg,
		txsCh:             txsCh,
		txTrackCh:         txTrackCh,
		blockTime:         cfg.BlockTime,
		client:            client,
		lock:              &sync.RWMutex{},
		txTrackCache:      lru.New(TxTrackCacheSize),
		gasCal:            newGasCalculator(cfg, client, GasPriceUpdateInterval),
	}

	return w
}

func (w *Watcher) init() {
	vaults, err := w.db.GetVaults(w.cfg.Chain)
	if err != nil {
		panic(err)
	}

	if len(vaults) > 0 {
		w.vault = vaults[0]
		log.Infof("Saved gateway in the db for chain %s is %s", w.cfg.Chain, w.vault)
	} else {
		log.Infof("Vault for chain %s is not set yet", w.cfg.Chain)
	}
}

func (w *Watcher) SetVault(addr string, token string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	log.Verbosef("Setting vault for chain %s with address %s", w.cfg.Chain, addr)
	err := w.db.SetVault(w.cfg.Chain, addr, token)
	if err == nil {
		w.vault = strings.ToLower(addr)
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

		// Pass this block to the receipt fetcher
		log.Info(w.cfg.Chain, " Block length = ", len(block.Transactions()))
		txs, averageTip := w.processBlock(block)
		log.Info(w.cfg.Chain, " Filtered txs = ", len(txs))

		if w.cfg.UseGasEip1559 {
			w.gasCal.AddNewBlock(block)
		}

		if len(txs) > 0 {
			w.receiptFetcher.fetchReceipts(block.Number().Int64(), txs, block.BaseFee(), averageTip)
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
		Chain:       w.cfg.Chain,
		Block:       response.blockNumber,
		BlockHash:   response.blockHash,
		Arr:         arr,
		BaseFee:     response.baseFee,
		PriorityFee: response.priorityFee,
	}
}

func (w *Watcher) getSuggestedGasPrice() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RpcTimeOut)
	defer cancel()

	return w.client.SuggestGasPrice(ctx)
}

func (w *Watcher) processBlock(block *ethtypes.Block) ([]*ethtypes.Transaction, *big.Int) {
	ret := make([]*ethtypes.Transaction, 0)

	totalTip := big.NewInt(0)
	count := 0

	for _, tx := range block.Transactions() {
		// Check tx gas base fee
		switch tx.Type() {
		case ethtypes.DynamicFeeTxType:
			tipFee := tx.GasTipCap()
			totalTip = totalTip.Add(totalTip, tipFee)
			count++
		default:

		}

		if _, ok := w.txTrackCache.Get(tx.Hash().String()); ok {
			ret = append(ret, tx)
			continue
		}

		if w.acceptTx(tx) {
			ret = append(ret, tx)
		}
	}

	averageTip := big.NewInt(0)
	if count > 0 {
		averageTip = new(big.Int).Div(totalTip, big.NewInt(int64(count)))
	}

	return ret, averageTip
}

func (w *Watcher) acceptTx(tx *ethtypes.Transaction) bool {
	if tx.To() != nil {
		if strings.EqualFold(tx.To().String(), w.vault) {
			return true
		}
	}

	return false
}

func (w *Watcher) getFromAddress(chain string, tx *ethtypes.Transaction) (common.Address, error) {
	signer := libchain.GetEthChainSigner(chain)
	if signer == nil {
		return common.Address{}, fmt.Errorf("cannot find signer for chain %s", chain)
	}

	msg, err := tx.AsMessage(ethtypes.NewLondonSigner(tx.ChainId()), nil)
	if err != nil {
		return common.Address{}, err
	}

	return msg.From(), nil
}

func (w *Watcher) GetNonce(address string) int64 {
	cAddr := common.HexToAddress(address)
	nonce, err := w.client.PendingNonceAt(context.Background(), cAddr)
	if err == nil {
		return int64(nonce)
	} else {
		log.Error("cannot get nonce of chain", w.cfg.Chain, " at", address)
	}

	return 0
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

func (w *Watcher) getTransactionReceipt(txHash common.Hash) (*ethtypes.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RpcTimeOut)
	defer cancel()
	receipt, err := w.client.TransactionReceipt(ctx, txHash)

	if err == nil {
		return receipt, nil
	}

	return nil, fmt.Errorf("Cannot find receipt for tx hash: %s", txHash.String())
}

func (w *Watcher) GetGasInfo() deyesethtypes.GasInfo {
	if w.cfg.UseGasEip1559 {
		return deyesethtypes.GasInfo{
			GasPrice: w.gasCal.GetGasPrice().Int64(),
		}
	} else {
		return deyesethtypes.GasInfo{
			BaseFee: w.gasCal.GetBaseFee().Int64(),
			Tip:     w.gasCal.GetTip().Int64(),
		}
	}
}
