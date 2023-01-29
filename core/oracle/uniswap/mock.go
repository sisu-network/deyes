package uniswap

import (
	"github.com/sisu-network/deyes/types"
)

type MockNewUniwapManager struct {
	GetPriceFromUniswapFunc func(tokenAddress1 string, tokenAddress2 string, tokenName string) (*types.TokenPrice, error)
}

func (m *MockNewUniwapManager) GetPriceFromUniswap(tokenAddress1 string, tokenAddress2 string, tokenName string) (*types.TokenPrice, error) {
	if m.GetPriceFromUniswapFunc != nil {
		return m.GetPriceFromUniswapFunc(tokenAddress1, tokenAddress2, tokenName)
	}

	return nil, nil
}
