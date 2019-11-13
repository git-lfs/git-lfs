// +build !windows

package tools

import "os"

func RobustRename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func RobustOpen(name string) (*os.File, error) {
	return os.Open(name)
}
