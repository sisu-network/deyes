package eth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/config"
	libchain "github.com/sisu-network/lib/chain"
	"github.com/sisu-network/lib/log"
	"golang.org/x/net/html"
)

type NoHealthyClientErr struct {
	chain string
}

func NewNoHealthyClientErr(chain string) error {
	return &NoHealthyClientErr{chain: chain}
}

func (e *NoHealthyClientErr) Error() string {
	return fmt.Sprintf("No healthy client for chain %s", e.chain)
}

// A wrapper around eth.client so that we can mock in watcher tests.
type EthClient interface {
	Start()

	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SendTransaction(ctx context.Context, tx *ethtypes.Transaction) error
	BalanceAt(ctx context.Context, from common.Address, block *big.Int) (*big.Int, error)
}

type defaultEthClient struct {
	chain           string
	useExternalRpcs bool

	clients     []*ethclient.Client
	healthies   []bool
	initialRpcs []string
	rpcs        []string

	lock *sync.RWMutex
}

func NewEthClients(initialRpcs []string, cfg config.Chain, useExternalRpcs bool) EthClient {
	c := &defaultEthClient{
		chain:           cfg.Chain,
		useExternalRpcs: useExternalRpcs,
		initialRpcs:     initialRpcs,
		lock:            &sync.RWMutex{},
	}

	return c
}

func (c *defaultEthClient) Start() {
	go c.loopCheck()
}

// loopCheck
func (c *defaultEthClient) loopCheck() {
	// Sleep time = 10 mins.
	sleepTime := time.Second * 60 * 30
	for {
		time.Sleep(sleepTime)
		c.updateRpcs()
	}
}

func (c *defaultEthClient) updateRpcs() {
	c.lock.RLock()
	rpcs := c.initialRpcs
	c.lock.RUnlock()

	if c.useExternalRpcs {
		// Get external rpcs.
		externals, err := c.GetExtraRpcs()
		if err != nil {
			log.Errorf("Failed to get external rpc info, err = ", err)
		} else {
			rpcs = append(rpcs, externals...)
		}
	}

	c.lock.RLock()
	oldClients := c.clients
	c.lock.RUnlock()

	rpcs, clients, healthies := c.getRpcsHealthiness(rpcs)

	// Close all the old clients
	c.lock.Lock()
	if oldClients != nil {
		for _, client := range oldClients {
			client.Close()
		}
	}

	c.rpcs, c.clients, c.healthies = rpcs, clients, healthies
	c.lock.Unlock()
}

func (c *defaultEthClient) getRpcsHealthiness(allRpcs []string) ([]string, []*ethclient.Client, []bool) {
	clients := make([]*ethclient.Client, 0)
	rpcs := make([]string, 0)
	healthies := make([]bool, 0)

	for _, rpc := range allRpcs {
		client, err := ethclient.Dial(rpc)
		if err == nil {
			block, err := client.BlockByNumber(context.Background(), nil)
			if err == nil && block.Number() != nil {
				clients = append(clients, client)
				rpcs = append(rpcs, rpc)
				healthies = append(healthies, true)
			}
		}
	}

	return rpcs, clients, healthies
}

func (c *defaultEthClient) processData(text string) []string {
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

func (c *defaultEthClient) GetExtraRpcs() ([]string, error) {
	chainId := libchain.GetChainIntFromId(c.chain)
	url := fmt.Sprintf("https://chainlist.org/chain/%d", chainId)
	log.Verbosef("Getting extra rpcs status from remote link %s for chain %s", url, c.chain)

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

	ret := c.processData(string(bz))

	return ret, nil
}

func (c *defaultEthClient) shuffle() ([]*ethclient.Client, []bool, []string) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	n := len(c.clients)

	clients := make([]*ethclient.Client, n)
	healthy := make([]bool, n)
	rpcs := make([]string, n)

	copy(clients, c.clients)
	copy(healthy, c.healthies)
	copy(rpcs, c.rpcs)

	for i := 0; i < 20; i++ {
		x := rand.Intn(n)
		y := rand.Intn(n)

		tmpClient := clients[x]
		clients[x] = clients[y]
		clients[y] = tmpClient

		tmpHealth := healthy[x]
		healthy[x] = healthy[y]
		healthy[y] = tmpHealth

		tmpRpc := rpcs[x]
		rpcs[x] = rpcs[y]
		rpcs[y] = tmpRpc
	}

	return clients, healthy, rpcs
}

func (c *defaultEthClient) getHealthyClient() (*ethclient.Client, string) {
	c.lock.RLock()
	if c.clients == nil {
		c.lock.RUnlock()
		c.updateRpcs()
	} else {
		c.lock.RUnlock()
	}

	// Shuffle rpcs so that we will use different healthy rpc
	clients, healthies, rpcs := c.shuffle()
	for i, healthy := range healthies {
		if healthy {
			return clients[i], rpcs[i]
		}
	}

	return nil, ""
}

func (c *defaultEthClient) execute(f func(client *ethclient.Client, rpc string) (any, error)) (any, error) {
	client, rpc := c.getHealthyClient()
	if client == nil {
		return nil, NewNoHealthyClientErr(c.chain)
	}

	ret, err := f(client, rpc)
	if err == nil {
		return ret, nil
	}

	return ret, err
}

func (c *defaultEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	num, err := c.execute(func(client *ethclient.Client, rpc string) (any, error) {
		return client.BlockNumber(ctx)
	})

	return num.(uint64), err
}

func (c *defaultEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error) {
	block, err := c.execute(func(client *ethclient.Client, rpc string) (any, error) {
		return client.BlockByNumber(ctx, number)
	})

	return block.(*ethtypes.Block), err
}

func (c *defaultEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	receipt, err := c.execute(func(client *ethclient.Client, rpc string) (any, error) {
		return client.TransactionReceipt(ctx, txHash)
	})

	return receipt.(*ethtypes.Receipt), err
}

func (c *defaultEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	gas, err := c.execute(func(client *ethclient.Client, rpc string) (any, error) {
		return client.SuggestGasPrice(ctx)
	})

	return gas.(*big.Int), err
}

func (c *defaultEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	nonce, err := c.execute(func(client *ethclient.Client, rpc string) (any, error) {
		return client.PendingNonceAt(ctx, account)
	})

	return nonce.(uint64), err
}

func (c *defaultEthClient) SendTransaction(ctx context.Context, tx *ethtypes.Transaction) error {
	_, err := c.execute(func(client *ethclient.Client, rpc string) (any, error) {
		err := client.SendTransaction(ctx, tx)
		return 0, err
	})

	return err
}

func (c *defaultEthClient) BalanceAt(ctx context.Context, from common.Address, block *big.Int) (*big.Int, error) {
	balance, err := c.execute(func(client *ethclient.Client, rpc string) (any, error) {
		balance, err := client.BalanceAt(ctx, from, block)
		if err == nil && balance != nil && balance.Cmp(big.NewInt(0)) == 0 {
			log.Verbosef("Balance is 0 for using URL %s", rpc)
		}

		return balance, err
	})

	return balance.(*big.Int), err
}
