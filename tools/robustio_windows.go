// +build windows

package tools

import (
	"os"
	"syscall"

	"github.com/avast/retry-go"
)

const (
	// This is private in [go]/src/internal/syscall/windows/syscall_windows.go :(
	ERROR_SHARING_VIOLATION syscall.Errno = 32
)

func underlyingError(err error) error {
	switch err := err.(type) {
	case *os.PathError:
		return err.Err
	case *os.LinkError:
		return err.Err
	case *os.SyscallError:
		return err.Err
	}
	return err
}

// isEphemeralError returns true if err may be resolved by waiting.
func isEphemeralError(err error) bool {
	// TODO: Use this instead for Go >= 1.13
	// return errors.Is(err, ERROR_SHARING_VIOLATION)

	err = underlyingError(err)
	return err == ERROR_SHARING_VIOLATION
}

func RobustRename(oldpath, newpath string) error {
	return retry.Do(
		func() error {
			return os.Rename(oldpath, newpath)
		},
		retry.RetryIf(isEphemeralError),
		retry.LastErrorOnly(true),
	)
}

func RobustOpen(name string) (*os.File, error) {
	var result *os.File
	return result, retry.Do(
		func() error {
			f, err := os.Open(name)
			result = f
			return err
		},
		retry.RetryIf(isEphemeralError),
		retry.LastErrorOnly(true),
	)
}
