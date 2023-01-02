package utils

import "math/big"

var (
	OneEtherInWei = int64(1_000_000_000_000_000_000)
	OneGweiInWei  = int64(1_000_000_000)
)

func FloatToWei(value float64) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(value)

	bigval = bigval.Mul(bigval, new(big.Float).SetInt(big.NewInt(OneEtherInWei)))

	result := new(big.Int)
	bigval.Int(result)
	return result
}
