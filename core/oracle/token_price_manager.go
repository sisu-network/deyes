package oracle

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

type Response struct {
	Data map[string]struct {
		Quote struct {
			Usd struct {
				Value float32 `json:"price"`
			} `json:"USD"`
		} `json:"quote"`
	} `json:"data"`
}

type TokenPriceManager interface {
	Start()
	Stop()
}

type DefaultTokenPriceManager struct {
	cfg         config.Deyes
	updateCh    chan types.TokenPrices
	stop        atomic.Value
	db          database.Database
	networkHttp network.Http
}

func NewTokenPriceManager(cfg config.Deyes, db database.Database, updateCh chan types.TokenPrices, networkHttp network.Http) TokenPriceManager {
	return &DefaultTokenPriceManager{
		cfg:         cfg,
		updateCh:    updateCh,
		db:          db,
		networkHttp: networkHttp,
	}
}

func (m *DefaultTokenPriceManager) Start() {
	req := m.getRequest()
	m.stop.Store(false)

	for {
		if m.stop.Load().(bool) == true {
			break
		}

		if len(m.cfg.PriceOracleUrl) == 0 {
			log.Warn("Orable price url is not set")
			break
		}

		response := m.getResponse(req)
		if response != nil {
			tokenPrices := make([]*types.TokenPrice, 0)
			for _, token := range m.cfg.PriceTokenList {
				for key, value := range response.Data {
					if key == token {
						tokenPrice := &types.TokenPrice{
							Id:       token,
							PublicId: token,
							Price:    value.Quote.Usd.Value,
						}

						tokenPrices = append(tokenPrices, tokenPrice)
						break
					}
				}
			}

			// Update the database.
			m.db.SaveTokenPrices(tokenPrices)

			// Broadcast the result we have.
			m.updateCh <- tokenPrices
		}

		time.Sleep(time.Second * time.Duration(m.cfg.PricePollFrequency))
	}
}

func (m *DefaultTokenPriceManager) getRequest() *http.Request {
	url := m.cfg.PriceOracleUrl
	tokenList := m.cfg.PriceTokenList

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Add("symbol", strings.Join(tokenList, ","))
	req.URL.RawQuery = q.Encode()

	return req
}

func (m *DefaultTokenPriceManager) getResponse(req *http.Request) *Response {
	data, err := m.networkHttp.Get(req)
	if err != nil {
		return nil
	}

	target := &Response{}
	err = json.Unmarshal(data, &target)
	if err != nil {
		return nil
	}

	return target
}

func (m *DefaultTokenPriceManager) Stop() {
	m.stop.Store(true)
}
