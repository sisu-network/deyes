package core

import (
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
	client          CardanoClient
	blockTime       int
	lastBlockHeight atomic.Int32

	// A map between an interested address to the number of transaction to it (according to our data).
	interestedAddr map[string]bool
	lock           *sync.RWMutex
}

func NewWatcher(cfg config.Chain, db database.Database, txsCh chan *types.Txs, client CardanoClient) *Watcher {
	return &Watcher{
		cfg:            cfg,
		db:             db,
		txsCh:          txsCh,
		blockTime:      cfg.BlockTime,
		interestedAddr: map[string]bool{},
		lock:           &sync.RWMutex{},
		client:         client,
	}
}

func (w *Watcher) init() {
	healthy := w.client.IsHealthy()
	if !healthy {
		err := fmt.Errorf("Blockfrost is not healthy")
		log.Error(err)
		panic(err)
	}

	addrs := w.db.LoadWatchAddresses(w.cfg.Chain)
	log.Info("Watch address for chain ", w.cfg.Chain, ": ", addrs)

	w.lock.Lock()
	for _, addr := range addrs {
		w.interestedAddr[addr.Address] = true
	}
	w.lock.Unlock()

	w.lastBlockHeight.Store(0)
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

		w.lastBlockHeight.Store(int32(block.Height))
		w.blockTime = w.blockTime - w.cfg.AdjustTime/4

		// Make a copy of the w.interestedAddr map
		copy := make(map[string]bool)
		w.lock.RLock()
		for addr, txCount := range w.interestedAddr {
			copy[addr] = txCount
		}
		w.lock.RUnlock()

		// Process each address in the intersted addr.
		txArr := make([]*types.Tx, 0)

		utxos, err := w.client.NewTxs(block.Height, copy)
		if err != nil {
			log.Error("Cannot get list of new transaction at block ", block.Height, " err = ", err)
			time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
			continue
		}

		log.Verbose("Filtered txs sizes = ", len(utxos), " on chain ", w.cfg.Chain)

		for _, utxo := range utxos {
			bz, err := json.Marshal(utxo)
			if err != nil {
				log.Error("Cannot serialize utxo, err = ", err)
				continue
			}

			txArr = append(txArr, &types.Tx{
				Hash:       utxo.Hash(),
				Serialized: bz,
				To:         utxo.Spender.String(),
			})
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

func (w *Watcher) getLatestBlock() (*blockfrost.Block, error) {
	block := w.client.LatestBlock()

	if block == nil {
		err := fmt.Errorf("Cannot get latest block")
		return nil, err
	}

	if block.Height == int(w.lastBlockHeight.Load()) {
		return nil, BlockNotFound
	}

	return block, nil
}

func (w *Watcher) AddWatchAddr(addr string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.interestedAddr[addr] = true
}
