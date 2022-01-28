package oracle

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sisu-network/deyes/config"
	mockdb "github.com/sisu-network/deyes/tests/mock/database"
	mocknetwork "github.com/sisu-network/deyes/tests/mock/network"
	"github.com/sisu-network/deyes/types"
	"github.com/stretchr/testify/require"
)

func TestStartManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

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
	mockDb.EXPECT().SaveTokenPrices(gomock.Any()).Times(1)

	updateCh := make(chan types.TokenPrices)

	priceManager := NewTokenPriceManager(cfg, mockDb, updateCh, mockNetwork)
	go priceManager.Start()

	result := <-updateCh

	require.Equal(t, 2, len(result))
}
