package localstorage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
)

const (
	tempDirPerms       = 0755
	localMediaDirPerms = 0755
	localLogDirPerms   = 0755
)

var (
	objects        *LocalStorage
	notInRepoErr   = errors.New("not in a repository")
	TempDir        = filepath.Join(os.TempDir(), "git-lfs")
	checkedTempDir string
)

func Objects() *LocalStorage {
	return objects
}

func InitStorage() error {
	if len(config.LocalGitStorageDir) == 0 || len(config.LocalGitDir) == 0 {
		return notInRepoErr
	}

	TempDir = filepath.Join(config.LocalGitDir, "lfs", "tmp") // temp files per worktree
	objs, err := NewStorage(
		filepath.Join(config.LocalGitStorageDir, "lfs", "objects"),
		filepath.Join(TempDir, "objects"),
	)

	if err != nil {
		return errors.Wrap(err, "init LocalStorage")
	}

	objects = objs
	config.LocalLogDir = filepath.Join(objs.RootDir, "logs")
	if err := os.MkdirAll(config.LocalLogDir, localLogDirPerms); err != nil {
		return errors.Wrap(err, "create log dir")
	}

	return nil
}

func InitStorageOrFail() {
	if err := InitStorage(); err != nil {
		if err == notInRepoErr {
			return
		}

		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}

func ResolveDirs() {
	config.ResolveGitBasicDirs()
	InitStorageOrFail()
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
