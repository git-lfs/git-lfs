package odb

import "io"

// Storer implements a storage engine for reading, writing, and creating
// io.ReadWriters that can store information about loose objects
type Storer interface {
	// Open returns a handle on an existing object keyed by the given SHA.
	// It returns an error if that file does not already exist.
	Open(sha []byte) (f io.ReadWriteCloser, err error)

	// Create returns a handle on a new object keyed by the given SHA. It
	// returns an error if that file already exists (acting as if the
	// `os.O_EXCL` mode is given in a bitmask to os.Open.
	Create(sha []byte) (f io.ReadWriteCloser, err error)
}
