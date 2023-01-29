package oracle

import (
	"testing"

	"github.com/sisu-network/lib/log"

	"github.com/sisu-network/deyes/config"
	"github.com/stretchr/testify/require"

	"github.com/sisu-network/deyes/core/oracle/sushiswap"
	"github.com/sisu-network/deyes/core/oracle/uniswap"

	"github.com/sisu-network/deyes/network"
)

func TestIntegrationTokenManager(t *testing.T) {
	t.Skip()

	cfg := config.Deyes{
		PriceOracleUrl: "http://example.com",
		DbHost:         "127.0.0.1",
		DbSchema:       "deyes",
		InMemory:       true,
		EthRpc:         "https://rpc.ankr.com/eth",
		EthTokens: map[string]config.TokenPair{
			"btc": {Token1: "BTC", Token2: "DAI", Address1: "0xB83c27805aAcA5C7082eB45C868d955Cf04C337F",
				Address2: "0x6B175474E89094C44Da98b954EedeAC495271d0F"},
			"eth": {Token1: "ETH", Token2: "DAI", Address1: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
				Address2: "0x6B175474E89094C44Da98b954EedeAC495271d0F"},
		},
	}

	networkHttp := network.NewHttp()
	uniswapManager := uniswap.NewUniwapManager(cfg)
	sushiSwapManager := sushiswap.NewSushiSwapManager(cfg)
	m := NewTokenPriceManager(cfg, networkHttp, uniswapManager, sushiSwapManager)

	price, err := m.GetTokenPrice("ETH")
	require.Nil(t, err)

	log.Info("price = ", price)
}
