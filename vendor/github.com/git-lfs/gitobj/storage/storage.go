package storage

import "io"

// Storage implements an interface for reading, but not writing, objects in an
// object database.
type Storage interface {
	// Open returns a handle on an existing object keyed by the given object
	// ID.  It returns an error if that file does not already exist.
	Open(oid []byte) (f io.ReadCloser, err error)

	// Close closes the filesystem, after which no more operations are
	// allowed.
	Close() error

	// Compressed indicates whether data read from this storage source will
	// be zlib-compressed.
	IsCompressed() bool
}

// WritableStorage implements an interface for reading and writing objects in
// an object database.
type WritableStorage interface {
	Storage

	// Store copies the data given in "r" to the unique object path given by
	// "oid". It returns an error if that file already exists (acting as if
	// the `os.O_EXCL` mode is given in a bitmask to os.Open).
	Store(oid []byte, r io.Reader) (n int64, err error)
}
