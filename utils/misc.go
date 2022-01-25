package utils

import (
	"math/big"
	"sort"
)

// Min returns the smaller of x or y.
func MinInt(x, y int64) int64 {
	if x > y {
		return y
	}
	return x
}

// Max returns the larger of x or y.
func MaxInt(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func GetMedianBigInt(a []*big.Int) *big.Int {
	if len(a) == 0 {
		return big.NewInt(0)
	}

	sort.SliceStable(a, func(i, j int) bool {
		return a[i].Cmp(a[j]) < 0
	})
	return a[len(a)/2]
}
