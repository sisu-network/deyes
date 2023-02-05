package utils

import (
	"fmt"
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

// UsdToSisuPrice converts a token price in USD format (e.g. "68.382") to an int Sisu token price
// with 18 decimals.
func UsdToSisuPrice(s string) (*big.Int, error) {
	bigVal, ok := new(big.Float).SetString(s)
	if !ok {
		return nil, fmt.Errorf("Invalid price %s", s)
	}

	coff := new(big.Float).SetInt64(SisuUnit)
	bigVal = bigVal.Mul(bigVal, coff)
	result := new(big.Int)
	bigVal.Int(result)

	return result, nil
}
