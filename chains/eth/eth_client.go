package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/lib/log"
)

// EthClient A wrapper around eth.client so that we can mock in watcher tests.
type EthClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
}

type defaultEthClient struct {
	client *ethclient.Client
}

func NewEthClients(rpcs []string) []EthClient {
	clients := make([]EthClient, 0)

	for _, rpc := range rpcs {
		client, err := dial(rpc)
		if err == nil {
			clients = append(clients, client)
			log.Info("Adding eth client at rpc: ", rpc)
		}
	}

	return clients
}

func dial(rawurl string) (EthClient, error) {
	client, err := ethclient.Dial(rawurl)
	if err != nil {
		return nil, err
	}

	return &defaultEthClient{
		client: client,
	}, nil
}

func (c *defaultEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	return c.client.BlockNumber(ctx)
}

func (c *defaultEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error) {
	return c.client.BlockByNumber(ctx, number)
}

func (c *defaultEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	return c.client.TransactionReceipt(ctx, txHash)
}

func (c *defaultEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return c.client.SuggestGasPrice(ctx)
}

func (c *defaultEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return c.client.PendingNonceAt(ctx, account)
}
