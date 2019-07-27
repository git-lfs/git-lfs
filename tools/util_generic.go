// +build !linux !cgo
// +build !darwin

package tools

import (
	"io"
)

func CloneFile(writer io.Writer, reader io.Reader) (bool, error) {
	return false, nil
}

func CloneFileByPath(_, _ string) (bool, error) {
	return false, nil
}
