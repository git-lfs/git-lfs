// +build !linux
// +build !darwin
// +build !windows

package tools

import (
	"io"

	"github.com/git-lfs/git-lfs/errors"
)

func CheckCloneFileSupported(dir string) (supported bool, err error) {
	return false, errors.New("unsupported platform")
}

func CloneFile(writer io.Writer, reader io.Reader) (bool, error) {
	return false, nil
}

func CloneFileByPath(_, _ string) (bool, error) {
	return false, nil
}
