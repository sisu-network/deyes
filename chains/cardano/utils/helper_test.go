package utils

import (
	"testing"

	"github.com/sisu-network/deyes/types"
	"github.com/stretchr/testify/require"
)

func TestHelper_MapToJSONStruct(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		m := map[string]interface{}{
			"chain":         "ethereum",
			"recipient":     "0x123",
			"token_address": "0xtokenaddress",
		}

		txAddInfo := &types.TxAdditionInfo{}
		require.NoError(t, MapToJSONStruct(m, txAddInfo))
		require.Equal(t, "0x123", txAddInfo.Recipient)
		require.Equal(t, "ethereum", txAddInfo.Chain)
		require.Equal(t, "0xtokenaddress", txAddInfo.TokenAddress)
	})
}
