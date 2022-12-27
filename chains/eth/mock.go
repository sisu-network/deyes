package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type MockEthClient struct {
	StartFunc              func()
	BlockNumberFunc        func(ctx context.Context) (uint64, error)
	BlockByNumberFunc      func(ctx context.Context, number *big.Int) (*ethtypes.Block, error)
	TransactionReceiptFunc func(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error)
	SuggestGasPriceFunc    func(ctx context.Context) (*big.Int, error)
	PendingNonceAtFunc     func(ctx context.Context, account common.Address) (uint64, error)
	SendTransactionFunc    func(ctx context.Context, tx *ethtypes.Transaction) error
	BalanceAtFunc          func(ctx context.Context, from common.Address, block *big.Int) (*big.Int, error)
}

func (c *MockEthClient) Start() {
	if c.StartFunc != nil {
		c.StartFunc()
	}
}

func (c *MockEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	if c.BlockNumberFunc != nil {
		return c.BlockNumberFunc(ctx)
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

func (c *MockEthClient) SendTransaction(ctx context.Context, tx *ethtypes.Transaction) error {
	if c.SendTransactionFunc != nil {
		return c.SendTransactionFunc(ctx, tx)
	}

	return nil
}

func (c *MockEthClient) BalanceAt(ctx context.Context, from common.Address, block *big.Int) (*big.Int, error) {
	if c.BalanceAtFunc != nil {
		return c.BalanceAtFunc(ctx, from, block)
	}

	return nil, nil
}

//////

type mockTrieHasher struct{}

func (h *mockTrieHasher) Reset() {}

func (h *mockTrieHasher) Update([]byte, []byte) {}

func (h *mockTrieHasher) Hash() common.Hash {
	return [32]byte{}
}

//////

type mockRpcChecker struct {
	GetExtraRpcsFunc func(chainId int) ([]string, error)
}

func (m *mockRpcChecker) GetExtraRpcs(chainId int) ([]string, error) {
	if m.GetExtraRpcsFunc != nil {
		return m.GetExtraRpcsFunc(chainId)
	}

	return nil, nil
}
