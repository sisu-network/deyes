package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelper_MapToJSONStruct(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		type TxAdditionInfo struct {
			DestinationChain        string `json:"destination_chain,omitempty"`
			DestinationRecipient    string `json:"destination_recipient,omitempty"`
			DestinationTokenAddress string `json:"destination_token_address,omitempty"`
		}

		m := map[string]interface{}{
			"destination_chain":         "ethereum",
			"destination_recipient":     "0x123",
			"destination_token_address": "0xtokenaddress",
		}

		txAddInfo := &TxAdditionInfo{}
		require.NoError(t, MapToJSONStruct(m, txAddInfo))
		require.Equal(t, "0x123", txAddInfo.DestinationRecipient)
		require.Equal(t, "ethereum", txAddInfo.DestinationChain)
		require.Equal(t, "0xtokenaddress", txAddInfo.DestinationTokenAddress)
	})
}
