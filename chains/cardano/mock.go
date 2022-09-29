package cardano

import (
	"context"

	"github.com/echovl/cardano-go"
	providertypes "github.com/sisu-network/deyes/chains/cardano/types"
	"github.com/sisu-network/deyes/types"
)

type MockCardanoClient struct {
	IsHealthyFunc      func() bool
	LatestBlockFunc    func() (*providertypes.Block, error)
	GetBlockFunc       func(hashOrNumber string) (*providertypes.Block, error)
	BlockHeightFunc    func() (int, error)
	NewTxsFunc         func(fromHeight int, gateway string) ([]*types.CardanoTransactionUtxo, error)
	SubmitTxFunc       func(tx *cardano.Tx) (*cardano.Hash32, error)
	ProtocolParamsFunc func() (*cardano.ProtocolParams, error)
	AddressUTXOsFunc   func(ctx context.Context, address string, query providertypes.APIQueryParams) ([]cardano.UTxO, error)
	BalanceFunc        func(address string, maxBlock int64) (*cardano.Value, error)
	TipFunc            func(blockHeight uint64) (*cardano.NodeTip, error)
}

func (c *MockCardanoClient) IsHealthy() bool {
	if c.IsHealthyFunc != nil {
		return c.IsHealthyFunc()
	}

	return false
}

func (c *MockCardanoClient) LatestBlock() (*providertypes.Block, error) {
	if c.LatestBlockFunc != nil {
		return c.LatestBlockFunc()
	}

	return nil, nil
}

func (c *MockCardanoClient) BlockHeight() (int, error) {
	if c.BlockHeightFunc != nil {
		return c.BlockHeightFunc()
	}

	return 0, nil
}

func (c *MockCardanoClient) NewTxs(fromHeight int, gateway string) ([]*types.CardanoTransactionUtxo, error) {
	if c.NewTxsFunc != nil {
		return c.NewTxsFunc(fromHeight, gateway)
	}

	return nil, nil
}

func (c *MockCardanoClient) SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error) {
	if c.SubmitTxFunc != nil {
		return c.SubmitTxFunc(tx)
	}

	return nil, nil
}

func (c *MockCardanoClient) GetBlock(hashOrNumber string) (*providertypes.Block, error) {
	if c.GetBlockFunc != nil {
		return c.GetBlockFunc(hashOrNumber)
	}

	return nil, nil
}

func (c *MockCardanoClient) ProtocolParams() (*cardano.ProtocolParams, error) {
	if c.ProtocolParamsFunc != nil {
		return c.ProtocolParamsFunc()
	}

	return nil, nil
}

func (c *MockCardanoClient) AddressUTXOs(ctx context.Context, address string, query providertypes.APIQueryParams) ([]cardano.UTxO, error) {
	if c.AddressUTXOsFunc != nil {
		return c.AddressUTXOsFunc(ctx, address, query)
	}

	return nil, nil
}

func (c *MockCardanoClient) Balance(address string, maxBlock int64) (*cardano.Value, error) {
	if c.BalanceFunc != nil {
		return c.BalanceFunc(address, maxBlock)
	}

	return nil, nil
}

func (c *MockCardanoClient) Tip(blockHeight uint64) (*cardano.NodeTip, error) {
	if c.TipFunc != nil {
		return c.TipFunc(blockHeight)
	}

	return nil, nil
}
