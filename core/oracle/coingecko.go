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

type CoingeckoProvider struct {
	providerCfg config.PriceProvider
	networkHttp network.Http
}

func NewCoingeckoProvider(networkHttp network.Http, providerCfg config.PriceProvider) Provider {
	return &CoingeckoProvider{
		networkHttp: networkHttp,
		providerCfg: providerCfg,
	}
}

func (p *CoingeckoProvider) GetPrice(token config.Token) (*big.Int, error) {
	if token.CoincapName == "" {
		return nil, fmt.Errorf("Empty token lowercase name in coin cap, symbol = %s", token.Symbol)
	}

	baseUrl := fmt.Sprintf("%s?ids=%s&vs_currencies=usd", p.providerCfg.Url, token.CoincapName)
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()

	type Response struct {
		USD float32 `json:"usd"`
	}

	data, err := p.networkHttp.Get(req)
	if err != nil {
		return nil, err
	}

	response := map[string]Response{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return utils.UsdToSisuPrice(fmt.Sprintf("%f", response[token.CoincapName].USD))
}

func (p *CoingeckoProvider) randomSecret() string {
	secrets := strings.Split(p.providerCfg.Secrets, ",")
	if len(secrets) == 0 {
		return ""
	}

	return secrets[rand.Intn(len(secrets))]
}
