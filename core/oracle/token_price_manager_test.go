package oracle

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/network"
	"github.com/stretchr/testify/require"
)

func TestStartManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

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
			PricePollFrequency: 1,
			PriceOracleUrl:     "http://example.com",

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
}
