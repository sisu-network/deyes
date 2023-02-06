package oracle

import (
	"bytes"
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

type CoinBrainProvider struct {
	providerCfg config.PriceProvider
	networkHttp network.Http
}

func NewCoinBrainProvider(networkHttp network.Http, providerCfg config.PriceProvider) Provider {
	return &CoinBrainProvider{
		networkHttp: networkHttp,
		providerCfg: providerCfg,
	}
}

func (p *CoinBrainProvider) GetPrice(token config.Token) (*big.Int, error) {
	if token.NameLowerCase == "" {
		return nil, fmt.Errorf("Empty token lowercase name in coin cap, symbol = %s", token.Symbol)
	}
	body := map[string][]string{token.ChainId: {token.Address}}
	jsonData, err := json.Marshal(body)
	req, err := http.NewRequest("POST", p.providerCfg.Url, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	secret := p.randomSecret()

	if len(secret) >= 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secret))
	}

	type Response struct {
		PriceUsd float32
	}

	data, err := p.networkHttp.Get(req)
	if err != nil {
		return nil, err
	}
	response := make([]Response, 0)
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(data), response[0])
	return utils.UsdToSisuPrice(fmt.Sprintf("%f", response[0].PriceUsd))
}

func (p *CoinBrainProvider) randomSecret() string {
	secrets := strings.Split(p.providerCfg.Secrets, ",")
	if len(secrets) == 0 {
		return ""
	}

	return secrets[rand.Intn(len(secrets))]
}
