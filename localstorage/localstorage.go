package localstorage

import (
	"regexp"
)

const (
	chanBufSize = 100
)

var (
	oidRE = regexp.MustCompile(`\A[[:alnum:]]{64}`)
)

// LocalStorage manages the locally stored LFS objects for a repository.
type LocalStorage struct {
	Root string
}

// Object represents a locally stored LFS object.
type Object struct {
	Oid  string
	Size int64
}

func New(dir string) *LocalStorage {
	return &LocalStorage{dir}
}
