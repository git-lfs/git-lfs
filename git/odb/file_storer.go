package odb

import (
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// FileStorer implements the Storer interface by writing to the .git/objects
// directory on disc.
type FileStorer struct {
	// root is the top level /objects directory's path on disc.
	root string
}

var _ Storer = (*FileStorer)(nil)

// NewFileStorer returns a new FileStorer instance with the given root.
func NewFileStorer(root string) *FileStorer {
	return &FileStorer{
		root: root,
	}
}

// Open implements the Storer.Open function, and returns a io.ReadWriteCloser
// for the given SHA. If the file does not exist, or if there was any other
// error in opening the file, an error will be returned.
//
// It is the caller's responsibility to close the given file "f" after its use
// is complete.
func (fs *FileStorer) Open(sha []byte) (f io.ReadWriteCloser, err error) {
	return fs.open(fs.path(sha), os.O_RDONLY)
}

// Store implements the Storer.Store function and returns the number of bytes
// written, along with any error encountered in copying the given io.Reader, "r"
// into the object database on disk at a path given by "sha".
//
// If the file already exists, could not be created, or opened, an error will be
// returned.
func (fs *FileStorer) Store(sha []byte, r io.Reader) (n int64, err error) {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		return 0, err
	}

	n, err = io.Copy(tmp, r)
	if err != nil {
		return n, err
	}

	path := fs.path(sha)

	if _, err := os.Stat(path); os.IsExist(err) {
		return n, errors.Errorf("git/odb: file storer cannot copy into file %q, which already exists", path)
	}

	if err = os.Rename(tmp.Name(), path); err != nil {
		return n, err
	}

	if err = tmp.Close(); err != nil {
		return n, err
	}
	return n, nil
}

// open opens a given file.
func (fs *FileStorer) open(path string, flag int) (*os.File, error) {
	return os.OpenFile(path, flag, 0)
}

// path returns an absolute path on disk to the object given by the OID "sha".
func (fs *FileStorer) path(sha []byte) string {
	encoded := hex.EncodeToString(sha)

	return filepath.Join(fs.root, encoded[:2], encoded[2:])
}
