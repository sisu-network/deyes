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

type DefaultTokenPriceManager struct {
	cfg         config.Deyes
	stop        atomic.Value
	db          database.Database
	networkHttp network.Http
	cache       *sync.Map
}

func NewTokenPriceManager(cfg config.Deyes, db database.Database, networkHttp network.Http) TokenPriceManager {
	return &DefaultTokenPriceManager{
		cfg:         cfg,
		db:          db,
		networkHttp: networkHttp,
		cache:       &sync.Map{},
	}
}

func (m *DefaultTokenPriceManager) Start() {
	m.stop.Store(false)
	m.initTokenPrices()

	for {
		if m.stop.Load().(bool) == true {
			break
		}

		if len(m.cfg.PriceOracleUrl) == 0 {
			log.Warn("Orable price url is not set")
			break
		}

		tokenPrices, err := m.getResponse(m.cfg.PriceTokenList)
		if err == nil {
			// Update the database.
			m.db.SaveTokenPrices(tokenPrices)
		} else {
			log.Error("Cannot get response, err = ", err)
		}

		time.Sleep(time.Second * time.Duration(m.cfg.PricePollFrequency))
	}
}

// initTokenPrices loads prices from db and store in-memory. If the db is empty, take the default
// prices.
func (m *DefaultTokenPriceManager) initTokenPrices() {
	prices := m.db.LoadPrices()
	if len(prices) == 0 {
		prices = getDefaultTokenPriceList()
	}

	for _, price := range prices {
		m.cache.Store(price.Id, &priceCache{
			id:    price.Id,
			price: price.Price,
		})
	}
}

func (m *DefaultTokenPriceManager) getRequest(tokenList []string) *http.Request {
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

func (m *DefaultTokenPriceManager) getResponse(tokenList []string) ([]*types.TokenPrice, error) {
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

	for key, value := range response.Data {
		m.cache.Store(key, &types.TokenPrice{
			Id:    key,
			Price: utils.FloatToWei(value.Quote.Usd.Value),
		})
	}

	return tokenPrices, nil
}

func (m *DefaultTokenPriceManager) Stop() {
	m.stop.Store(true)
}

func (m *DefaultTokenPriceManager) GetTokenPrice(id string) (*big.Int, error) {
	if TestTokenPrices[id] != nil {
		return TestTokenPrices[id], nil
	}

	value, ok := m.cache.Load(id)
	if ok {
		// check expiration time
		now := time.Now()
		cache, ok := value.(*priceCache)
		if ok {
			if cache.updateTime+UpdateFrequency > now.UnixMilli() {
				return cache.price, nil
			}
		}
	}

	// Load from server.
	tokenPrices, err := m.getResponse([]string{id})
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
