package oracle

import (
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/lib/log"
)

const (
	UpdateFrequency = 1000 * 60 * 60 // 1 hour
)

type priceCache struct {
	id         string
	price      *big.Int
	updateTime int64
}

type TokenPriceManager interface {
	GetPrice(id string) (*big.Int, error)
}

type defaultTokenPriceManager struct {
	priceProviderCfgs map[string]config.PriceProvider
	networkHttp       network.Http
	cache             *sync.Map
	updateFrequency   int64
	providers         map[string]Provider
	tokens            map[string]config.Token
}

func NewTokenPriceManager(providerCfgs map[string]config.PriceProvider,
	tokens map[string]config.Token, networkHttp network.Http) TokenPriceManager {

	providers := make(map[string]Provider)
	for name, providerCfg := range providerCfgs {
		switch name {
		case "coin_cap":
			provider := NewCoinCapProvider(networkHttp, providerCfg)
			providers[name] = provider

		case "coin_market_cap":
			provider := NewCoinMarketCap(networkHttp, providerCfg)
			providers[name] = provider

		case "coin_brain":
			provider := NewCoinBrainProvider(networkHttp, providerCfg)
			providers[name] = provider

		case "coingecko":
			provider := NewCoingeckoProvider(networkHttp, providerCfg)
			providers[name] = provider

		case "portal_fi":
			provider := NewPortalFiProvider(networkHttp, providerCfg)
			providers[name] = provider

		default:
			log.Errorf("Unknown price provider %s", name)
		}
	}

	return &defaultTokenPriceManager{
		priceProviderCfgs: providerCfgs,
		networkHttp:       networkHttp,
		cache:             &sync.Map{},
		updateFrequency:   UpdateFrequency,
		tokens:            tokens,
		providers:         providers,
	}
}

func (m *defaultTokenPriceManager) getTokenPrices(id string) (*big.Int, error) {
	token, ok := m.tokens[id]
	if !ok {
		return nil, fmt.Errorf("Token %s not supported", id)
	}

	priceMap := &sync.Map{}
	wg := &sync.WaitGroup{}
	// TODO: Run each provider with a timeout.
	for name, provider := range m.providers {
		wg.Add(1)
		go func(name string, provider Provider) {
			defer wg.Done()

			price, err := provider.GetPrice(token)
			if err != nil {
				log.Errorf("Failed to get token price for provider %s, err = %s", name, err)
				return
			}

			priceMap.Store(name, price)
		}(name, provider)
	}
	wg.Wait()

	// Accumulate prices
	prices := make([]*big.Int, 0)
	priceMap.Range(func(key, value interface{}) bool { // name, price
		prices = append(prices, value.(*big.Int))

		return true
	})

	if len(prices) == 0 {
		return nil, fmt.Errorf("Cannot find price from any provider for token %s", id)
	}

	// Sort all prices
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Cmp(prices[j]) < 0
	})

	// Get the median
	return prices[len(prices)/2], nil
}

func (m *defaultTokenPriceManager) GetPrice(id string) (*big.Int, error) {
	if TestTokenPrices[id] != nil {
		return TestTokenPrices[id], nil
	}

	value, ok := m.cache.Load(id)
	if ok {
		// check expiration time
		now := time.Now()
		cache, ok := value.(*priceCache)
		if ok {
			if cache.updateTime+m.updateFrequency > now.UnixMilli() {
				return cache.price, nil
			}
		}
	}

	// Load from server.
	price, err := m.getTokenPrices(id)

	if err == nil {
		// Save into the cache
		m.cache.Store(id, &priceCache{
			id:         id,
			price:      price,
			updateTime: time.Now().UnixMilli(),
		})
	}

	return price, err
}
