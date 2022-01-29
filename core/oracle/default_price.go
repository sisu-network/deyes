package oracle

import "github.com/sisu-network/deyes/types"

var (
	DEFAULT_PRICES = map[string]float32{
		"ETH":  1000,
		"DAI":  1,
		"SISU": 0.02,
	}
)

func getDefaultTokenPriceList() types.TokenPrices {
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
