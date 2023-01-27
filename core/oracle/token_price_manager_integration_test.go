package oracle

import (
	"github.com/sisu-network/deyes/core/oracle/sushiswap"
	"github.com/sisu-network/deyes/core/oracle/uniswap"
	"github.com/sisu-network/deyes/types"
	"math/big"
	"testing"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/lib/log"
	"github.com/stretchr/testify/require"
)

func TestGettingTokenPrice(t *testing.T) {
	t.Skip()
	cfg := config.Deyes{
		PricePollFrequency: 1000,
		PriceOracleUrl:     "",
		PriceOracleSecret:  "",
		PriceTokenList:     []string{"ETH", "BTC", "AVAX"},

		DbHost:   "127.0.0.1",
		DbSchema: "deyes",
		InMemory: true,
	}

	dbInstance := database.NewDb(&cfg)
	err := dbInstance.Init()
	if err != nil {
		panic(err)
	}

	sushiswap := &sushiswap.MockSushiSwapManager{
		GetPriceFromSushiswapFunc: func(tokenAddress string) (*types.TokenPrice, error) {
			price, err := new(big.Int).SetString("10", 10)
			require.Nil(t, err)
			return &types.TokenPrice{
				Price: price,
				Id:    "",
			}, nil
		},
	}

	uniswap := &uniswap.MockNewUniwapManager{
		GetPriceFromUniswapFunc: func(tokenAddress string) (*types.TokenPrice, error) {
			price, err := new(big.Int).SetString("10", 10)
			require.Nil(t, err)
			return &types.TokenPrice{
				Price: price,
				Id:    "",
			}, nil
		},
	}

	tpm := NewTokenPriceManager(cfg, dbInstance, network.NewHttp(), uniswap, sushiswap)
	go tpm.Start()
	defer tpm.Stop()

	price, err := tpm.GetTokenPrice("ETH")
	require.Nil(t, err)
	log.Verbosef("price = ", price)
}
