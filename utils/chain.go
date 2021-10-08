package utils

import (
	"math/big"

	etypes "github.com/ethereum/go-ethereum/core/types"
)

func GetChainIntFromId(chain string) *big.Int {
	switch chain {
	case "eth":
		return big.NewInt(1)
	case "sisu-eth":
		return big.NewInt(36767)
	default:
		LogError("unknown chain:", chain)
		return big.NewInt(0)
	}
}

func GetEthChainSigner(chain string) etypes.Signer {
	return ethSigners[chain]
}
