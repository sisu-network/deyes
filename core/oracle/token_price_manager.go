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
	Start(outCh chan types.TokenPrices)
	Stop()
}

type DefaultTokenPriceManager struct {
	cfg         config.Deyes
	stop        atomic.Value
	db          database.Database
	networkHttp network.Http
}

func NewTokenPriceManager(cfg config.Deyes, db database.Database, networkHttp network.Http) TokenPriceManager {
	return &DefaultTokenPriceManager{
		cfg:         cfg,
		db:          db,
		networkHttp: networkHttp,
	}
}

func (m *DefaultTokenPriceManager) Start(outCh chan types.TokenPrices) {
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

		response, err := m.getResponse(req)
		if err == nil {
			tokenPrices := make([]*types.TokenPrice, 0)
			for _, token := range m.cfg.PriceTokenList {
				var tokenPrice *types.TokenPrice
				for key, value := range response.Data {
					if key == token {
						tokenPrice = &types.TokenPrice{
							Id:       token,
							PublicId: token,
							Price:    value.Quote.Usd.Value,
						}
						break
					}
				}

				if tokenPrice == nil {
					// Get default price
					if value, ok := DEFAULT_PRICES[token]; ok {
						// Get the default price
						tokenPrice = &types.TokenPrice{
							Id:       token,
							PublicId: token,
							Price:    value,
						}
					}
				}

				if tokenPrice != nil {
					tokenPrices = append(tokenPrices, tokenPrice)
				} else {
					log.Error("Cannot find price for token ", token)
				}
			}

			// Update the database.
			m.db.SaveTokenPrices(tokenPrices)

			// Broadcast the result we have.
			outCh <- tokenPrices
		} else {
			log.Error("Cannot get response, err = ", err)
		}

		time.Sleep(time.Second * time.Duration(m.cfg.PricePollFrequency))
	}
}

func (m *DefaultTokenPriceManager) getRequest() *http.Request {
	baseUrl := m.cfg.PriceOracleUrl
	tokenList := m.cfg.PriceTokenList

	log.Info("tokenList = ", tokenList)

	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("X-CMC_PRO_API_KEY", m.cfg.PriceOracleSecret)

	q := req.URL.Query()
	q.Add("symbol", strings.Join(tokenList, ","))
	req.URL.RawQuery = q.Encode()

	log.Info("url = ", req.URL.String())

	return req
}

func (m *DefaultTokenPriceManager) getResponse(req *http.Request) (*Response, error) {
	data, err := m.networkHttp.Get(req)
	if err != nil {
		return nil, err
	}

	target := &Response{}
	err = json.Unmarshal(data, &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (m *DefaultTokenPriceManager) Stop() {
	m.stop.Store(true)
}
