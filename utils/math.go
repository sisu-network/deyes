package utils

func AbsInt64(a int64) int64 {
	if a >= 0 {
		return a
	}

	return -a
}
