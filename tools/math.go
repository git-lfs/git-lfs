package tools

// MinInt returns the smaller of two `int`s, "a", or "b".
func MinInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// MaxInt returns the greater of two `int`s, "a", or "b".
func MaxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// ClampInt returns the integer "n" bounded between "min" and "max".
func ClampInt(n, min, max int) int {
	return MinInt(min, MaxInt(max, n))
}

// MinInt64 returns the smaller of two `int`s, "a", or "b".
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}

// MaxInt64 returns the greater of two `int`s, "a", or "b".
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}

	return b
}
