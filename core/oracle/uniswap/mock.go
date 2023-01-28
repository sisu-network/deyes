package uniswap

import (
	"github.com/sisu-network/deyes/types"
)

type MockNewUniwapManager struct {
	GetPriceFromUniswapFunc func(tokenAddress string, tokenName string) (*types.TokenPrice, error)
}

func (m *MockNewUniwapManager) GetPriceFromUniswap(tokenAddress string, tokenName string) (*types.TokenPrice, error) {
	if m.GetPriceFromUniswapFunc != nil {
		return m.GetPriceFromUniswapFunc(tokenAddress, tokenName)
	}

	return nil, nil
}
