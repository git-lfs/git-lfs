// +build linux,cgo

package tools

/*
#include <sys/ioctl.h>

#undef BTRFS_IOCTL_MAGIC
#define BTRFS_IOCTL_MAGIC 0x94
#undef BTRFS_IOC_CLONE
#define BTRFS_IOC_CLONE _IOW (BTRFS_IOCTL_MAGIC, 9, int)
*/
import "C"

import (
	"io"
	"io/ioutil"
	"os"
	"syscall"
)

const (
	BtrfsIocClone = C.BTRFS_IOC_CLONE
)

// CheckCloneFileSupported runs explicit test of clone file on supplied directory.
// This function creates some (src and dst) file in the directory and remove after test finished.
//
// If check failed (e.g. directory is read-only), returns err.
func CheckCloneFileSupported(dir string) (supported bool, err error) {
	src, err := ioutil.TempFile(dir, "src")
	if err != nil {
		return false, err
	}
	defer os.Remove(src.Name())

	dst, err := ioutil.TempFile(dir, "dst")
	if err != nil {
		return false, err
	}
	defer os.Remove(dst.Name())

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
		if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fdst.Fd(), BtrfsIocClone, fsrc.Fd()); err != 0 {
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
	dstFile, err := os.Create(dst) //truncating, it if it already exists.
	if err != nil {
		return false, err
	}

	return CloneFile(dstFile, srcFile)
}
