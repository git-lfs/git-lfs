// +build darwin,cgo

package tools

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/git-lfs/git-lfs/errors"
	"golang.org/x/sys/unix"
)

/*
#include <fcntl.h>
#include <errno.h>
#include <sys/clonefile.h>
*/
import "C"

var cloneFileSupported bool

func init() {
	cloneFileSupported = checkCloneFileSupported()
}

// checkCloneFileSupported return iff Mac OS version is greater or equal to 10.12.x Sierra.
//
// clonefile is supported since Mac OS X 10.12
// https://www.manpagez.com/man/2/clonefile/
//
// kern.osrelease mapping
// 17.x.x. macOS 10.13.x High Sierra.
// 16.x.x  macOS 10.12.x Sierra.
// 15.x.x  OS X  10.11.x El Capitan.
func checkCloneFileSupported() bool {
	bytes, err := unix.Sysctl("kern.osrelease")
	if err != nil {
		return false
	}

	versionString := strings.Split(string(bytes), ".") // major.minor.patch
	if len(versionString) < 2 {
		return false
	}

	major, err := strconv.Atoi(versionString[0])
	if err != nil {
		return false
	}

	return major >= 16
}

// CheckCloneFileSupported runs explicit test of clone file on supplied directory.
// This function creates some (src and dst) file in the directory and remove after test finished.
//
// If check failed (e.g. directory is read-only), returns err.
func CheckCloneFileSupported(dir string) (supported bool, err error) {
	if !cloneFileSupported {
		return false, errors.New("unsupported OS version. >= 10.12.x Sierra required")
	}

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

	return CloneFileByPath(dst.Name(), src.Name())
}

type CloneFileError struct {
	Unsupported bool
	errorString string
}

func (c *CloneFileError) Error() string {
	return c.errorString
}

func CloneFile(_ io.Writer, _ io.Reader) (bool, error) {
	return false, nil // Cloning from io.Writer(file descriptor) is not supported by Darwin.
}

func CloneFileByPath(dst, src string) (bool, error) {
	if !cloneFileSupported {
		return false, &CloneFileError{Unsupported: true, errorString: "clonefile is not supported"}
	}

	if FileExists(dst) {
		if err := os.Remove(dst); err != nil {
			return false, err // File should be not exists before create
		}
	}

	if err := cloneFileSyscall(dst, src); err != nil {
		return false, err
	}

	return true, nil
}

func cloneFileSyscall(dst, src string) *CloneFileError {
	srcCString, err := unix.BytePtrFromString(src)
	if err != nil {
		return &CloneFileError{errorString: err.Error()}
	}
	dstCString, err := unix.BytePtrFromString(dst)
	if err != nil {
		return &CloneFileError{errorString: err.Error()}
	}

	atFDCwd := C.AT_FDCWD // current directory.

	_, _, errNo := unix.Syscall6(
		unix.SYS_CLONEFILEAT,
		uintptr(atFDCwd),
		uintptr(unsafe.Pointer(srcCString)),
		uintptr(atFDCwd),
		uintptr(unsafe.Pointer(dstCString)),
		uintptr(C.CLONE_NOFOLLOW),
		0,
	)
	if errNo != 0 {
		return &CloneFileError{
			Unsupported: errNo == C.ENOTSUP,
			errorString: fmt.Sprintf("%s. from %v to %v", unix.ErrnoName(errNo), src, dst),
		}
	}

	return nil
}
