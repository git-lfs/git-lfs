package longpathos

import (
	"os"
	"time"
)

func Stat(name string) (os.FileInfo, error) {
	return os.Stat(fixLongPath(name))
}

func Open(name string) (*os.File, error) {
	return os.Open(fixLongPath(name))
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(fixLongPath(name), flag, perm)
}

func Link(oldname, newname string) error {
	return os.Link(fixLongPath(oldname), fixLongPath(newname))
}

func Chdir(dir string) error {
	return os.Chdir(fixLongPath(dir))
}

func Chtimes(name string, atime, mtime time.Time) error {
	return os.Chtimes(fixLongPath(name), atime, mtime)
}
