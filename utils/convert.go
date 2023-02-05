package utils

import (
	"math/big"
)

var (
	OneEtherInWei = int64(1_000_000_000_000_000_000)
	OneGweiInWei  = int64(1_000_000_000)
	SisuUnit      = int64(1_000_000_000_000_000_000)
)

func FloatToWei(value float64) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(value)

	bigval = bigval.Mul(bigval, new(big.Float).SetInt(big.NewInt(OneEtherInWei)))

	result := new(big.Int)
	bigval.Int(result)
	return result
}

// ToSisuPrice converts a price value with a specific decimal to Sisu unit (with 18 decimals).
func ToSisuPrice(price *big.Int, decimal int) *big.Int {
	ret := new(big.Int).Mul(price, big.NewInt(SisuUnit))
	ret = ret.Div(ret, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil))

	return ret
}
