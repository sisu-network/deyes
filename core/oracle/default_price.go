package oracle

import (
	"math/big"

	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

var (
	TestTokenPrices = map[string]*big.Int{
		"TIGER":    big.NewInt(utils.OneEtherInWei * 2),
		"KANGAROO": big.NewInt(utils.OneEtherInWei * 2),
		"MOUSE":    big.NewInt(utils.OneEtherInWei * 2),
		"MONKEY":   big.NewInt(utils.OneEtherInWei * 2),
		"BUNNY":    big.NewInt(utils.OneEtherInWei * 2),
	}
)

func getDefaultTokenPriceList() []*types.TokenPrice {
	prices := make([]*types.TokenPrice, 0)

	for token, price := range TestTokenPrices {
		tokenPrice := &types.TokenPrice{
			Id:       token,
			PublicId: token,
			Price:    price,
		}

		prices = append(prices, tokenPrice)
	}

	return prices
}
