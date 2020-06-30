package gitobj

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/git-lfs/gitobj/v2/errors"
)

// fileStorer implements the storer interface by writing to the .git/objects
// directory on disc.
type fileStorer struct {
	// root is the top level /objects directory's path on disc.
	root string

	// temp directory, defaults to os.TempDir
	tmp string
}

// NewFileStorer returns a new fileStorer instance with the given root.
func newFileStorer(root, tmp string) *fileStorer {
	return &fileStorer{
		root: root,
		tmp:  tmp,
	}
}

// Open implements the storer.Open function, and returns a io.ReadCloser
// for the given SHA. If the file does not exist, or if there was any other
// error in opening the file, an error will be returned.
//
// It is the caller's responsibility to close the given file "f" after its use
// is complete.
func (fs *fileStorer) Open(sha []byte) (f io.ReadCloser, err error) {
	f, err = fs.open(fs.path(sha), os.O_RDONLY)
	if os.IsNotExist(err) {
		return nil, errors.NoSuchObject(sha)
	}
	return f, err
}

// Store implements the storer.Store function and returns the number of bytes
// written, along with any error encountered in copying the given io.Reader, "r"
// into the object database on disk at a path given by "sha".
//
// If the file could not be created, or opened, an error will be returned.
func (fs *fileStorer) Store(sha []byte, r io.Reader) (n int64, err error) {
	path := fs.path(sha)
	dir := filepath.Dir(path)

	if stat, err := os.Stat(path); stat != nil || os.IsExist(err) {
		// If the file already exists, there is no work left for us to
		// do, since the object already exists (or there is a SHA1
		// collision).
		_, err = io.Copy(ioutil.Discard, r)
		if err != nil {
			return 0, fmt.Errorf("discard pre-existing object data: %s", err)
		}

		return 0, nil
	}

	tmp, err := ioutil.TempFile(fs.tmp, "")
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

	// Since .git/objects partitions objects based on the first two
	// characters of their ASCII-encoded SHA1 object ID, ensure that
	// the directory exists before copying a file into it.
	if err = os.MkdirAll(dir, 0755); err != nil {
		return n, err
	}

	if err = os.Rename(tmp.Name(), path); err != nil {
		return n, err
	}

	return n, nil
}

// Root gives the absolute (fully-qualified) path to the file storer on disk.
func (fs *fileStorer) Root() string {
	return fs.root
}

// Close closes the file storer.
func (fs *fileStorer) Close() error {
	return nil
}

// IsCompressed returns true, because the file storer returns compressed data.
func (fs *fileStorer) IsCompressed() bool {
	return true
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
