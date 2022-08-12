package utils

import "math/big"

var (
	ONE_ETHER_IN_WEI = big.NewInt(1000000000000000000)
)

func FloatToWei(value float64) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(value)

	bigval = bigval.Mul(bigval, new(big.Float).SetInt(ONE_ETHER_IN_WEI))

	result := new(big.Int)
	bigval.Int(result)
	return result
}

func WeiToFloat(value *big.Int) float64 {
	f := new(big.Float).Quo(new(big.Float).SetInt(value), new(big.Float).SetInt(ONE_ETHER_IN_WEI))
	ret, _ := f.Float64()

	return ret
}
