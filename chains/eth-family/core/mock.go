package core

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type MockEthClient struct {
	BlockNumberFunc        func(ctx context.Context) (uint64, error)
	BlockByNumberFunc      func(ctx context.Context, number *big.Int) (*ethtypes.Block, error)
	TransactionReceiptFunc func(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error)
	SuggestGasPriceFunc    func(ctx context.Context) (*big.Int, error)
	PendingNonceAtFunc     func(ctx context.Context, account common.Address) (uint64, error)
}

func (c *MockEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	if c.BlockNumberFunc != nil {
		return c.BlockNumber(ctx)
	}
	return 0, nil
}

func (c *MockEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error) {
	if c.BlockByNumberFunc != nil {
		return c.BlockByNumberFunc(ctx, number)
	}

	return nil, nil
}

func (c *MockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	if c.TransactionReceiptFunc != nil {
		return c.TransactionReceiptFunc(ctx, txHash)
	}

	return nil, nil
}

func (c *MockEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if c.SuggestGasPriceFunc != nil {
		return c.SuggestGasPriceFunc(ctx)
	}
	return nil, nil
}

func (c *MockEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	if c.PendingNonceAtFunc != nil {
		return c.PendingNonceAtFunc(ctx, account)
	}

	return 0, nil
}
