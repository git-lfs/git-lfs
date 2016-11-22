package longpathos

import "os"

func Stat(name string) (os.FileInfo, error) {
	return os.Stat(fixLongPath(name))
}

func Open(name string) (*os.File, error) {
	return os.Open(fixLongPath(name))
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(fixLongPath(name), flag, perm)
}
