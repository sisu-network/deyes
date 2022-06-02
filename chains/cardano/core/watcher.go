package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/blockfrost/blockfrost-go"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

type Watcher struct {
	cfg    config.Chain
	db     database.Database
	txsCh  chan *types.Txs
	client blockfrost.APIClient

	interestedAddrs *sync.Map
}

func NewWatcher(cfg config.Chain, db database.Database, txsCh chan *types.Txs) *Watcher {
	return &Watcher{
		cfg:   cfg,
		db:    db,
		txsCh: txsCh,
	}
}

func (w *Watcher) init() {
	// We use blockfrost for now
	w.client = blockfrost.NewAPIClient(
		blockfrost.APIClientOptions{
			ProjectID: w.cfg.RpcSecret, // Exclude to load from env:BLOCKFROST_PROJECT_ID
			Server:    "https://cardano-testnet.blockfrost.io/api/v0",
		},
	)

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
	for _, addr := range addrs {
		w.interestedAddrs.Store(addr, true)
	}
}

func (w *Watcher) Start() {
	w.init()

	go w.scanChain()
}

func (w *Watcher) scanChain() {
}

func (w *Watcher) AddWatchAddr(addr string) {
}
