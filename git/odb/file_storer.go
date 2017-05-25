package odb

import (
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// fileStorer implements the storer interface by writing to the .git/objects
// directory on disc.
type fileStorer struct {
	// root is the top level /objects directory's path on disc.
	root string
}

// NewFileStorer returns a new fileStorer instance with the given root.
func newFileStorer(root string) *fileStorer {
	return &fileStorer{
		root: root,
	}
}

// Open implements the storer.Open function, and returns a io.ReadWriteCloser
// for the given SHA. If the file does not exist, or if there was any other
// error in opening the file, an error will be returned.
//
// It is the caller's responsibility to close the given file "f" after its use
// is complete.
func (fs *fileStorer) Open(sha []byte) (f io.ReadWriteCloser, err error) {
	return fs.open(fs.path(sha), os.O_RDONLY)
}

// Store implements the storer.Store function and returns the number of bytes
// written, along with any error encountered in copying the given io.Reader, "r"
// into the object database on disk at a path given by "sha".
//
// If the file already exists, could not be created, or opened, an error will be
// returned.
func (fs *fileStorer) Store(sha []byte, r io.Reader) (n int64, err error) {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		return 0, err
	}

	n, err = io.Copy(tmp, r)
	if err = tmp.Close(); err != nil {
		return n, err
	}
	if err != nil {
		return n, err
	}

	path := fs.path(sha)
	dir := filepath.Dir(path)

	// Since .git/objects partitions objects based on the first two
	// characters of their ASCII-encoded SHA1 object ID, ensure that
	// the directory exists before copying a file into it.
	if err = os.MkdirAll(dir, 0755); err != nil {
		return n, err
	}

	if _, err := os.Stat(path); os.IsExist(err) {
		return n, errors.Errorf("git/odb: file storer cannot copy into file %q, which already exists", path)
	}

	if err = os.Rename(tmp.Name(), path); err != nil {
		return n, err
	}

	return n, nil
}

// open opens a given file.
func (fs *fileStorer) open(path string, flag int) (*os.File, error) {
	return os.OpenFile(path, flag, 0)
}

// path returns an absolute path on disk to the object given by the OID "sha".
func (fs *fileStorer) path(sha []byte) string {
	encoded := hex.EncodeToString(sha)

	return filepath.Join(fs.root, encoded[:2], encoded[2:])
}
