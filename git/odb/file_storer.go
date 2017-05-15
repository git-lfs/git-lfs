package odb

import (
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
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

// Create implements the Storere.Create function and returns an
// io.ReadWriteCloser for the given SHA, provided that an object of that SHA is
// not already indexed in the database.
//
// If the file already exists, could not be created, or opened, an error will be
// returned.
//
// It is the caller's responsibility to close the given file "f" after its use
// is complete.
func (fs *FileStorer) Create(sha []byte) (f io.ReadWriteCloser, err error) {
	path := fs.path(sha)
	dir := filepath.Dir(path)

	if err = os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return fs.open(path, os.O_RDWR|os.O_CREATE|os.O_EXCL)
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
