// Package localstorage handles LFS content stored locally
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package localstorage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	chanBufSize = 100
)

var (
	oidRE                = regexp.MustCompile(`\A[[:alnum:]]{64}`)
	dirPerms os.FileMode = 0666
)

// LocalStorage manages the locally stored LFS objects for a repository.
type LocalStorage struct {
	RootDir string
	TempDir string
}

// Object represents a locally stored LFS object.
type Object struct {
	Oid  string
	Size int64
}

func NewStorage(storageDir, tempDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(storageDir, dirPerms); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(tempDir, dirPerms); err != nil {
		return nil, err
	}

	return &LocalStorage{storageDir, tempDir}, nil
}

func (s *LocalStorage) ObjectPath(oid string) string {
	return filepath.Join(localObjectDir(s, oid), oid)
}

func (s *LocalStorage) BuildObjectPath(oid string) (string, error) {
	dir := localObjectDir(s, oid)
	if err := os.MkdirAll(dir, dirPerms); err != nil {
		return "", fmt.Errorf("Error trying to create local storage directory in %q: %s", dir, err)
	}

	return filepath.Join(dir, oid), nil
}

func localObjectDir(s *LocalStorage, oid string) string {
	return filepath.Join(s.RootDir, oid[0:2], oid[2:4])
}
