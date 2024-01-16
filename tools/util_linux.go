//go:build linux
// +build linux

package tools

import (
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// CheckCloneFileSupported runs explicit test of clone file on supplied directory.
// This function creates some (src and dst) file in the directory and remove after test finished.
//
// If check failed (e.g. directory is read-only), returns err.
func CheckCloneFileSupported(dir string) (supported bool, err error) {
	src, err := os.CreateTemp(dir, "src")
	if err != nil {
		return false, err
	}
	defer func() {
		src.Close()
		os.Remove(src.Name())
	}()

	dst, err := os.CreateTemp(dir, "dst")
	if err != nil {
		return false, err
	}
	defer func() {
		dst.Close()
		os.Remove(dst.Name())
	}()

	if ok, err := CloneFile(dst, src); err != nil {
		return false, err
	} else {
		return ok, nil
	}
}

func CloneFile(writer io.Writer, reader io.Reader) (bool, error) {
	fdst, fdstFound := writer.(*os.File)
	fsrc, fsrcFound := reader.(*os.File)
	if fdstFound && fsrcFound {
		if err := unix.IoctlFileClone(int(fdst.Fd()), int(fsrc.Fd())); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func CloneFileByPath(dst, src string) (bool, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst) //truncating, it if it already exists.
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	return CloneFile(dstFile, srcFile)
}
