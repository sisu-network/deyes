package oracle

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"strings"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/deyes/utils"
)

type CoinMarketCap struct {
	providerCfg config.PriceProvider
	networkHttp network.Http
}

func NewCoinMarketCap(networkHttp network.Http, providerCfg config.PriceProvider) Provider {
	return &CoinMarketCap{
		networkHttp: networkHttp,
		providerCfg: providerCfg,
	}
}

func (p *CoinMarketCap) GetPrice(token config.Token) (*big.Int, error) {
	baseUrl := p.providerCfg.Url
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		panic(err)
	}

	secret := p.randomSecret()
	if len(secret) == 0 {
		return nil, fmt.Errorf("Invalid secret %s", p.providerCfg.Secrets)
	}
	req.Header.Add("X-CMC_PRO_API_KEY", secret)

	q := req.URL.Query()
	q.Add("symbol", token.Symbol)
	req.URL.RawQuery = q.Encode()

	data, err := p.networkHttp.Get(req)
	if err != nil {
		return nil, err
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

	response := &Response{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	tokenPrice, ok := response.Data[token.Symbol]
	if !ok {
		return nil, fmt.Errorf("Token %s not found in the response %s", token.Symbol, string(data))
	}

	price := utils.FloatToWei(tokenPrice.Quote.Usd.Value)

	return price, nil
}

func (p *CoinMarketCap) randomSecret() string {
	secrets := strings.Split(p.providerCfg.Secrets, ",")
	if len(secrets) == 0 {
		return ""
	}

	return secrets[rand.Intn(len(secrets))]
}
