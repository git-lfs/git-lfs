//go:build windows
// +build windows

package tools

import (
	"errors"
	"os"

	"github.com/avast/retry-go"
	"golang.org/x/sys/windows"
)

// isEphemeralError returns true if err may be resolved by waiting.
func isEphemeralError(err error) bool {
	return errors.Is(err, windows.ERROR_SHARING_VIOLATION)
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
