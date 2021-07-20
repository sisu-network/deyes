package core

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/database"
)

type EthClient struct {
	chain       string
	rpcEndpoint string
	client      *ethclient.Client
	blockHeight int64
	blockTime   int
	db          *database.Database
}

func NewClient(db *database.Database, rpcEndpoint string, blockTime int, chain string) *EthClient {
	return &EthClient{
		db:          db,
		rpcEndpoint: rpcEndpoint,
		blockTime:   blockTime,
		chain:       chain,
	}
}

func (c *EthClient) init() {
	var err error
	c.client, err = ethclient.Dial(c.rpcEndpoint)
	if err != nil {
		panic(err)
	}

	blockHeight, err := c.db.LoadBlockHeight(c.chain)
	if err != nil {
		panic(err)
	}

	fmt.Println("blockHeight from db = ", blockHeight)

	c.blockHeight = blockHeight
}

func (c *EthClient) Start() {
	go func() {
		c.init()

		c.scanBlocks()
	}()
}

func (c *EthClient) scanBlocks() {
	for {
		fmt.Println("Getting block")
		// Get the blockheight
		block, err := c.getBlock(c.blockHeight)
		fmt.Println("err = ", err)

		switch err {
		case nil:
			fmt.Println("Height = ", block.Number())
		case ethereum.NotFound:
			// Ping block for every second.
		}

		// TODO: Save this block into a channel or a storage
		c.processBlock(block)

		c.blockHeight++
		time.Sleep(time.Duration(c.blockTime) * time.Millisecond)
	}
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
