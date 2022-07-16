package oracle

import (
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

var (
	DEFAULT_PRICES = map[string]float64{
		"ETH":  1000.0,
		"DAI":  1.0,
		"SISU": 0.02,
	}
)

func getDefaultTokenPriceList() []*types.TokenPrice {
	prices := make([]*types.TokenPrice, 0)

	for token, price := range DEFAULT_PRICES {
		tokenPrice := &types.TokenPrice{
			Id:       token,
			PublicId: token,
			Price:    utils.FloatToWei(price),
		}

		prices = append(prices, tokenPrice)
	}

	return prices
}
