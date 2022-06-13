package core

import "github.com/sisu-network/deyes/types"

type MockClient struct {
	TryDialFunc              func()
	PingFunc                 func(source string) error
	BroadcastTxsFunc         func(txs *types.Txs) error
	PostDeploymentResultFunc func(result *types.DispatchedTxResult) error
	UpdateTokenPricesFunc    func(prices []*types.TokenPrice) error
}

func (c *MockClient) TryDial() {
	if c.TryDialFunc != nil {
		c.TryDialFunc()
	}
}

func (c *MockClient) Ping(source string) error {
	if c.PingFunc != nil {
		return c.PingFunc(source)
	}

	return nil
}

func (c *MockClient) BroadcastTxs(txs *types.Txs) error {
	if c.BroadcastTxsFunc != nil {
		return c.BroadcastTxsFunc(txs)
	}

	return nil
}

func (c *MockClient) PostDeploymentResult(result *types.DispatchedTxResult) error {
	if c.PostDeploymentResultFunc != nil {
		return c.PostDeploymentResultFunc(result)
	}

	return nil
}

func (c *MockClient) UpdateTokenPrices(prices []*types.TokenPrice) error {
	if c.UpdateTokenPricesFunc != nil {
		return c.UpdateTokenPricesFunc(prices)
	}

	return nil
}
