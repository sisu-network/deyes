package lisk_test

import (
	"github.com/sisu-network/deyes/chains/lisk"
	"github.com/sisu-network/deyes/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLiskDispatcher_DeserializeTx(t *testing.T) {
	txHash := "08021000184c2080c2d72f2a20f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289321e088094ebdc031214445f5ae1342837a1231f9d36d34a79145c1cd0141a003a402023365b254049ee9a28d244d6f65b690564ad325a3d5bb79f9d31805b4b11f9cb05b8f8255a89572a7043a3c8966b4237086f891feccdb7cd25a41a292c4802"
	tx := types.DispatchedTxRequest{Chain: "lisk-testnet", TxHash: txHash}
	transactionResult := "08021000184c2080c2d72f2a20f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289321e088094ebdc031214445f5ae1342837"
	client := &lisk.MockLiskClient{
		CreateTransactionFunc: func(txHash string) (string, error) {
			return transactionResult, nil
		},
	}
	dispatcher := lisk.NewDispatcher("lisk-testnet", client)
	dpResult := dispatcher.Dispatch(&tx)
	require.Equal(t, dpResult.TxHash, transactionResult)

}
