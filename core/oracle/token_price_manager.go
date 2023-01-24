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

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
	"github.com/sisu-network/lib/log"
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
	cfg             config.Deyes
	stop            atomic.Value
	db              database.Database
	networkHttp     network.Http
	cache           *sync.Map
	updateFrequency int64
}

func NewTokenPriceManager(cfg config.Deyes, db database.Database, networkHttp network.Http) TokenPriceManager {
	return &defaultTokenPriceManager{
		cfg:             cfg,
		db:              db,
		networkHttp:     networkHttp,
		cache:           &sync.Map{},
		updateFrequency: UpdateFrequency,
	}
}

func (m *defaultTokenPriceManager) Start() {
	m.stop.Store(false)
	m.initTokenPrices()

	_, err := m.getTokenPrices(m.cfg.PriceTokenList)
	if err != nil {
		log.Error("Cannot get response, err = ", err)
	}
}

// initTokenPrices loads prices from db and store in-memory. If the db is empty, take the default
// prices.
func (m *defaultTokenPriceManager) initTokenPrices() {
	prices := m.db.LoadPrices()
	if len(prices) == 0 {
		prices = getDefaultTokenPriceList()
	}

	for _, price := range prices {
		m.cache.Store(price.Id, &priceCache{
			id:         price.Id,
			price:      price.Price,
			updateTime: 0,
		})
	}
}

func (m *defaultTokenPriceManager) getRequest(tokenList []string) *http.Request {
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
	req := m.getRequest(tokenList)
	data, err := m.networkHttp.Get(req)
	if err != nil {
		return nil, err
	}

	response := &Response{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	tokenPrices := make([]*types.TokenPrice, 0)
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

	// Save all data to the cached & db
	if len(tokenPrices) > 0 {
		m.db.SaveTokenPrices(tokenPrices)
	}

	now := time.Now()
	for key, value := range response.Data {
		m.cache.Store(key, &priceCache{
			id:         key,
			price:      utils.FloatToWei(value.Quote.Usd.Value),
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
