package oracle

import (
	"math/big"

	"github.com/sisu-network/deyes/types"
)

var (
	DEFAULT_PRICES = map[string]*big.Float{
		"ETH":  big.NewFloat(1000),
		"DAI":  big.NewFloat(1),
		"SISU": big.NewFloat(0.02),
	}
)

func getDefaultTokenPriceList() []*types.TokenPrice {
	prices := make([]*types.TokenPrice, 0)

	for token, price := range DEFAULT_PRICES {
		tokenPrice := &types.TokenPrice{
			Id:       token,
			PublicId: token,
			Price:    price,
		}

		prices = append(prices, tokenPrice)
	}

	return prices
}
