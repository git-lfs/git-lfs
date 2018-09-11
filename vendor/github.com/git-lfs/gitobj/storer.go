package gitobj

import "io"

// storer implements a storage engine for reading, writing, and creating
// io.ReadWriters that can store information about loose objects
type storer interface {
	// Open returns a handle on an existing object keyed by the given SHA.
	// It returns an error if that file does not already exist.
	Open(sha []byte) (f io.ReadCloser, err error)

	// Store copies the data given in "r" to the unique object path given by
	// "sha". It returns an error if that file already exists (acting as if
	// the `os.O_EXCL` mode is given in a bitmask to os.Open).
	Store(sha []byte, r io.Reader) (n int64, err error)
}
