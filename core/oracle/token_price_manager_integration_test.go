package oracle

import (
	"os"
	"testing"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/lib/log"
	"github.com/stretchr/testify/require"
)

func TestTokenPriceManager(t *testing.T) {
	t.Skip()

	providerCfgs := map[string]config.PriceProvider{
		"coin_cap": {
			Url: "https://api.coincap.io/v2/rates",
		},
		"coin_market_cap": {
			Url:    "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest",
			Secret: "", // Add your secret here.
		},
	}
	tokens := map[string]config.Token{
		"ETH": {
			Symbol:        "ETH",
			NameLowerCase: "ethereum",
		},
	}

	tpm := NewTokenPriceManager(providerCfgs, tokens, network.NewHttp())
	price, err := tpm.GetPrice("ETH")
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestCoinCapProvider(t *testing.T) {
	t.Skip()

	p := NewCoinCapProvider(network.NewHttp(), config.PriceProvider{
		Url:    "https://api.coincap.io/v2/rates",
		Secret: os.Getenv("SECRET"),
	})
	price, err := p.GetPrice(config.Token{NameLowerCase: "binance-coin"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestCoinMarketCap(t *testing.T) {
	t.Skip()

	p := NewCoinMarketCap(network.NewHttp(), config.PriceProvider{
		Url:    "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest",
		Secret: os.Getenv("SECRET"),
	})

	price, err := p.GetPrice(config.Token{Symbol: "ETH"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}
