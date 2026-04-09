//go:build !windows
// +build !windows

package tools

import "os"

func RobustRename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func RobustOpen(name string) (*os.File, error) {
	return os.Open(name)
}

func RobustRemove(path string) error {
	return os.Remove(path)
}
