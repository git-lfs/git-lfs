//go:build !windows
// +build !windows

package config

import "syscall"

func umask() int {
	// umask(2), which this function wraps, also sets the umask, so set it
	// back.
	umask := syscall.Umask(022)
	syscall.Umask(umask)
	return umask
}
