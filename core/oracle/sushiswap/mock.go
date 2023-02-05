package sushiswap

import (
	"math/big"
)

type MockSushiSwapManager struct {
	GetPriceFromSushiswapFunc func(ttokenAddress1 string, tokenAddress2 string, amount *big.Int) (
		*big.Int, error)
}

func (m *MockSushiSwapManager) GetPriceFromSushiswap(tokenAddress1 string, tokenAddress2 string,
	amount *big.Int) (*big.Int, error) {
	if m.GetPriceFromSushiswapFunc != nil {
		return m.GetPriceFromSushiswapFunc(tokenAddress1, tokenAddress2, amount)
	}

	return nil, nil
}
