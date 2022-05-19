package oracle

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sisu-network/deyes/config"
	mockdb "github.com/sisu-network/deyes/tests/mock/database"
	mocknetwork "github.com/sisu-network/deyes/tests/mock/network"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
	"github.com/stretchr/testify/require"
)

func TestStartManager(t *testing.T) {
	t.Parallel()

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
		}

		mockNetwork := mocknetwork.NewMockHttp(ctrl)
		mockNetwork.EXPECT().Get(gomock.Any()).
			Return([]byte(`{"data":{"BTC":{"quote":{"USD":{"price":36367.076791144566}}},"ETH":{"quote":{"USD":{"price":2410.875945408672}}}}}`), nil).
			Times(1)

		mockDb := mockdb.NewMockDatabase(ctrl)
		mockDb.EXPECT().LoadPrices().Times(1)
		mockDb.EXPECT().SaveTokenPrices(gomock.Any()).Times(1)

		updateCh := make(chan []*types.TokenPrice)

		priceManager := NewTokenPriceManager(cfg, mockDb, mockNetwork)
		go priceManager.Start(updateCh)

		result := <-updateCh

		require.Equal(t, 2, len(result))
	})

	t.Run("Get price fail", func(t *testing.T) {
		cfg := config.Deyes{
			PricePollFrequency: 1,
			PriceOracleUrl:     "http://example.com",
			PriceTokenList:     []string{"INVALID_TOKEN"},
		}

		mockNetwork := mocknetwork.NewMockHttp(ctrl)
		mockNetwork.EXPECT().Get(gomock.Any()).
			Return([]byte(`nothing`), nil).
			AnyTimes()

		mockDb := mockdb.NewMockDatabase(ctrl)
		mockDb.EXPECT().LoadPrices().Times(1)
		updateCh := make(chan []*types.TokenPrice)

		priceManager := NewTokenPriceManager(cfg, mockDb, mockNetwork)

		timeOut := time.After(3 * time.Second)
		go priceManager.Start(updateCh)

		select {
		case <-timeOut:
			log.Debug("API is in maintain or token is not found")
		case <-updateCh:
			t.Fail()
		}
	})
}
