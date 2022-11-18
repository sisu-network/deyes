package solana

import (
	"testing"

	solanatypes "github.com/sisu-network/deyes/chains/solana/types"

	"github.com/mr-tron/base58"
	"github.com/near/borsh-go"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

const RPC = ""

// Sanity testing
func TestWatcherBlockScanning(t *testing.T) {
	t.Skip()

	cfg := config.Chain{
		Chain:                 "solana-devnet",
		Rpcs:                  []string{RPC},
		SolanaBridgeProgramId: "GWP9AoY6ZvUqLzm4fS5jqSJAJ8rnrMf4d1kiU1wSXwED",
	}

	w := NewWatcher(cfg, nil, nil, nil)
	result, err := w.getBlockNumber(172121080)
	if err != nil {
		panic(err)
	}

	for _, outerTx := range result.Transactions {
		if w.acceptTx(outerTx) {
			log.Info("Bridge program id found!!!!!")

			for _, ix := range outerTx.TransactionInner.Message.Instructions {
				bytesArr, err := base58.Decode(ix.Data)
				if err != nil {
					panic(err)
				}

				borshBz := bytesArr[1:]
				transferData := new(solanatypes.TransferOutData)
				err = borsh.Deserialize(transferData, borshBz)
				if err != nil {
					panic(err)
				}

				log.Verbose(*transferData)
			}
		}
	}
}

func TestFullWatcher(t *testing.T) {
	t.Skip()

	txsCh := make(chan *types.Txs)
	txTrackCh := make(chan *chainstypes.TrackUpdate)

	w := NewWatcher(config.Chain{
		Chain:                 "solana-devnet",
		BlockTime:             1000,
		AdjustTime:            500,
		Rpcs:                  []string{RPC},
		SolanaBridgeProgramId: "GWP9AoY6ZvUqLzm4fS5jqSJAJ8rnrMf4d1kiU1wSXwED",
	}, nil, txsCh, txTrackCh)

	w.Start()

	select {
	case txs := <-txsCh:
		log.Verbose("There is a transaction, txs = ", txs)
	}
}
