package types

import (
	"testing"

	"github.com/near/borsh-go"
	"github.com/stretchr/testify/require"
)

func TestAsset_Serialize(t *testing.T) {
	moduleId := uint32(100)
	assetId := uint32(100)
	fee := uint64(100)
	nonce := uint64(100)

	txMsg := TransactionMessage{
		ModuleID:        &moduleId,
		AssetID:         &assetId,
		Fee:             &fee,
		Asset:           []byte{19},
		Nonce:           &nonce,
		SenderPublicKey: []byte{100}, // TODO: check if this is correct
	}

	bz, err := borsh.Serialize(txMsg)
	require.Nil(t, err)

	tx := &TransactionMessage{}
	err = borsh.Deserialize(tx, bz)
	require.Nil(t, err)

	require.Equal(t, &txMsg, tx)
}
