package core

import (
	"encoding/json"
	"testing"

	"github.com/echovl/cardano-go"
	"github.com/sisu-network/deyes/types"
	"github.com/stretchr/testify/require"
)

func TestCardanoDispatcher_DeserializeTx(t *testing.T) {
	// Construct tx
	tx := &cardano.Tx{
		Body: cardano.TxBody{},
	}
	txHash, err := cardano.NewHash32("bc82779c18b98f0f5628b0cae12af618020e5388258d3bcce936c380583298dc")
	require.Nil(t, err)
	receiver, err := cardano.NewAddress("addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy")
	require.Nil(t, err)
	txIn := cardano.NewTxInput(txHash, 0, cardano.NewValue(cardano.Coin(100)))
	txOut := cardano.NewTxOutput(receiver, cardano.NewValue(1000000))
	tx.Body.Inputs = append(tx.Body.Inputs, txIn)
	tx.Body.Outputs = append(tx.Body.Outputs, txOut)

	bz, err := json.Marshal(tx)
	require.Nil(t, err)

	// Test serializing & deserializing tx.
	client := &MockCardanoClient{}
	client.SubmitTxFunc = func(tx *cardano.Tx) (*cardano.Hash32, error) {
		return &txHash, nil
	}
	dispatcher := NewDispatcher(client)

	result := dispatcher.Dispatch(&types.DispatchedTxRequest{
		Tx: bz,
	})

	require.True(t, result.Success)
	require.Equal(t, result.TxHash, txHash.String())
}
