package chains

import (
	"fmt"
	"os"

	ethCore "github.com/sisu-network/deyes/chains/eth-family/core"
	"github.com/sisu-network/deyes/client"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

// This struct handles the logic in deyes.
// TODO: Make this processor to support multiple chains at the same time.
type TxProcessor struct {
	chain      string
	db         *database.Database
	txsCh      chan *types.Txs
	blockTime  int
	sisuClient *client.Client

	watchers map[string]Watcher
}

func NewTxProcessor(chain string, blockTime int, db *database.Database,
	sisuClient *client.Client) *TxProcessor {
	return &TxProcessor{
		chain:      chain,
		db:         db,
		blockTime:  blockTime,
		sisuClient: sisuClient,
		watchers:   make(map[string]Watcher),
	}
}

func (tp *TxProcessor) Start() {
	utils.LogInfo("Starting tx processor...")

	tp.txsCh = make(chan *types.Txs)
	go tp.listen()

	switch tp.chain {
	case "eth":
		watcher := ethCore.NewWatcher(tp.db, os.Getenv("CHAIN_RPC_URL"), tp.blockTime, tp.chain, tp.txsCh)
		watcher.Start()
		tp.watchers[tp.chain] = watcher

	default:
		panic(fmt.Errorf("Unknown chain"))
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
