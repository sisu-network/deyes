package oracle

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/network"
	"github.com/stretchr/testify/require"
)

func TestTokenManager(t *testing.T) {
	t.Run("Get price success", func(t *testing.T) {
		cfg := config.Deyes{
			PricePollFrequency: 1,
			PriceOracleUrl:     "http://example.com",
			PriceTokenList:     []string{"ETH", "BTC"},

			DbHost:   "127.0.0.1",
			DbSchema: "deyes",
			InMemory: true,
		}

		net := &network.MockHttp{
			GetFunc: func(req *http.Request) ([]byte, error) {
				return []byte(`{"data":{"BTC":{"quote":{"USD":{"price":36367.076791144566}}},"ETH":{"quote":{"USD":{"price":2410.875945408672}}}}}`), nil
			},
		}

		dbInstance := database.NewDb(&cfg)
		err := dbInstance.Init()
		if err != nil {
			panic(err)
		}

		priceManager := NewTokenPriceManager(cfg, dbInstance, net)
		price, err := priceManager.GetTokenPrice("ETH")
		require.Nil(t, err)
		require.Equal(t, "2410875945408672038912", price.String())

		price, err = priceManager.GetTokenPrice("BTC")
		require.Nil(t, err)
		require.Equal(t, "36367076791144564129792", price.String())
	})

	t.Run("Get price fail", func(t *testing.T) {
		cfg := config.Deyes{
			DbHost:   "127.0.0.1",
			DbSchema: "deyes",
			InMemory: true,
		}

		net := &network.MockHttp{
			GetFunc: func(req *http.Request) ([]byte, error) {
				return nil, fmt.Errorf("Not found")
			},
		}

		dbInstance := database.NewDb(&cfg)
		err := dbInstance.Init()
		require.Nil(t, err)

		priceManager := NewTokenPriceManager(cfg, dbInstance, net)
		_, err = priceManager.GetTokenPrice("ETH")
		require.NotNil(t, err)
	})

	t.Run("token price cache", func(t *testing.T) {
		cfg := config.Deyes{
			DbHost:   "127.0.0.1",
			DbSchema: "deyes",
			InMemory: true,
		}

		count := 0
		net := &network.MockHttp{
			GetFunc: func(req *http.Request) ([]byte, error) {
				if count == 0 {
					count++
					return []byte(`{"data":{"ETH":{"quote":{"USD":{"price":1000}}}}}`), nil
				}

				return []byte(`{"data":{"ETH":{"quote":{"USD":{"price":2000}}}}}`), nil
			},
		}

		dbInstance := database.NewDb(&cfg)
		err := dbInstance.Init()
		require.Nil(t, err)

		priceManager := NewTokenPriceManager(cfg, dbInstance, net)
		price, err := priceManager.GetTokenPrice("ETH")
		require.Nil(t, err)
		require.Equal(t, "1000000000000000000000", price.String())

		price, err = priceManager.GetTokenPrice("ETH")
		require.Nil(t, err)
		require.Equal(t, "1000000000000000000000", price.String())

		// Change the updateFrequency to invalidate cache.
		priceManager.(*defaultTokenPriceManager).updateFrequency = 0

		price, err = priceManager.GetTokenPrice("ETH")
		require.Nil(t, err)
		require.Equal(t, "2000000000000000000000", price.String())
	})
}
