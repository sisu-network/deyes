package core

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/database"
)

type EthClient struct {
	rpcEndpoint string
	client      *ethclient.Client
	height      int64
	blockTime   int
	db          *database.Database
}

func NewClient(db *database.Database, rpcEndpoint string, blockTime int) *EthClient {
	return &EthClient{
		db:          db,
		rpcEndpoint: rpcEndpoint,
		blockTime:   blockTime,
	}
}

func (c *EthClient) init() {
	var err error
	c.client, err = ethclient.Dial(c.rpcEndpoint)
	if err != nil {
		panic(err)
	}

	// TODO: Load last blockheight
}

func (c *EthClient) Start() {
	go func() {
		c.init()
		// Get the blockheight
		block, err := c.getBlock(c.height)
		if err == ethereum.NotFound {
			// Ping block for every second.
		}

		// TODO: Save this block into a channel or a storage
		c.processBlock(block)

		c.height++
		time.Sleep(time.Millisecond * time.Duration(c.blockTime))
	}()
}

func (c *EthClient) getBlock(height int64) (*etypes.Block, error) {
	return c.client.BlockByNumber(context.Background(), big.NewInt(height))
}

func (c *EthClient) processBlock(block *etypes.Block) error {
	txBytes := make([][]byte, 0)

	for _, tx := range block.Transactions() {
		bytes, err := tx.MarshalJSON()
		if err != nil {
			return err
		}

		txBytes = append(txBytes, bytes)
	}

	return nil
}
