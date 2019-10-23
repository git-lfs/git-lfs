// +build linux,cgo

package tools

/*
#include <sys/ioctl.h>

#undef FICLONE
#define FICLONE		_IOW(0x94, 9, int)
// copy from https://github.com/torvalds/linux/blob/v5.2/include/uapi/linux/fs.h#L195 for older header files.
// This is equal to the older BTRFS_IOC_CLONE value.
*/
import "C"

import (
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/sys/unix"
)

const (
	ioctlFiClone = C.FICLONE
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
		if _, _, err := unix.Syscall(unix.SYS_IOCTL, fdst.Fd(), ioctlFiClone, fsrc.Fd()); err != 0 {
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

// This is almost identical to os.rename but doesn't replace newname if it already exists
func RenameNoReplace(oldname, newname string) error {
	fi, err := os.Lstat(newname)
	if err == nil && fi.IsDir() {
		if _, err := os.Lstat(oldname); err != nil {
			if pe, ok := err.(*os.PathError); ok {
				err = pe.Err
			}
			return &os.LinkError{Op: "rename", Old: oldname, New: newname, Err: err}
		}
		return &os.LinkError{Op: "rename", Old: oldname, New: newname, Err: unix.EEXIST}
	}
	err = unix.Renameat2(unix.AT_FDCWD, oldname, unix.AT_FDCWD, newname, unix.RENAME_NOREPLACE)
	if err != nil {
		return &os.LinkError{Op: "rename", Old: oldname, New: newname, Err: err}
	}
	return nil
}
