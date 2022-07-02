package core

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
	gateway         string

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

	var err error
	w.gateway, err = w.db.GetGateway(w.cfg.Chain)
	if err != nil {
		panic(err)
	}

	log.Infof("Saved gateway in the db for chain %s is %s", w.cfg.Chain, w.gateway)

	w.lastBlockHeight.Store(0)
}

func (w *Watcher) Start() {
	w.init()

	go w.scanChain()
}

func (w *Watcher) scanChain() {
	log.Info("Start scanning chain: ", w.cfg.Chain)

	for {
		// Get next block to scan
		block, err := w.getNextBlock()

		if err != nil {
			w.blockTime = w.blockTime + w.cfg.AdjustTime
			time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
			if !strings.Contains(err.Error(), "Not Found") {
				// This error is not a Block not found error. Print the error here.
				log.Errorf("Error when getting block for height/hash = %d, error = %v\n", w.client.LatestBlock().Height, err)
			} else {
				log.Verbose("Block not found. We need to wait more. w.blockTime = ", w.blockTime)
			}
			continue
		}

		log.Info("Block height for Cardano Scanner = ", block.Height)

		w.lastBlockHeight.Store(int32(block.Height))
		w.blockTime = w.blockTime - w.cfg.AdjustTime/4

		// Make a copy of the w.interestedAddr map
		copy := make(map[string]bool)
		w.lock.RLock()
		for addr, txCount := range w.interestedAddr {
			copy[addr] = txCount
		}
		w.lock.RUnlock()

		// Process each address in the interested addr.
		txArr := make([]*types.Tx, 0)

		txsIn, err := w.client.NewTxs(block.Height, copy)
		if err != nil {
			log.Error("Cannot get list of new transaction at block ", block.Height, " err = ", err)
			time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
			continue
		}

		log.Verbose("Filtered txs sizes = ", len(txsIn), " on chain ", w.cfg.Chain)

		for _, txIn := range txsIn {
			bz, err := json.Marshal(txIn)
			if err != nil {
				log.Error("Cannot serialize utxo, err = ", err)
				continue
			}

			txArr = append(txArr, &types.Tx{
				Hash:       txIn.Hash,
				Serialized: bz,
				To:         txIn.Address,
				Success:    true,
			})
		}

		if len(txArr) > 0 {
			txs := types.Txs{
				Chain: w.cfg.Chain,
				Block: int64(block.Height),
				Arr:   txArr,
			}

			// Broadcast the result
			w.txsCh <- &txs
		}

		// Sleep until next block
		time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
	}
}

func (w *Watcher) getNextBlock() (*blockfrost.Block, error) {
	lastScanBlock := int(w.lastBlockHeight.Load())
	nextBlock := lastScanBlock + 1
	if lastScanBlock == 0 {
		nextBlock = w.client.LatestBlock().Height
	}

	block, err := w.client.GetBlock(strconv.Itoa(nextBlock))
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (w *Watcher) SetGateway(addr string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	err := w.db.SetGateway(w.cfg.Chain, addr)
	if err == nil {
		w.gateway = addr
	} else {
		log.Error("Failed to save gateway")
	}
}
