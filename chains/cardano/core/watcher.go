package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/blockfrost/blockfrost-go"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
	"go.uber.org/atomic"
)

var (
	BlockNotFound = fmt.Errorf("Block not found")
)

type Watcher struct {
	cfg             config.Chain
	db              database.Database
	txsCh           chan *types.Txs
	client          *blockfrostClient
	blockTime       int
	lastBlockHeight atomic.Int32

	// A map between an interested address to the number of transaction to it (according to our data).
	interestedAddr map[string]int
	lock           *sync.RWMutex
}

func NewWatcher(cfg config.Chain, db database.Database, txsCh chan *types.Txs) *Watcher {
	return &Watcher{
		cfg:            cfg,
		db:             db,
		txsCh:          txsCh,
		blockTime:      cfg.BlockTime,
		interestedAddr: map[string]int{},
		lock:           &sync.RWMutex{},
	}
}

func (w *Watcher) init() {
	// We use blockfrost for now
	w.client = newAPIClient(blockfrost.APIClientOptions{
		ProjectID: w.cfg.RpcSecret, // Exclude to load from env:BLOCKFROST_PROJECT_ID
		Server:    w.cfg.Rpcs[0],
	})

	status, err := w.client.Health(context.Background())
	if err != nil {
		panic(err)
	}

	if !status.IsHealthy {
		err := fmt.Errorf("Blockfrost is not healthy")
		log.Error(err)
		panic(err)
	}

	addrs := w.db.LoadWatchAddresses(w.cfg.Chain)
	log.Info("Watch address for chain ", w.cfg.Chain, ": ", addrs)

	w.lock.Lock()
	for _, addr := range addrs {
		w.interestedAddr[addr.Address] = addr.TxCount
	}
	w.lock.Unlock()
}

func (w *Watcher) Start() {
	w.init()

	go w.scanChain()
}

func (w *Watcher) scanChain() {
	log.Info("Start scanning chain: ", w.cfg.Chain)

	for {
		// Get latest block
		block, err := w.getLatestBlock()
		if err != nil && err != BlockNotFound {
			time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
			continue
		}

		// Block not available yet
		if err == BlockNotFound {
			w.blockTime = w.blockTime + w.cfg.AdjustTime
			time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
			continue
		}

		log.Verbose("Block height = ", block.Height)
		w.lastBlockHeight.Store(int32(block.Height))
		w.blockTime = w.blockTime - w.cfg.AdjustTime/4

		// Make a copy of the w.interestedAddr map
		copy := make(map[string]int)
		w.lock.RLock()
		for addr, txCount := range w.interestedAddr {
			copy[addr] = txCount
		}
		w.lock.RUnlock()

		// Process each address in the intersted addr.
		txArr := make([]*types.Tx, 0)
		for addr, txCount := range copy {
			txs, err := w.processAddr(addr, txCount)
			if err != nil {
				log.Error(err)
				continue
			}

			for _, txutxo := range txs {
				serialized, err := json.Marshal(txutxo)
				if err != nil {
					log.Critical("cannot marshal txutxo, err = ", err)
					continue
				}

				txArr = append(txArr, &types.Tx{
					Hash:       txutxo.Hash,
					Serialized: serialized,
				})
			}
		}

		txs := types.Txs{
			Chain: w.cfg.Chain,
			Block: int64(block.Height),
			Arr:   txArr,
		}

		// Broadcast the result
		w.txsCh <- &txs

		// Sleep until next block
		time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
	}
}

func (w *Watcher) processAddr(addr string, lastTxCount int) ([]*blockfrost.TransactionUTXOs, error) {
	addrDetails, err := w.client.AddressDetails(context.Background(), addr)
	if err != nil {
		return nil, err
	}

	txCount := addrDetails.TxCount
	if txCount == lastTxCount {
		// No new transaction
		return make([]*blockfrost.TransactionUTXOs, 0), nil
	}

	if txCount < lastTxCount {
		err := fmt.Errorf("Tx count is smaller than lastTxCount")
		return nil, err
	}

	addrTxs, err := w.client.AddressTransactions(context.Background(), addr, blockfrost.APIQueryParams{
		Count: txCount - lastTxCount,
		Order: "desc",
	})

	txs := make([]*blockfrost.TransactionUTXOs, 0)

	for _, addrTx := range addrTxs {
		txUtxos, err := w.client.TransactionUTXOs(context.Background(), addrTx.TxHash)
		if err != nil {
			return nil, err
		}

		txs = append(txs, &txUtxos)
	}

	return txs, nil
}

func (w *Watcher) getLatestBlock() (*blockfrost.Block, error) {
	block, err := w.client.BlockLatest(context.Background())
	if err != nil {
		return nil, err
	}

	if block.Height == int(w.lastBlockHeight.Load()) {
		return nil, BlockNotFound
	}

	return &block, nil
}

func (w *Watcher) AddWatchAddr(addr string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.interestedAddr[addr] = 0
}
