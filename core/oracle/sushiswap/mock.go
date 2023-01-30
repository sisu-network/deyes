package sushiswap

import (
	"github.com/sisu-network/deyes/types"
)

type MockSushiSwapManager struct {
	GetPriceFromSushiswapFunc func(ttokenAddress1 string, tokenAddress2 string, tokenName string) (*types.TokenPrice, error)
}

func (m *MockSushiSwapManager) GetPriceFromSushiswap(tokenAddress1 string, tokenAddress2 string, tokenName string) (*types.TokenPrice, error) {
	if m.GetPriceFromSushiswapFunc != nil {
		return m.GetPriceFromSushiswapFunc(tokenAddress1, tokenAddress2, tokenName)
	}

	return nil, nil
}
