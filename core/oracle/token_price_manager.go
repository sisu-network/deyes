package oracle

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/core/oracle/sushiswap"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"context"
	"github.com/daoleno/uniswapv3-sdk/examples/contract"
	"github.com/daoleno/uniswapv3-sdk/examples/helper"
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

	for {
		if m.stop.Load().(bool) == true {
			break
		}

		if len(m.cfg.PriceOracleUrl) == 0 {
			log.Warn("Orable price url is not set")
			break
		}

		_, err := m.getResponse(m.cfg.PriceTokenList)
		if err != nil {
			log.Error("Cannot get response, err = ", err)
		}

		time.Sleep(time.Second * time.Duration(m.cfg.PricePollFrequency))
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

func (m *defaultTokenPriceManager) getPriceFromUniswap(tokenAddress string) (*types.TokenPrice, error) {
	rpcEth := m.cfg.EthRpc
	daiTokenAddress := m.cfg.DaiTokenAddress
	client, err := ethclient.Dial(rpcEth)
	if err != nil {
		return nil, err
	}
	quoterContract, err := contract.NewUniswapv3Quoter(common.HexToAddress(helper.ContractV3Quoter), client)
	if err != nil {
		return nil, err
	}

	token0 := common.HexToAddress(tokenAddress)
	token1 := common.HexToAddress(daiTokenAddress) // DAI
	fee := big.NewInt(3000)
	amountIn := helper.FloatStringToBigInt("1.00", 18)
	sqrtPriceLimitX96 := big.NewInt(0)

	var out []interface{}
	rawCaller := &contract.Uniswapv3QuoterRaw{Contract: quoterContract}
	err = rawCaller.Call(nil, &out, "quoteExactInputSingle", token0, token1,
		fee, amountIn, sqrtPriceLimitX96)
	if err != nil {
		return nil, err
	}

	return &types.TokenPrice{
		Id:    tokenAddress,
		Price: out[0].(*big.Int),
	}, nil
}

func (m *defaultTokenPriceManager) getPriceFromSushiswap(tokenAddress string) (*types.TokenPrice, error) {
	rpcEth := m.cfg.EthRpc
	daiTokenAddress := m.cfg.DaiTokenAddress
	ctx := context.Background()

	ec, _ := ethclient.DialContext(ctx, rpcEth)

	client := sushiswap.NewClient(ec)
	price, err := client.GetExchangeAmount(big.NewInt(1), common.HexToAddress(tokenAddress),
		common.HexToAddress(daiTokenAddress))
	if err != nil {
		return nil, err
	}
	return &types.TokenPrice{
		Id:    tokenAddress,
		Price: price,
	}, nil
}

func (m *defaultTokenPriceManager) getResponse(tokenList []string) ([]*types.TokenPrice, error) {
	tokenPrices := make([]*types.TokenPrice, 0)
	tokensNotAvailable := make([]string, 0)
	tokens := m.cfg.Tokens

	for _, token := range tokenList {
		address := tokens[strings.ToLower(token)].Address
		tokenPrice, err := m.getPriceFromUniswap(address)

		if err != nil {
			tokenPrice, err = m.getPriceFromSushiswap(address)
			if err != nil {
				tokensNotAvailable = append(tokensNotAvailable, token)
			}
		}
		tokenPrices = append(tokenPrices, tokenPrice)
	}

	req := m.getRequest(tokensNotAvailable)
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
