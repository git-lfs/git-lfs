// +build !linux !cgo
// +build !darwin !cgo
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

// This is almost identical to os.rename but doesn't replace newname if it already exists
func RenameNoReplace(oldname, newname string) error {
	return false, errors.New("unsupported platform")
}
