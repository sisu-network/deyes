package oracle

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"strings"
	"time"

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
	if token.CoincapName == "" {
		return nil, fmt.Errorf("Empty token lowercase name in coin cap, symbol = %s", token.Symbol)
	}

	baseUrl := fmt.Sprintf("%s/%s", p.providerCfg.Url, token.CoincapName)

	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		panic(err)
	}

	secret := p.randomSecret()
	if len(secret) == 0 {
		return nil, fmt.Errorf("Invalid secret %s", p.providerCfg.Secrets)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secret))

	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()

	type Response struct {
		Data struct {
			PriceUsd string `json:"priceUsd"`
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

	return utils.UsdToSisuPrice(response.Data.PriceUsd)
}

func (p *CoinCapProvider) randomSecret() string {
	secrets := strings.Split(p.providerCfg.Secrets, ",")
	if len(secrets) == 0 {
		return ""
	}
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return secrets[r1.Intn(len(secrets))]
}
