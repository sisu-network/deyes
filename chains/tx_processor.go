package chains

import (
	"fmt"
	"os"

	ethCore "github.com/sisu-network/deyes/chains/eth-family/core"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

type TxProcessor struct {
	chain     string
	db        *database.Database
	txsCh     chan *types.Txs
	blockTime int
}

func NewTxProcessor(chain string, blockTime int, db *database.Database) *TxProcessor {
	return &TxProcessor{
		chain:     chain,
		db:        db,
		blockTime: blockTime,
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

	default:
		panic(fmt.Errorf("Unknown chain"))
	}
}

func (tp *TxProcessor) listen() {
	for {
		select {
		case txs := <-tp.txsCh:
			// Broadcast this to Sisu.
			fmt.Println(len(txs.Arr))
		}
	}
}
