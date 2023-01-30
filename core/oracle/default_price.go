package oracle

import (
	"math/big"

	"github.com/sisu-network/deyes/utils"
)

var (
	TestTokenPrices = map[string]*big.Int{
		"SISU":            big.NewInt(utils.OneEtherInWei * 4),
		"NATIVE_GANACHE1": big.NewInt(utils.OneEtherInWei * 2),
		"NATIVE_GANACHE2": big.NewInt(utils.OneEtherInWei * 3),
		"TIGER":           big.NewInt(utils.OneEtherInWei * 2),
		"KANGAROO":        big.NewInt(utils.OneEtherInWei * 2),
		"MOUSE":           big.NewInt(utils.OneEtherInWei * 2),
		"MONKEY":          big.NewInt(utils.OneEtherInWei * 2),
		"BUNNY":           big.NewInt(utils.OneEtherInWei * 2),
	}
)
