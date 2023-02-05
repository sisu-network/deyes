package oracle

import (
	"fmt"
	"math/big"
	"net/http"
	"testing"

	"github.com/sisu-network/deyes/core/oracle/sushiswap"
	"github.com/sisu-network/deyes/core/oracle/uniswap"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/network"
	"github.com/stretchr/testify/require"
)

func TestTokenManager(t *testing.T) {
	t.Run("get_price_success", func(t *testing.T) {
		cfg := config.Deyes{
			PriceOracleUrl: "http://example.com",
			DbHost:         "127.0.0.1",
			DbSchema:       "deyes",
			InMemory:       true,
			EthRpcs:        []string{"http://example.com"},
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
			},
		}

		net := &network.MockHttp{
			GetFunc: func(req *http.Request) ([]byte, error) {
				return []byte(
					`{"data":{"BTC":{"quote":{"USD":{"price":36367.076791144566}}},"ETH":{"quote":{"USD":{"price":2410.875945408672}}}}}`), nil
			},
		}

		sushiswap := &sushiswap.MockSushiSwapManager{
			GetPriceFromSushiswapFunc: func(tokenAddress1 string, tokenAddress2 string, amount *big.Int) (
				*big.Int, error) {
				price, ok := new(big.Int).SetString("10000000", 10)
				require.Equal(t, ok, true)

				return price, nil
			},
		}

		uniswap := &uniswap.MockNewUniwapManager{
			GetPriceFromUniswapFunc: func(tokenAddress1 string, tokenAddress2 string) (*big.Int, error) {
				price, ok := new(big.Int).SetString("10000000", 10)
				require.Equal(t, ok, true)

				return price, nil
			},
		}

		priceManager := NewTokenPriceManager(cfg, net, uniswap, sushiswap)
		price, err := priceManager.GetTokenPrice("ETH")
		require.Nil(t, err)
		require.Equal(t, "10000000", price.String())
	})

	t.Run("get_price_failure", func(t *testing.T) {
		cfg := config.Deyes{
			DbHost:   "127.0.0.1",
			DbSchema: "deyes",
			InMemory: true,
			EthRpcs:  []string{"http://example.com"},
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
			},
		}

		net := &network.MockHttp{
			GetFunc: func(req *http.Request) ([]byte, error) {
				return nil, fmt.Errorf("Not found")
			},
		}
		sushiswap := &sushiswap.MockSushiSwapManager{
			GetPriceFromSushiswapFunc: func(tokenAddress1 string, tokenAddress2 string, amount *big.Int) (
				*big.Int, error) {
				return nil, fmt.Errorf("Not found")
			},
		}

		uniswap := &uniswap.MockNewUniwapManager{
			GetPriceFromUniswapFunc: func(tokenAddress1 string, tokenAddress2 string) (*big.Int, error) {
				return nil, fmt.Errorf("Not found")
			},
		}

		priceManager := NewTokenPriceManager(cfg, net, uniswap, sushiswap)
		_, err := priceManager.GetTokenPrice("ETH")
		require.NotNil(t, err)
	})

	t.Run("get_price_cache", func(t *testing.T) {
		cfg := config.Deyes{
			DbHost:   "127.0.0.1",
			DbSchema: "deyes",
			InMemory: true,
			EthRpcs:  []string{"http://example.com"},
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
			GetPriceFromSushiswapFunc: func(tokenAddress1 string, tokenAddress2 string, amount *big.Int) (
				*big.Int, error) {
				price := new(big.Int)
				if count == 0 {
					count++
					price, _ = new(big.Int).SetString("10000000000", 10)
				} else {
					price, _ = new(big.Int).SetString("20000000000", 10)
				}

				return price, nil
			},
		}

		uniswap := &uniswap.MockNewUniwapManager{
			GetPriceFromUniswapFunc: func(tokenAddress1 string, tokenAddress2 string) (*big.Int, error) {
				price := new(big.Int)
				if count == 0 {
					count++
					price, _ = new(big.Int).SetString("10000000000", 10)
				} else {
					price, _ = new(big.Int).SetString("20000000000", 10)
				}

				return price, nil
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
