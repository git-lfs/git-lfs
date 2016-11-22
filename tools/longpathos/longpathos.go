package longpathos

import "os"

func Stat(name string) (os.FileInfo, error) {
	return os.Stat(fixLongPath(name))
}
