//go:build !windows
// +build !windows

package tools

func isCygwin() bool {
	return false
}
