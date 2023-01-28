package sushiswap

import (
	"github.com/sisu-network/deyes/types"
)

type MockSushiSwapManager struct {
	GetPriceFromSushiswapFunc func(tokenAddress string, tokenName string) (*types.TokenPrice, error)
}

func (m *MockSushiSwapManager) GetPriceFromSushiswap(tokenAddress string, tokenName string) (*types.TokenPrice, error) {
	if m.GetPriceFromSushiswapFunc != nil {
		return m.GetPriceFromSushiswapFunc(tokenAddress, tokenName)
	}

	return nil, nil
}
