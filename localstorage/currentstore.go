package localstorage

import (
	"fmt"
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

func InitStorage(cfg *config.Configuration) error {
	if len(cfg.LocalGitStorageDir()) == 0 || len(cfg.LocalGitDir()) == 0 {
		return notInRepoErr
	}

	storCfg := NewConfig(cfg)
	TempDir = filepath.Join(storCfg.LfsStorageDir, "tmp") // temp files per worktree
	objs, err := NewStorage(
		filepath.Join(storCfg.LfsStorageDir, "objects"),
		filepath.Join(TempDir, "objects"),
	)

	if err != nil {
		return errors.Wrap(err, "init LocalStorage")
	}

	objects = objs
	if err := os.MkdirAll(cfg.LocalLogDir(), localLogDirPerms); err != nil {
		return errors.Wrap(err, "create log dir")
	}

	return nil
}

func InitStorageOrFail(cfg *config.Configuration) {
	if err := InitStorage(cfg); err != nil {
		if err == notInRepoErr {
			return
		}

		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}

func ResolveDirs(cfg *config.Configuration) {
	InitStorageOrFail(cfg)
}

func ResetTempDir() error {
	checkedTempDir = ""
	return os.RemoveAll(TempDir)
}
