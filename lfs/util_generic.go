// +build !linux !cgo

package lfs

import (
	"io"
)

func CloneFile(writer io.Writer, reader io.Reader) (bool, error) {
	return false, nil
}
