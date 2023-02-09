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
			Secrets: "", // Add your secret here.
		},
		"coin_market_cap": {
			Url:     "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest",
			Secrets: "", // Add your secret here.
		},
		"coingecko": {
			Url: "https://api.coingecko.com/api/v3/simple/price",
		},
	}
	tokens := map[string]config.Token{
		"ETH": {
			Symbol:        "ETH",
			CoincapName:   "ethereum",
			CoinGeckoName: "ethereum",
		},
		"AVAX": {
			Symbol:        "AVAX",
			CoincapName:   "avalanche",
			CoinGeckoName: "avalanche-2",
		},
		"FTM": {
			Symbol:        "FTM",
			CoincapName:   "fantom",
			CoinGeckoName: "fantom",
		},
		"SOL": {
			Symbol:        "SOL",
			CoincapName:   "solana",
			CoinGeckoName: "solana",
		},
		"BNB": {
			Symbol:        "BNB",
			CoincapName:   "binance-coin",
			CoinGeckoName: "binancecoin",
		},
		"MATIC": {
			Symbol:        "MATIC",
			CoincapName:   "polygon",
			CoinGeckoName: "matic-network",
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
	price, err := p.GetPrice(config.Token{CoincapName: "avalanche"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}

func TestCoingeckoProvider(t *testing.T) {
	t.Skip()

	p := NewCoingeckoProvider(network.NewHttp(), config.PriceProvider{
		Url: "https://api.coingecko.com/api/v3/simple/price",
	})
	price, err := p.GetPrice(config.Token{CoinGeckoName: "matic-network"})
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

	price, err := p.GetPrice(config.Token{Symbol: "MATIC"})
	require.Nil(t, err)

	log.Infof("Price = %s", price)
}
