package oracle

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	mocknetwork "github.com/sisu-network/deyes/tests/mock/network"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
	"github.com/stretchr/testify/require"
)

func TestStartManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	t.Run("Get price success", func(t *testing.T) {
		t.Parallel()

		cfg := config.Deyes{
			PricePollFrequency: 1,
			PriceOracleUrl:     "http://example.com",
			PriceTokenList:     []string{"ETH", "BTC"},

			DbHost:   "127.0.0.1",
			DbSchema: "deyes",
			InMemory: true,
		}

		mockNetwork := mocknetwork.NewMockHttp(ctrl)
		mockNetwork.EXPECT().Get(gomock.Any()).
			Return([]byte(`{"data":{"BTC":{"quote":{"USD":{"price":36367.076791144566}}},"ETH":{"quote":{"USD":{"price":2410.875945408672}}}}}`), nil).
			Times(1)

		dbInstance := database.NewDb(&cfg)
		err := dbInstance.Init()
		if err != nil {
			panic(err)
		}

		updateCh := make(chan []*types.TokenPrice)

		priceManager := NewTokenPriceManager(cfg, dbInstance, mockNetwork)
		go priceManager.Start(updateCh)
		defer priceManager.Stop()

		result := <-updateCh

		require.Equal(t, 2, len(result))
	})

	t.Run("Get price fail", func(t *testing.T) {
		cfg := config.Deyes{
			PricePollFrequency: 1,
			PriceOracleUrl:     "http://example.com",
			PriceTokenList:     []string{"INVALID_TOKEN"},

			DbHost:   "127.0.0.1",
			DbSchema: "deyes",
			InMemory: true,
		}

		mockNetwork := mocknetwork.NewMockHttp(ctrl)
		mockNetwork.EXPECT().Get(gomock.Any()).
			Return([]byte(`nothing`), nil).
			AnyTimes()

		dbInstance := database.NewDb(&cfg)
		err := dbInstance.Init()
		if err != nil {
			panic(err)
		}

		updateCh := make(chan []*types.TokenPrice)

		priceManager := NewTokenPriceManager(cfg, dbInstance, mockNetwork)

		timeOut := time.After(3 * time.Second)
		go priceManager.Start(updateCh)

		select {
		case <-timeOut:
			log.Error("API is in maintain or token is not found")
		case <-updateCh:
			t.Fail()
		}
	})
}
