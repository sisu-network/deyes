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
		EthRpcs:        []string{"https://rpc.ankr.com/eth"},
		EthTokens: map[string]config.TokenPair{
			"BTC": {
				Token1:   "BTC",
				Token2:   "DAI",
				Address1: "0xB83c27805aAcA5C7082eB45C868d955Cf04C337F",
				Address2: "0x6B175474E89094C44Da98b954EedeAC495271d0F",
				Decimal1: 18,
				Decimal2: 18,
			},
			"ETH": {
				Token1:   "ETH",
				Token2:   "DAI",
				Address1: "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
				Address2: "0x6B175474E89094C44Da98b954EedeAC495271d0F",
				Decimal1: 18,
				Decimal2: 18,
			},
			"MATIC": {
				Token1:   "MATIC",
				Token2:   "DAI",
				Address1: "0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0",
				Address2: "0x6B175474E89094C44Da98b954EedeAC495271d0F",
				Decimal1: 18,
				Decimal2: 18,
			},
		},
	}

	networkHttp := network.NewHttp()
	uniswapManager := uniswap.NewUniwapManager(cfg)
	sushiSwapManager := sushiswap.NewSushiSwapManager(cfg)
	m := NewTokenPriceManager(cfg, networkHttp, uniswapManager, sushiSwapManager)

	price, err := m.GetTokenPrice("MATIC")
	require.Nil(t, err)

	log.Info("price = ", price)
}
