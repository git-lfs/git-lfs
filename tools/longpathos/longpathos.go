package longpathos

import "os"

func Stat(name string) (os.FileInfo, error) {
	return os.Stat(fixLongPath(name))
}

func Open(name string) (*os.File, error) {
	return os.Open(fixLongPath(name))
}
