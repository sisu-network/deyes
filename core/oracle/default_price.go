package oracle

import (
	"math/big"

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
