package tools

// ClampInt returns the integer "n" bounded between "low" and "high".
func ClampInt(n, low, high int) int {
	return min(high, max(low, n))
}
