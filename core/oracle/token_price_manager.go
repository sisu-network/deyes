package oracle

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sisu-network/deyes/core/oracle/sushiswap"
	"github.com/sisu-network/deyes/core/oracle/uniswap"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

const (
	UpdateFrequency = 1000 * 60 * 3 // 3 minutes
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
	cfg              config.Deyes
	stop             atomic.Value
	networkHttp      network.Http
	cache            *sync.Map
	updateFrequency  int64
	uniswapManager   uniswap.UniswapManager
	sushiswapManager sushiswap.SushiSwapManager
}

func NewTokenPriceManager(cfg config.Deyes, networkHttp network.Http,
	uniswapManager uniswap.UniswapManager, sushiswapManager sushiswap.SushiSwapManager) TokenPriceManager {
	return &defaultTokenPriceManager{
		cfg:              cfg,
		networkHttp:      networkHttp,
		cache:            &sync.Map{},
		updateFrequency:  UpdateFrequency,
		uniswapManager:   uniswapManager,
		sushiswapManager: sushiswapManager,
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
	tokenPrices := make([]*types.TokenPrice, 0)
	tokensNotAvailable := make([]string, 0)
	tokens := m.cfg.EthTokens

	for _, token := range tokenList {
		address1 := tokens[strings.ToLower(token)].Address1
		address2 := tokens[strings.ToLower(token)].Address2

		tokenPrice, err := m.uniswapManager.GetPriceFromUniswap(address1, address2, token)
		if err != nil {
			// Get from SushiSwap.
			tokenPrice, err = m.sushiswapManager.GetPriceFromSushiswap(address1, address2, token)
			if err != nil {
				tokensNotAvailable = append(tokensNotAvailable, token)
			} else {
				tokenPrices = append(tokenPrices, tokenPrice)
			}
		} else {
			tokenPrices = append(tokenPrices, tokenPrice)
		}
	}

	// Get price from coin marketcap
	if len(tokensNotAvailable) > 0 {
		req := m.getPriceFromCoinmarketcap(tokensNotAvailable)
		data, err := m.networkHttp.Get(req)
		if err != nil {
			return nil, err
		}

		response := &Response{}
		err = json.Unmarshal(data, &response)
		if err != nil {
			return nil, err
		}

		if len(response.Data) > 0 {
			for key, value := range response.Data {
				for _, token := range tokenList {
					if key == token {
						tokenPrices = append(tokenPrices, &types.TokenPrice{
							Id:    token,
							Price: utils.FloatToWei(value.Quote.Usd.Value),
						})

						break
					}
				}
			}
		}
	}

	now := time.Now()
	for key, value := range tokenPrices {
		m.cache.Store(key, &priceCache{
			id:         value.Id,
			price:      value.Price,
			updateTime: now.UnixMilli(),
		})

	}

	return tokenPrices, nil
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
