package solana

import (
	"os"
	"os/signal"
	"syscall"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

// Sanity testing
func TestWatcherBlockScanning(t *testing.T) {
	t.Skip()

	w := NewWatcher(config.Chain{Chain: "solana-devnet"}, nil, nil, nil)
	result, err := w.getBlockNumber(171905242)
	if err != nil {
		panic(err)
	}

	for _, outerTx := range result.Transactions {
		tx, err := outerTx.GetTransaction()
		if err != nil {
			panic(err)
		}

		decodedTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(outerTx.Transaction.GetBinary()))
		for _, acc := range tx.Message.AccountKeys {
			if acc.String() == "GWP9AoY6ZvUqLzm4fS5jqSJAJ8rnrMf4d1kiU1wSXwED" {
				log.Verbose("Bridge program ID found!!!")

				log.Verbose("decodedTx.Message.Instructions[0].ProgramIDIndex = ", decodedTx.Message.Instructions[0].Data)
				data := decodedTx.Message.Instructions[0].Data
				log.Verbose([]byte(data))
			}
		}
	}
}

func TestFullWatcher(t *testing.T) {
	// t.Skip()

	txsCh := make(chan *types.Txs)
	txTrackCh := make(chan *chainstypes.TrackUpdate)

	w := NewWatcher(config.Chain{
		Chain:      "solana-devnet",
		BlockTime:  1000,
		AdjustTime: 500,
		Rpcs:       []string{"SOLANA_URL"},
	}, nil, txsCh, txTrackCh)

	w.SetVault("ckv5WFUVu8wjsgNRydx4ib3cDq2jHQ1RkiG4WL6wbJi", "ADA")

	w.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	<-c
}
