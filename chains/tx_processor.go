package chains

import (
	"fmt"

	ethCore "github.com/sisu-network/deyes/chains/eth-family/core"
	"github.com/sisu-network/deyes/client"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

// This struct handles the logic in deyes.
// TODO: Make this processor to support multiple chains at the same time.
type TxProcessor struct {
	db         *database.Database
	txsCh      chan *types.Txs
	sisuClient client.Client

	watchers    map[string]Watcher
	dispatchers map[string]Dispatcher
	cfg         *config.Deyes
}

func NewTxProcessor(cfg *config.Deyes, db *database.Database, sisuClient client.Client) *TxProcessor {
	return &TxProcessor{
		cfg:         cfg,
		db:          db,
		watchers:    make(map[string]Watcher),
		dispatchers: make(map[string]Dispatcher),
		sisuClient:  sisuClient,
	}
}

func (tp *TxProcessor) Start() {
	utils.LogInfo("Starting tx processor...")

	for chain, cfg := range tp.cfg.Chains {
		tp.txsCh = make(chan *types.Txs)
		go tp.listen()

		switch chain {
		case "eth":
			watcher := ethCore.NewWatcher(tp.db, &cfg, tp.txsCh)
			watcher.Start()
			tp.watchers[chain] = watcher

			// Dispatcher
			dispatcher := NewDispatcher(chain, cfg.RpcUrl)
			dispatcher.Start()
			tp.dispatchers[chain] = dispatcher

		default:
			panic(fmt.Errorf("Unknown chain"))
		}
	}
}

func (tp *TxProcessor) listen() {
	for {
		select {
		case txs := <-tp.txsCh:
			// Broadcast this to Sisu.
			tp.sisuClient.BroadcastTxs(txs)
		}
	}
}

func (tp *TxProcessor) AddWatchAddresses(chain string, addrs []string) {
	watcher := tp.watchers[chain]
	if watcher != nil {
		for _, addr := range addrs {
			utils.LogInfo("Adding watched addr", addr, "for chain", chain)
			watcher.AddWatchAddr(addr)
		}
	}
}

func (tp *TxProcessor) DispatchTx(chain string, tx []byte) error {
	dispatcher := tp.dispatchers[chain]
	if dispatcher == nil {
		return fmt.Errorf("unknown chain %s", chain)
	}

	return dispatcher.Dispatch(tx)
}
