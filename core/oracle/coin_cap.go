package oracle

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/deyes/utils"
)

type CoinCapProvider struct {
	providerCfg config.PriceProvider
	networkHttp network.Http
}

func NewCoinCapProvider(networkHttp network.Http, providerCfg config.PriceProvider) Provider {
	return &CoinCapProvider{
		networkHttp: networkHttp,
		providerCfg: providerCfg,
	}
}

func (p *CoinCapProvider) GetPrice(token config.Token) (*big.Int, error) {
	fmt.Println("token = ", token)

	if token.NameLowerCase == "" {
		return nil, fmt.Errorf("Empty token lowercase name in coin cap, symbol = %s", token.Symbol)
	}

	baseUrl := fmt.Sprintf("%s/%s", p.providerCfg.Url, token.NameLowerCase)
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()

	type Response struct {
		Data struct {
			RateUsd string `json:"rateUsd"`
		} `json:"data"`
	}

	data, err := p.networkHttp.Get(req)
	if err != nil {
		return nil, err
	}

	response := &Response{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return utils.UsdToSisuPrice(response.Data.RateUsd)
}
