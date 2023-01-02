package lisk

import (
	"testing"

	"github.com/sisu-network/deyes/config"
	"github.com/stretchr/testify/require"
)

func TestIntegration_CreateNewClient(t *testing.T) {
	t.Skip()
	config := config.Chain{Chain: "lisk-testnet", Rpcs: []string{"https://testnet-service.lisk.com/api/v2"}}
	client := NewLiskClient(config).(*defaultLiskClient)

	require.Equal(t, client.rpc, config.Rpcs[0])
	require.Equal(t, client.chain, config.Chain)

	blockNumber, err := client.BlockNumber()
	require.Nil(t, err)

	block, err := client.BlockByHeight(blockNumber)
	require.Nil(t, err)
	require.NotNil(t, block)
}
