package utils

import "math/big"

var (
	ONE_ETHER_IN_WEI = big.NewInt(1_000_000_000_000_000_000)
)

func FloatToWei(value float64) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(value)

	bigval = bigval.Mul(bigval, new(big.Float).SetInt(ONE_ETHER_IN_WEI))

	result := new(big.Int)
	bigval.Int(result)
	return result
}
