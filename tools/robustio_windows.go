//go:build windows
// +build windows

package tools

import (
	"errors"
	"os"
	"time"

	"github.com/avast/retry-go"
	"golang.org/x/sys/windows"
)

// isEphemeralError returns true if err may be resolved by waiting.
func isEphemeralError(err error) bool {
	return errors.Is(err, windows.ERROR_SHARING_VIOLATION)
}

// isFileInUseError returns true if remove err may be resolved by waiting.
// See https://github.com/git/git/blob/7f19e4e1b6a3ad259e2ed66033e01e03b8b74c5e/compat/mingw.c#L162-L171.
func isFileInUseError(err error) bool {
	return errors.Is(err, windows.ERROR_SHARING_VIOLATION) ||
		errors.Is(err, windows.ERROR_ACCESS_DENIED)
}

// robustRemoveRetryDelaysMs uses the same delay pattern as Git on Windows.
// See https://github.com/git/git/blob/7f19e4e1b6a3ad259e2ed66033e01e03b8b74c5e/compat/mingw.c#L243.
// Unlike Git, we do not prompt the user with "Unlink of file '%s' failed.
// Should I try again? (y/n)" when retries are exhausted, nor do we support
// the GIT_ASK_YESNO environment variable.
var robustRemoveRetryDelaysMs = [5]int{0, 1, 10, 20, 40}

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

func RobustRemove(path string) error {
	return retry.Do(
		func() error {
			return os.Remove(path)
		},
		retry.RetryIf(isFileInUseError),
		retry.DelayType(func(n uint, config *retry.Config) time.Duration {
			return time.Duration(robustRemoveRetryDelaysMs[n]) * time.Millisecond
		}),
		retry.Attempts(uint(len(robustRemoveRetryDelaysMs)+1)),
		retry.LastErrorOnly(true),
	)
}
