package core

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
	libchain "github.com/sisu-network/lib/chain"
	"github.com/sisu-network/lib/log"
)

const (
	minGasPrice = 10_000_000_000
)

type GasPriceGetter func(ctx context.Context) (*big.Int, error)

// TODO: Move this to the chains package.
type Watcher struct {
	cfg         config.Chain
	client      *ethclient.Client
	blockHeight int64
	blockTime   int
	db          database.Database
	txsCh       chan *types.Txs
	gasPriceCh  chan *types.GasPriceRequest
	// A set of address we are interested in. Only send information about transaction to these
	// addresses back to Sisu.
	interestedAddrs *sync.Map

	signers map[string]etypes.Signer

	gasPrice        *big.Int
	gasPriceLocker  sync.RWMutex
	gasPriceGetters []GasPriceGetter
}

func NewWatcher(db database.Database, cfg config.Chain, txsCh chan *types.Txs, gasPriceCh chan *types.GasPriceRequest) *Watcher {
	w := &Watcher{
		db:              db,
		cfg:             cfg,
		txsCh:           txsCh,
		gasPriceCh:      gasPriceCh,
		interestedAddrs: &sync.Map{},
	}

	gasPriceGetters := []GasPriceGetter{w.getGasPriceFromNode}
	w.gasPriceGetters = gasPriceGetters
	return w
}

func (w *Watcher) init() {
	var err error

	log.Info("RPC endpoint for chain", w.cfg.Chain, "is", w.cfg.RpcUrl)
	w.client, err = ethclient.Dial(w.cfg.RpcUrl)
	if err != nil {
		panic(err)
	}

	// Set the block height for the watcher.
	w.setBlockHeight()

	// Load watch addresses
	addrs := w.db.LoadWatchAddresses(w.cfg.Chain)

	log.Info("Watch address for chain ", w.cfg.Chain, ": ", addrs)

	for _, addr := range addrs {
		w.interestedAddrs.Store(addr, true)
	}
}

func (w *Watcher) setBlockHeight() {
	for {
		number, err := w.client.BlockNumber(context.Background())
		if err != nil {
			log.Error("cannot get latest block number. Sleeping for a few seconds")
			time.Sleep(time.Second * 5)
			continue
		}

		w.blockHeight = int64(number)
		break
	}

	log.Info("Watching from block", w.blockHeight, " for chain ", w.cfg.Chain)
}

func (w *Watcher) AddWatchAddr(addr string) {
	w.interestedAddrs.Store(addr, true)
	w.db.SaveWatchAddress(w.cfg.Chain, addr)
}

func (w *Watcher) Start() {
	log.Info("Starting Watcher...")

	w.init()
	go w.scanBlocks()
}

func (w *Watcher) scanBlocks() {
	latestBlock, err := w.getLatestBlock()
	if err == nil {
		log.Error(err)
	}

	if latestBlock != nil {
		w.blockHeight = latestBlock.Header().Number.Int64()
	}
	log.Info(w.cfg.Chain, "Latest height = ", w.blockHeight)

	for {
		// Only update gas price at deterministic block height
		// Ex: updateBlockHeight = startBlockHeight + (n * interval) (n is an integer from 0 ... )
		chainParams := config.ChainParamsMap[w.cfg.Chain]
		if (w.blockHeight-chainParams.GasPriceStartBlockHeight)%chainParams.Interval == 0 {
			w.updateGasPrice(context.Background())
			w.gasPriceCh <- &types.GasPriceRequest{
				Chain:    w.cfg.Chain,
				Height:   w.blockHeight,
				GasPrice: w.GetGasPrice(),
			}
		}

		// Get the blockheight
		block, err := w.tryGetBlock()
		if err != nil || block == nil {
			log.Error("Cannot get block at height", w.blockHeight, "for chain", w.cfg.Chain)
			time.Sleep(time.Duration(w.cfg.BlockTime) * time.Millisecond)
			continue
		}

		filteredTxs, err := w.processBlock(block)
		if err != nil {
			log.Error("cannot process block, err = ", err)
			continue
		}

		log.Verbose("Filtered txs sizes = ", len(filteredTxs.Arr))

		if len(filteredTxs.Arr) > 0 {
			// Send list of interested txs back to the listener.
			w.txsCh <- filteredTxs
		}

		// Save all txs into database for later references.
		w.db.SaveTxs(w.cfg.Chain, w.blockHeight, filteredTxs)

		w.blockHeight++

		time.Sleep(time.Duration(w.cfg.BlockTime) * time.Millisecond)
	}
}

// Get block with retry when block is not mined yet.
func (w *Watcher) tryGetBlock() (*etypes.Block, error) {
	block, err := w.getBlock(w.blockHeight)
	switch err {
	case nil:
		log.Debug(w.cfg.Chain, "Height = ", block.Number())
		return block, nil

	case ethereum.NotFound:
		// TODO: Ping block for every second.
		for i := 0; i < 10; i++ {
			block, err = w.getBlock(w.blockHeight)
			if err == nil {
				return block, err
			}

			time.Sleep(time.Duration(w.cfg.BlockTime) / 2 * time.Millisecond)
		}
	}

	return block, err
}

func (w *Watcher) getLatestBlock() (*etypes.Block, error) {
	return w.client.BlockByNumber(context.Background(), nil)
}

func (w *Watcher) getBlock(height int64) (*etypes.Block, error) {
	return w.client.BlockByNumber(context.Background(), big.NewInt(height))
}

func (w *Watcher) processBlock(block *etypes.Block) (*types.Txs, error) {
	arr := make([]*types.Tx, 0)

	log.Info(w.cfg.Chain, "Block length = ", len(block.Transactions()))

	for _, tx := range block.Transactions() {
		bz, err := tx.MarshalBinary()
		if err != nil {
			log.Error("Cannot serialize ETH tx, err = ", err)
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
		arr = append(arr, &types.Tx{
			Hash:       tx.Hash().String(),
			Serialized: bz,
			To:         to,
			From:       from.Hex(),
		})
	}

	return &types.Txs{
		Chain: w.cfg.Chain,
		Arr:   arr,
	}, nil
}

func (w *Watcher) acceptTx(tx *etypes.Transaction) bool {
	if tx.To() != nil {
		_, ok := w.interestedAddrs.Load(tx.To().Hex())
		if ok {
			return true
		}
	}

	// check from
	from, err := w.getFromAddress(w.cfg.Chain, tx)
	log.Verbose("from = ", from.Hex(), " to ", tx.To(), " hash = ", tx.Hash())
	if err == nil {
		_, ok := w.interestedAddrs.Load(from.Hex())
		if ok {
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

	msg, err := tx.AsMessage(etypes.NewEIP2930Signer(tx.ChainId()), nil)
	if err != nil {
		return common.Address{}, err
	}

	return msg.From(), nil
}

func (w *Watcher) GetNonce(address string) int64 {
	nonce, err := w.client.PendingNonceAt(context.Background(), common.HexToAddress(address))
	if err != nil {
		log.Error("cannot get nonce of chain", w.cfg.Chain, " at", address)
		return -1
	}

	return int64(nonce)
}

func (w *Watcher) GetGasPrice() int64 {
	w.gasPriceLocker.RLock()
	defer w.gasPriceLocker.RUnlock()
	return w.gasPrice.Int64()
}

func (w *Watcher) updateGasPrice(ctx context.Context) error {
	potentialGasPriceList := make([]*big.Int, 0)
	for _, getter := range w.gasPriceGetters {
		gasPrice, err := getter(ctx)
		if err != nil {
			return err
		}

		// make sure the gas price is at least 10 Gwei
		if gasPrice.Cmp(big.NewInt(minGasPrice)) < 0 {
			gasPrice = big.NewInt(minGasPrice)
		}

		potentialGasPriceList = append(potentialGasPriceList, gasPrice)
	}

	medianGasPrice := utils.GetMedianBigInt(potentialGasPriceList)

	w.gasPriceLocker.Lock()
	defer w.gasPriceLocker.Unlock()
	w.gasPrice = medianGasPrice

	return nil
}

func (w *Watcher) getGasPriceFromNode(ctx context.Context) (*big.Int, error) {
	gasPrice, err := w.client.SuggestGasPrice(ctx)
	if err != nil {
		log.Error("error when getting gas price", err)
		return big.NewInt(0), err
	}

	return gasPrice, nil
}
