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
