// +build !windows

package tools

import "syscall"

func GetMaxFileDescriptors() (uint64, error) {
	rlim := syscall.Rlimit{}
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim); err != nil {
		return 0, err
	}
	return rlim.Cur, nil
}
