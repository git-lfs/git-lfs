// +build !windows

package commands

// cleanRootPath is a no-op on every platform except Windows
func cleanRootPath(pattern string) string {
	return pattern
}

func osLineEnding() string {
	return "\n"
}
