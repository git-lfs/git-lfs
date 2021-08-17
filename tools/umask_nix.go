//go:build !windows
// +build !windows

package tools

import "syscall"

func doWithUmask(mask int, f func() error) error {
	mask = syscall.Umask(mask)
	defer syscall.Umask(mask)
	return f()
}
