package eth

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/lib/log"
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
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
}

type defaultEthClient struct {
	chain string

	clients []*ethclient.Client
	healthy []bool
	rpcs    []string

	lock *sync.RWMutex
}

func NewEthClients(initialRpcs []string, chain string) EthClient {
	clients := make([]*ethclient.Client, 0)
	rpcs := make([]string, 0)
	healthy := make([]bool, 0)

	for _, rpc := range initialRpcs {
		client, err := ethclient.Dial(rpc)
		if err == nil {
			clients = append(clients, client)
			rpcs = append(rpcs, rpc)
			healthy = append(healthy, true)
			log.Info("Adding eth client at rpc: ", rpc)
		}
	}

	return &defaultEthClient{
		chain:   chain,
		clients: clients,
		rpcs:    rpcs,
		healthy: healthy,
		lock:    &sync.RWMutex{},
	}
}

func (c *defaultEthClient) shuffle() ([]*ethclient.Client, []bool, []string) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	n := len(c.clients)

	clients := make([]*ethclient.Client, n)
	healthy := make([]bool, n)
	rpcs := make([]string, n)

	copy(clients, c.clients)
	copy(healthy, c.healthy)
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

func (c *defaultEthClient) getHealthyClient() (*ethclient.Client, int) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// Shuffle rpcs so that we will use different healthy rpc
	clients, healthies, _ := c.shuffle()
	for i, healthy := range healthies {
		if healthy {
			return clients[i], i
		}
	}

	return nil, -1
}

func (c *defaultEthClient) execute(f func(client *ethclient.Client) (any, error)) (any, error) {
	client, index := c.getHealthyClient()
	if client == nil {
		return nil, NewNoHealthyClientErr(c.chain)
	}

	ret, err := f(client)
	if err == nil {
		return ret, nil
	}

	if err != ethereum.NotFound {
		c.lock.Lock()
		fmt.Println("DDDDD Setting rpc to be unhealthy: ", c.rpcs[index], " err = ", err)
		// This client is not healthy anymore. We need to update the healthiness status.
		c.healthy[index] = false
		c.lock.Unlock()
	}

	return ret, err
}

func (c *defaultEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	num, err := c.execute(func(client *ethclient.Client) (any, error) {
		return client.BlockNumber(ctx)
	})

	return num.(uint64), err
}

func (c *defaultEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error) {
	block, err := c.execute(func(client *ethclient.Client) (any, error) {
		return client.BlockByNumber(ctx, number)
	})

	return block.(*ethtypes.Block), err
}

func (c *defaultEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	receipt, err := c.execute(func(client *ethclient.Client) (any, error) {
		return client.TransactionReceipt(ctx, txHash)
	})

	return receipt.(*ethtypes.Receipt), err
}

func (c *defaultEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	gas, err := c.execute(func(client *ethclient.Client) (any, error) {
		return client.SuggestGasPrice(ctx)
	})

	return gas.(*big.Int), err
}

func (c *defaultEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	nonce, err := c.execute(func(client *ethclient.Client) (any, error) {
		return client.PendingNonceAt(ctx, account)
	})

	return nonce.(uint64), err
}
