package core

import (
	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	"github.com/sisu-network/deyes/types"
)

type MockCardanoClient struct {
	IsHealthyFunc   func() bool
	LatestBlockFunc func() *blockfrost.Block
	BlockHeightFunc func() (int, error)
	NewTxsFunc      func(fromHeight int, interestedAddrs map[string]bool) ([]*types.CardanoUtxo, error)
	SubmitTxFunc    func(tx *cardano.Tx) (*cardano.Hash32, error)
}

func (c *MockCardanoClient) IsHealthy() bool {
	if c.IsHealthyFunc != nil {
		return c.IsHealthyFunc()
	}

	return false
}

func (c *MockCardanoClient) LatestBlock() *blockfrost.Block {
	if c.LatestBlockFunc != nil {
		return c.LatestBlockFunc()
	}

	return nil
}

func (c *MockCardanoClient) BlockHeight() (int, error) {
	if c.BlockHeightFunc != nil {
		return c.BlockHeightFunc()
	}

	return 0, nil
}

func (c *MockCardanoClient) NewTxs(fromHeight int, interestedAddrs map[string]bool) ([]*types.CardanoUtxo, error) {
	if c.NewTxsFunc != nil {
		return c.NewTxsFunc(fromHeight, interestedAddrs)
	}

	return nil, nil
}

func (c *MockCardanoClient) SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error) {
	if c.SubmitTxFunc != nil {
		return c.SubmitTxFunc(tx)
	}

	return nil, nil
}
