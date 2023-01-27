package sushiswap

import (
	"github.com/sisu-network/deyes/types"
)

type MockSushiSwapManager struct {
	GetPriceFromSushiswapFunc func(tokenAddress string) (*types.TokenPrice, error)
}

func (m *MockSushiSwapManager) GetPriceFromSushiswap(tokenAddress string) (*types.TokenPrice, error) {
	if m.GetPriceFromSushiswapFunc != nil {
		return m.GetPriceFromSushiswapFunc(tokenAddress)
	}

	return nil, nil
}
