package core

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
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
	blockHeight     int64
	blockTime       int
	db              database.Database
	txsCh           chan *types.Txs
	txTrackCh       chan *chainstypes.TrackUpdate
	chainAccount    string
	gateway         string
	signers         map[string]etypes.Signer
	gasPrice        *atomic.Int64
	gasPriceGetters []GasPriceGetter
	lock            *sync.RWMutex
	txTrackCache    *lru.Cache
}

func NewWatcher(db database.Database, cfg config.Chain, txsCh chan *types.Txs,
	txTrackCh chan *chainstypes.TrackUpdate, clients []EthClient) chains.Watcher {
	w := &Watcher{
		db:           db,
		cfg:          cfg,
		txsCh:        txsCh,
		txTrackCh:    txTrackCh,
		blockTime:    cfg.BlockTime,
		gasPrice:     atomic.NewInt64(0),
		clients:      clients,
		lock:         &sync.RWMutex{},
		txTrackCache: lru.New(TxTrackCacheSize),
	}

	gasPriceGetters := []GasPriceGetter{w.getGasPriceFromNode}
	w.gasPriceGetters = gasPriceGetters
	return w
}

func (w *Watcher) init() {
	// Set the block height for the watcher.
	w.setBlockHeight()

	var err error
	w.gateway, err = w.db.GetGateway(w.cfg.Chain)
	if err != nil {
		panic(err)
	}

	w.chainAccount, err = w.db.GetChainAccount(w.cfg.Chain)
	if err != nil {
		panic(err)
	}

	log.Infof("Saved gateway in the db for chain %s is %s", w.cfg.Chain, w.gateway)
}

func (w *Watcher) setBlockHeight() {
	for {
		number, err := w.getBlockNumber()
		if err != nil {
			log.Errorf("cannot get latest block number for chain %s. Sleeping for a few seconds", w.cfg.Chain)
			time.Sleep(time.Second * 5)
			continue
		}

		w.blockHeight = int64(number)
		break
	}

	log.Info("Watching from block", w.blockHeight, " for chain ", w.cfg.Chain)
}

func (w *Watcher) SetChainAccount(addr string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	err := w.db.SetChainAccount(w.cfg.Chain, addr)
	if err == nil {
		w.chainAccount = strings.ToLower(addr)
	} else {
		log.Error("Failed to save gateway")
	}
}

func (w *Watcher) SetGateway(addr string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	err := w.db.SetGateway(w.cfg.Chain, addr)
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
	latestBlock, err := w.getLatestBlock()
	if err != nil {
		log.Error("Failed to scan blocks, err = ", err)
	}

	if latestBlock != nil {
		w.blockHeight = latestBlock.Header().Number.Int64()
	}
	log.Info(w.cfg.Chain, " Latest height = ", w.blockHeight)

	for {
		log.Verbose("Block time on chain ", w.cfg.Chain, " is ", w.blockTime)

		// Only update gas price at deterministic block height
		// Ex: updateBlockHeight = startBlockHeight + (n * interval) (n is an integer from 0 ... )
		chainParams := config.ChainParamsMap[w.cfg.Chain]
		if (w.blockHeight-chainParams.GasPriceStartBlockHeight)%chainParams.Interval == 0 {
			go func() {
				gasPrice := w.GetGasPrice()
				if gasPrice == 0 {
					w.updateGasPrice(context.Background())
				}
			}()
		}

		// Get the blockheight
		block, err := w.tryGetBlock()
		if err != nil || block == nil {
			if _, ok := err.(*BlockHeightExceededError); !ok && err != ethereum.NotFound {
				// This err is not ETH not found or our custom error.
				log.Error("Cannot get block at height", w.blockHeight, "for chain", w.cfg.Chain, " err = ", err)
			}

			w.blockTime = w.blockTime + w.cfg.AdjustTime
			time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
			continue
		}

		w.blockTime = w.blockTime - w.cfg.AdjustTime/4
		filteredTxs, err := w.processBlock(block)
		if err != nil {
			log.Error("cannot process block, err = ", err)
			continue
		}

		log.Verbose("Filtered txs sizes = ", len(filteredTxs.Arr), " on chain ", w.cfg.Chain)

		if len(filteredTxs.Arr) > 0 {
			// Send list of interested txs back to the listener.
			w.txsCh <- filteredTxs
		}

		// Save all txs into database for later references.
		w.db.SaveTxs(w.cfg.Chain, w.blockHeight, filteredTxs)

		w.blockHeight++

		time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
	}
}

// Get block with retry when block is not mined yet.
func (w *Watcher) tryGetBlock() (*etypes.Block, error) {
	number, err := w.getBlockNumber()
	if err != nil {
		return nil, err
	}

	if number < uint64(w.blockHeight) {
		return nil, NewBlockHeightExceededError(number)
	}

	block, err := w.getBlock(w.blockHeight)
	switch err {
	case nil:
		log.Verbose(w.cfg.Chain, " Height = ", block.Number())
		return block, nil

	case ethereum.NotFound:
		// Sleep a few seconds and to get the block again.
		time.Sleep(time.Duration(utils.MinInt(w.blockTime/4, 3000)) * time.Millisecond)
		block, err = w.getBlock(w.blockHeight)

		// Extend the wait time a little bit more
		w.blockTime = w.blockTime + w.cfg.AdjustTime
		log.Verbose("New blocktime: ", w.blockTime)
	}

	return block, err
}

func (w *Watcher) getBlockNumber() (uint64, error) {
	for _, client := range w.clients {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(w.blockTime)*2*time.Millisecond)
		number, err := client.BlockNumber(ctx)
		cancel()

		if err == nil {
			return number, nil
		}
	}

	return 0, fmt.Errorf("Block number not found")
}

func (w *Watcher) getLatestBlock() (*etypes.Block, error) {
	return w.getBlock(-1)
}

func (w *Watcher) getBlock(height int64) (*etypes.Block, error) {
	for _, client := range w.clients {
		blockNum := big.NewInt(height)
		if height == -1 { // latest block
			blockNum = nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(w.blockTime)*2*time.Millisecond)
		block, err := client.BlockByNumber(ctx, blockNum)
		cancel()

		if err == nil {
			return block, nil
		}
	}

	return nil, ethereum.NotFound
}

func (w *Watcher) getTxReceipt(hash common.Hash) (*etypes.Receipt, error) {
	for _, client := range w.clients {
		receipt, err := client.TransactionReceipt(context.Background(), hash)
		if err == nil && receipt != nil {
			return receipt, nil
		}
	}

	return nil, fmt.Errorf("Failed to get tx receipt")
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

func (w *Watcher) processBlock(block *etypes.Block) (*types.Txs, error) {
	arr := make([]*types.Tx, 0)

	log.Info(w.cfg.Chain, " Block length = ", len(block.Transactions()))

	for _, tx := range block.Transactions() {
		bz, err := tx.MarshalBinary()
		if err != nil {
			log.Error("Cannot serialize ETH tx, err = ", err)
			continue
		}

		fmt.Println("Txhash = ", tx.Hash().String())

		if _, ok := w.txTrackCache.Get(tx.Hash().String()); ok {
			// This is a transaction that we are tracking. Inform Sisu about this.
			w.txTrackCh <- &chainstypes.TrackUpdate{
				Chain:       w.cfg.Chain,
				Bytes:       bz,
				Hash:        tx.Hash().String(),
				BlockHeight: int64(block.NumberU64()),
				Result:      chainstypes.TrackResultConfirmed,
			}

			continue
		}

		// Only filter our interested transaction.
		if !w.acceptTx(tx) {
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
		})
	}

	return &types.Txs{
		Chain:     w.cfg.Chain,
		Block:     block.Number().Int64(),
		BlockHash: block.Hash().String(),
		Arr:       arr,
	}, nil
}

func (w *Watcher) acceptTx(tx *etypes.Transaction) bool {
	if tx.To() != nil {
		to := strings.ToLower(tx.To().String())
		if to == w.gateway || to == w.chainAccount {
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
