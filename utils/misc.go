package utils

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
