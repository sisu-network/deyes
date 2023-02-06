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
			Url:     "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest",
			Secrets: "", // Add your secret here.
		},
		"coin_brain": {
			Url: "https://api.coinbrain.com/public/coin-info",
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
	// t.Skip()

	p := NewCoinCapProvider(network.NewHttp(), config.PriceProvider{
		Url:     "https://api.coincap.io/v2/rates",
		Secrets: "0d0cd537-2509-4386-bddf-ac07505804f0",
	})
	price, err := p.GetPrice(config.Token{NameLowerCase: "binance-coin"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestCoinBrainProvider(t *testing.T) {
	// t.Skip()

	p := NewCoinBrainProvider(network.NewHttp(), config.PriceProvider{
		Url: "https://api.coinbrain.com/public/coin-info",
	})
	price, err := p.GetPrice(config.Token{NameLowerCase: "binance-coin", ChainId: "56", Address: "0x1CE0c2827e2eF14D5C4f29a091d735A204794041"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestCoingeckoProvider(t *testing.T) {
	// t.Skip()

	p := NewCoingeckoProvider(network.NewHttp(), config.PriceProvider{
		Url: "https://api.coingecko.com/api/v3/simple/price",
	})
	price, err := p.GetPrice(config.Token{NameLowerCase: "ethereum"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestCoinMarketCap(t *testing.T) {
	t.Skip()

	p := NewCoinMarketCap(network.NewHttp(), config.PriceProvider{
		Url:     "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest",
		Secrets: os.Getenv("SECRET"),
	})

	price, err := p.GetPrice(config.Token{Symbol: "ETH"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}
