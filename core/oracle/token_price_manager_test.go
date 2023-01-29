package oracle

import (
	"fmt"
	"math/big"
	"net/http"
	"testing"

	"github.com/sisu-network/deyes/core/oracle/sushiswap"
	"github.com/sisu-network/deyes/core/oracle/uniswap"
	"github.com/sisu-network/deyes/types"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/network"
	"github.com/stretchr/testify/require"
)

func TestTokenManager(t *testing.T) {
	t.Run("Get price success", func(t *testing.T) {
		cfg := config.Deyes{
			PricePollFrequency: 1,
			PriceOracleUrl:     "http://example.com",
			PriceTokenList:     []string{"ETH", "BTC"},
			DaiTokenAddress:    "0x6B175474E89094C44Da98b954EedeAC495271d0F",
			DbHost:             "127.0.0.1",
			DbSchema:           "deyes",
			InMemory:           true,
			EthRpc:             "http://example.com",
			Tokens: map[string]config.Token{
				"btc": {Token: "BTC", Address: "0xB83c27805aAcA5C7082eB45C868d955Cf04C337F"},
				"eth": {Token: "ETH", Address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"},
			},
		}

		net := &network.MockHttp{
			GetFunc: func(req *http.Request) ([]byte, error) {
				return []byte(`{"data":{"BTC":{"quote":{"USD":{"price":36367.076791144566}}},"ETH":{"quote":{"USD":{"price":2410.875945408672}}}}}`), nil
			},
		}

		sushiswap := &sushiswap.MockSushiSwapManager{
			GetPriceFromSushiswapFunc: func(tokenAddress string, tokenName string) (*types.TokenPrice, error) {
				price, ok := new(big.Int).SetString("10000000", 10)
				require.Equal(t, ok, true)
				return &types.TokenPrice{
					Price: price,
					Id:    tokenName,
				}, nil
			},
		}

		uniswap := &uniswap.MockNewUniwapManager{
			GetPriceFromUniswapFunc: func(tokenAddress string, tokenName string) (*types.TokenPrice, error) {
				price, ok := new(big.Int).SetString("10000000", 10)
				require.Equal(t, ok, true)
				return &types.TokenPrice{
					Price: price,
					Id:    tokenName,
				}, nil
			},
		}

		priceManager := NewTokenPriceManager(cfg, net, uniswap, sushiswap)
		price, err := priceManager.GetTokenPrice("ETH")
		require.Nil(t, err)
		require.Equal(t, "10000000", price.String())
	})

	t.Run("Get price fail", func(t *testing.T) {
		cfg := config.Deyes{
			DbHost:          "127.0.0.1",
			DbSchema:        "deyes",
			DaiTokenAddress: "0x6B175474E89094C44Da98b954EedeAC495271d0F",
			InMemory:        true,
			EthRpc:          "http://example.com",
			Tokens: map[string]config.Token{
				"btc": {Token: "BTC", Address: "0xB83c27805aAcA5C7082eB45C868d955Cf04C337F"},
				"eth": {Token: "ETH", Address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"},
			},
		}

		net := &network.MockHttp{
			GetFunc: func(req *http.Request) ([]byte, error) {
				return nil, fmt.Errorf("Not found")
			},
		}
		sushiswap := &sushiswap.MockSushiSwapManager{
			GetPriceFromSushiswapFunc: func(tokenAddress string, tokenName string) (*types.TokenPrice, error) {
				return nil, fmt.Errorf("Not found")
			},
		}

		uniswap := &uniswap.MockNewUniwapManager{
			GetPriceFromUniswapFunc: func(tokenAddress string, tokenName string) (*types.TokenPrice, error) {
				return nil, fmt.Errorf("Not found")
			},
		}

		priceManager := NewTokenPriceManager(cfg, net, uniswap, sushiswap)
		_, err := priceManager.GetTokenPrice("ETH")
		require.NotNil(t, err)
	})

	t.Run("token price cache", func(t *testing.T) {
		cfg := config.Deyes{
			DbHost:          "127.0.0.1",
			DbSchema:        "deyes",
			DaiTokenAddress: "0x6B175474E89094C44Da98b954EedeAC495271d0F",
			InMemory:        true,
			EthRpc:          "http://example.com",
			Tokens: map[string]config.Token{
				"btc": {Token: "BTC", Address: "0xB83c27805aAcA5C7082eB45C868d955Cf04C337F"},
				"eth": {Token: "ETH", Address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"},
			},
		}

		count := 0
		net := &network.MockHttp{
			GetFunc: func(req *http.Request) ([]byte, error) {
				if count == 0 {
					count++
					return []byte(`{"data":{"ETH":{"quote":{"USD":{"price":10000000000}}}}}`), nil
				}

				return []byte(`{"data":{"ETH":{"quote":{"USD":{"price":20000000000}}}}}`), nil
			},
		}

		sushiswap := &sushiswap.MockSushiSwapManager{
			GetPriceFromSushiswapFunc: func(tokenAddress string, tokenName string) (*types.TokenPrice, error) {
				price := new(big.Int)
				if count == 0 {
					count++
					price, _ = new(big.Int).SetString("10000000000", 10)
				} else {
					price, _ = new(big.Int).SetString("20000000000", 10)
				}

				return &types.TokenPrice{
					Price: price,
					Id:    tokenName,
				}, nil
			},
		}

		uniswap := &uniswap.MockNewUniwapManager{
			GetPriceFromUniswapFunc: func(tokenAddress string, tokenName string) (*types.TokenPrice, error) {
				price := new(big.Int)
				if count == 0 {
					count++
					price, _ = new(big.Int).SetString("10000000000", 10)
				} else {
					price, _ = new(big.Int).SetString("20000000000", 10)
				}
				return &types.TokenPrice{
					Price: price,
					Id:    tokenName,
				}, nil
			},
		}

		priceManager := NewTokenPriceManager(cfg, net, uniswap, sushiswap)
		price, err := priceManager.GetTokenPrice("ETH")
		require.Nil(t, err)
		require.Equal(t, "10000000000", price.String())

		// Change the updateFrequency to invalidate cache.
		priceManager.(*defaultTokenPriceManager).updateFrequency = 0

		price, err = priceManager.GetTokenPrice("ETH")
		require.Nil(t, err)
		require.Equal(t, "20000000000", price.String())
	})
}
