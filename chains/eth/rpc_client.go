package eth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sisu-network/lib/log"

	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/net/html"
)

type RrpcClient interface {
	GetExtraRpcs(chainId int) ([]string, error)
}

type defaultRpcClient struct {
	rpcs    []string
	clients []*ethclient.Client
}

func NewRpcChecker(initialRpcs []string) RrpcClient {
	clients := make([]*ethclient.Client, 0, len(initialRpcs))
	for _, rpc := range initialRpcs {
		client, err := ethclient.Dial(rpc)
		if err != nil {
			continue
		}

		if err == nil {
			clients = append(clients, client)
			log.Info("Adding eth client at rpc: ", rpc)
		}
	}

	return &defaultRpcClient{
		rpcs:    initialRpcs,
		clients: clients,
	}
}

func (drc *defaultRpcClient) processData(text string) []string {
	tokenizer := html.NewTokenizer(strings.NewReader(text))
	var data string
	for {
		tokenType := tokenizer.Next()
		stop := false
		switch tokenType {
		case html.ErrorToken:
			stop = true
			break

		case html.TextToken:
			text := tokenizer.Token().Data
			var js json.RawMessage
			if json.Unmarshal([]byte(text), &js) == nil {
				data = text
				break
			}
		}

		if stop {
			break
		}
	}

	// Process the data
	type result struct {
		Props struct {
			PageProps struct {
				Chain struct {
					Name string `json:"name"`
					RPC  []struct {
						Url string `json:"url"`
					} `json:"rpc"`
				} `json:"chain"`
			} `json:"pageProps"`
		} `json:"props"`
	}

	r := &result{}
	err := json.Unmarshal([]byte(data), r)
	if err != nil {
		panic(err)
	}

	ret := make([]string, 0)
	for _, rpc := range r.Props.PageProps.Chain.RPC {
		ret = append(ret, rpc.Url)
	}

	return ret
}

func (drc *defaultRpcClient) GetExtraRpcs(chainId int) ([]string, error) {
	url := fmt.Sprintf("https://chainlist.org/chain/%d", chainId)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Failed to get chain list data, status code = %d", res.StatusCode)
	}

	bz, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	ret := drc.processData(string(bz))

	return ret, nil
}
