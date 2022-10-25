package solana

import (
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/lib/log"
)

func TestWatcher(t *testing.T) {
	t.Skip()

	w := NewWatcher(config.Chain{Chain: "solana-devnet"}, nil, nil, nil)
	result, err := w.getBlockNumber(170_990_453)
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
			if acc.String() == "HguMTvmDfspHuEWycDSP1XtVQJi47hVNAyLbFEf2EJEQ" {
				log.Verbose("Bridge program ID found!!!")

				log.Verbose("decodedTx.Message.Instructions[0].ProgramIDIndex = ", decodedTx.Message.Instructions[0].Data)
				data := decodedTx.Message.Instructions[0].Data
				log.Verbose([]byte(data))
			}
		}
	}
}
