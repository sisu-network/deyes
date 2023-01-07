package lisk

import (
	"github.com/sisu-network/deyes/chains/lisk/types"
)

type MockLiskClient struct {
	StartFunc              func()
	BlockNumberFunc        func() (uint64, error)
	BlockByHeightFunc      func(height uint64) (*types.Block, error)
	TransactionByBlockFunc func(block string) ([]*types.Transaction, error)
	CreateTransactionFunc  func(txHash string) (string, error)
}

func (c *MockLiskClient) Start() {
	if c.StartFunc != nil {
		c.StartFunc()
	}
}

func (c *MockLiskClient) BlockNumber() (uint64, error) {
	if c.BlockNumberFunc != nil {
		return c.BlockNumberFunc()
	}

	return 0, nil
}

func (c *MockLiskClient) BlockByHeight(height uint64) (*types.Block, error) {
	if c.BlockByHeightFunc != nil {
		return c.BlockByHeightFunc(height)
	}

	return nil, nil
}

func (c *MockLiskClient) CreateTransaction(txHash string) (string, error) {
	if c.CreateTransactionFunc != nil {
		return c.CreateTransactionFunc(txHash)
	}

	return "", nil
}

func (c *MockLiskClient) TransactionByBlock(block string) ([]*types.Transaction, error) {
	if c.TransactionByBlockFunc != nil {
		return c.TransactionByBlockFunc(block)
	}

	return nil, nil
}
