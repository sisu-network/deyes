package config_test

import (
	"encoding/json"
	"testing"

	"github.com/sisu-network/deyes/config"

	"github.com/stretchr/testify/require"
)

func TestConfigJsonUnmarshall(t *testing.T) {
	s := "[{\"chain\":\"ganache1\",\"block_time\":3000,\"starting_block\":0,\"adjust_time\":0,\"rpc_url\":\"http://localhost:7545\"},{\"chain\":\"ganache2\",\"block_time\":3000,\"starting_block\":0,\"adjust_time\":0,\"rpc_url\":\"http://localhost:8545\"},{\"chain\":\"polygon-testnet\",\"block_time\":10000,\"starting_block\":0,\"adjust_time\":1000,\"rpc_url\":\"https://rpc-mumbai.maticvigil.com\"}]"
	chains := make([]config.Chain, 0)
	err := json.Unmarshal([]byte(s), &chains)
	if err != nil {
		panic(err)
	}

	require.Equal(t, 3, len(chains))
	require.Equal(t, "ganache1", chains[0].Chain)
}
