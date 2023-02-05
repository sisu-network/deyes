package oracle

import (
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/deyes/types"
)

const (
	UpdateFrequency = 1000 * 60 * 60 // 1 hour
)

type priceCache struct {
	id         string
	price      *big.Int
	updateTime int64
}

type Response struct {
	Data map[string]struct {
		Quote struct {
			Usd struct {
				Value float64 `json:"price"`
			} `json:"USD"`
		} `json:"quote"`
	} `json:"data"`
}

type TokenPriceManager interface {
	Start()
	Stop()
	GetTokenPrice(id string) (*big.Int, error)
}

type defaultTokenPriceManager struct {
	cfg             config.Deyes
	stop            atomic.Value
	networkHttp     network.Http
	cache           *sync.Map
	updateFrequency int64
}

func NewTokenPriceManager(cfg config.Deyes, networkHttp network.Http) TokenPriceManager {
	return &defaultTokenPriceManager{
		cfg:             cfg,
		networkHttp:     networkHttp,
		cache:           &sync.Map{},
		updateFrequency: UpdateFrequency,
	}
}

func (m *defaultTokenPriceManager) Start() {
	m.stop.Store(false)
}

func (m *defaultTokenPriceManager) getPriceFromCoinmarketcap(tokenList []string) *http.Request {
	baseUrl := m.cfg.PriceOracleUrl
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("X-CMC_PRO_API_KEY", m.cfg.PriceOracleSecret)

	q := req.URL.Query()
	q.Add("symbol", strings.Join(tokenList, ","))
	req.URL.RawQuery = q.Encode()

	return req
}

func (m *defaultTokenPriceManager) getTokenPrices(tokenList []string) ([]*types.TokenPrice, error) {
	return nil, nil
}

func (m *defaultTokenPriceManager) Stop() {
	m.stop.Store(true)
}

func (m *defaultTokenPriceManager) GetTokenPrice(id string) (*big.Int, error) {
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
	tokenPrices, err := m.getTokenPrices([]string{id})
	if err != nil {
		return nil, err
	}

	if len(tokenPrices) == 0 {
		return nil, fmt.Errorf("Cannot get token prices for %s", id)
	}

	for _, tokenPrice := range tokenPrices {
		if tokenPrice.Id == id {
			return tokenPrice.Price, nil
		}
	}

	return nil, fmt.Errorf("Cannot get token prices for %s", id)
}
