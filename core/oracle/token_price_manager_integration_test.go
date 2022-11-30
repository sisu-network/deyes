package oracle

import (
	"fmt"
	"testing"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/deyes/types"
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

	outCh := make(chan []*types.TokenPrice)
	tpm := NewTokenPriceManager(cfg, dbInstance, network.NewHttp())
	go tpm.Start(outCh)

	select {
	case prices := <-outCh:
		for _, price := range prices {
			fmt.Println(price.Id, " ", price.Price)
		}

		tpm.Stop()
	}
}
