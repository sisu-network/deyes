package uniswap

import (
	"math/big"
)

type MockNewUniwapManager struct {
	GetPriceFromUniswapFunc func(tokenAddress1 string, tokenAddress2 string) (*big.Int, error)
}

func (m *MockNewUniwapManager) GetPriceFromUniswap(tokenAddress1 string, tokenAddress2 string) (*big.Int, error) {
	if m.GetPriceFromUniswapFunc != nil {
		return m.GetPriceFromUniswapFunc(tokenAddress1, tokenAddress2)
	}

	return nil, nil
}
