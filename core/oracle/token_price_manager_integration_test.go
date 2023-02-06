package oracle

import (
	"fmt"
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
			Url:     "https://api.coincap.io/v2/assets",
			Secrets: "0426cc1a-2517-47b4-a4fc-16aa14281506",
		},
		"coin_market_cap": {
			Url:     "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest",
			Secrets: "1d021473-1b46-4154-90a1-26ef3ea3bcbf", // Add your secret here.
		},
		"coin_brain": {
			Url: "https://api.coinbrain.com/public/coin-info",
		},
		"coingecko": {
			Url: "https://api.coingecko.com/api/v3/simple/price",
		},
		"portal_fi": {
			Url: "https://api.portals.fi/v2/tokens",
		},
	}
	tokens := map[string]config.Token{
		"ETH": {
			Symbol:        "ETH",
			NameLowerCase: "ethereum",
			ChainId:       "1",
			Address:       "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
			ChainName:     "ethereum",
		},
		"AVAX": {
			Symbol:        "AVAX",
			NameLowerCase: "avalanche",
			ChainId:       "43114",
			Address:       "0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7",
			ChainName:     "avalanche",
		},
		"FTM": {
			Symbol:        "FTM",
			NameLowerCase: "fantom",
			ChainId:       "1",
			Address:       "0x4E15361FD6b4BB609Fa63C81A2be19d873717870",
			ChainName:     "ethereum",
		},
		"SOL": {
			Symbol:        "SOL",
			NameLowerCase: "solana",
			ChainId:       "56",
			Address:       "0x570A5D26f7765Ecb712C0924E4De545B89fD43dF",
			ChainName:     "bsc",
		},
		"BNB": {
			Symbol:        "BNB",
			NameLowerCase: "binance-coin",
			ChainId:       "56",
			Address:       "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
			ChainName:     "bsc",
		},
		"MATIC": {
			Symbol:        "MATIC",
			NameLowerCase: "polygon",
			ChainId:       "1",
			Address:       "0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0",
			ChainName:     "ethereum",
		},
	}

	tpm := NewTokenPriceManager(providerCfgs, tokens, network.NewHttp())
	price, err := tpm.GetPrice("MATIC")
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestCoinCapProvider(t *testing.T) {
	t.Skip()

	p := NewCoinCapProvider(network.NewHttp(), config.PriceProvider{
		Url:     "https://api.coincap.io/v2/assets",
		Secrets: os.Getenv("COIN_CAP_SECRET"),
	})
	price, err := p.GetPrice(config.Token{NameLowerCase: "avalanche"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestCoinBrainProvider(t *testing.T) {
	t.Skip()

	p := NewCoinBrainProvider(network.NewHttp(), config.PriceProvider{
		Url: "https://api.coinbrain.com/public/coin-info",
	})
	price, err := p.GetPrice(config.Token{NameLowerCase: "binance-coin", ChainId: "56", Address: "0x1CE0c2827e2eF14D5C4f29a091d735A204794041"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestCoingeckoProvider(t *testing.T) {
	t.Skip()

	p := NewCoingeckoProvider(network.NewHttp(), config.PriceProvider{
		Url: "https://api.coingecko.com/api/v3/simple/price",
	})
	price, err := p.GetPrice(config.Token{NameLowerCase: "binance-coin"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestPortalFiProvider(t *testing.T) {
	t.Skip()
	p := NewPortalFiProvider(network.NewHttp(), config.PriceProvider{
		Url: "https://api.portals.fi/v2/tokens",
	})
	price, err := p.GetPrice(config.Token{NameLowerCase: "ethereum", ChainName: "ethereum", Address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestCoinMarketCap(t *testing.T) {
	t.Skip()
	fmt.Println(os.Getenv("COIN_MARKET_CAP_SECRET"))
	p := NewCoinMarketCap(network.NewHttp(), config.PriceProvider{
		Url:     "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest",
		Secrets: os.Getenv("COIN_MARKET_CAP_SECRET"),
	})

	price, err := p.GetPrice(config.Token{Symbol: "AVAX"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}
