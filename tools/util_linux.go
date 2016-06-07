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
	"os"
	"syscall"
)

const (
	BtrfsIocClone = C.BTRFS_IOC_CLONE
)

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
