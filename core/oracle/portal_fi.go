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

type PortalFiProvider struct {
	providerCfg config.PriceProvider
	networkHttp network.Http
}

func NewPortalFiProvider(networkHttp network.Http, providerCfg config.PriceProvider) Provider {
	return &PortalFiProvider{
		networkHttp: networkHttp,
		providerCfg: providerCfg,
	}
}

func (p *PortalFiProvider) GetPrice(token config.Token) (*big.Int, error) {
	if token.NameLowerCase == "" {
		return nil, fmt.Errorf("Empty token lowercase name in coin cap, symbol = %s", token.Symbol)
	}

	baseUrl := fmt.Sprintf("%s?addresses=%s:%s", p.providerCfg.Url, token.ChainName, token.Address)
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()

	type Response struct {
		Tokens []struct {
			Price float32 `json:"price"`
		} `json:"tokens"`
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

	return utils.UsdToSisuPrice(fmt.Sprintf("%f", response.Tokens[0].Price))
}

func (p *PortalFiProvider) randomSecret() string {
	secrets := strings.Split(p.providerCfg.Secrets, ",")
	if len(secrets) == 0 {
		return ""
	}
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return secrets[r1.Intn(len(secrets))]
}
