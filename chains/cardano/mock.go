package cardano

import (
	"github.com/echovl/cardano-go"
	providertypes "github.com/sisu-network/deyes/chains/cardano/types"
	"github.com/sisu-network/deyes/types"
)

type MockCardanoClient struct {
	IsHealthyFunc   func() bool
	LatestBlockFunc func() (*providertypes.Block, error)
	GetBlockFunc    func(hashOrNumber string) (*providertypes.Block, error)
	BlockHeightFunc func() (int, error)
	NewTxsFunc      func(fromHeight int, gateway string) ([]*types.CardanoTransactionUtxo, error)
	SubmitTxFunc    func(tx *cardano.Tx) (*cardano.Hash32, error)
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
