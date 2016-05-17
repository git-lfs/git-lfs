package localstorage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/config"
)

const (
	tempDirPerms       = 0755
	localMediaDirPerms = 0755
	localLogDirPerms   = 0755
)

var (
	objects        *LocalStorage
	TempDir        = filepath.Join(os.TempDir(), "git-lfs")
	checkedTempDir string
)

func Objects() *LocalStorage {
	return objects
}

func ResolveDirs() {

	config.ResolveGitBasicDirs()
	TempDir = filepath.Join(config.LocalGitDir, "lfs", "tmp") // temp files per worktree

	objs, err := NewStorage(
		filepath.Join(config.LocalGitStorageDir, "lfs", "objects"),
		filepath.Join(TempDir, "objects"),
	)

	if err != nil {
		panic(fmt.Sprintf("Error trying to init LocalStorage: %s", err))
	}

	objects = objs
	config.LocalLogDir = filepath.Join(objs.RootDir, "logs")
	if err := os.MkdirAll(config.LocalLogDir, localLogDirPerms); err != nil {
		panic(fmt.Errorf("Error trying to create log directory in '%s': %s", config.LocalLogDir, err))
	}
}

func TempFile(prefix string) (*os.File, error) {
	if checkedTempDir != TempDir {
		if err := os.MkdirAll(TempDir, tempDirPerms); err != nil {
			return nil, err
		}
		checkedTempDir = TempDir
	}

	return ioutil.TempFile(TempDir, prefix)
}

func ResetTempDir() error {
	checkedTempDir = ""
	return os.RemoveAll(TempDir)
}
