package storage

import (
	"io"

	"github.com/git-lfs/gitobj/v2/errors"
)

// Storage implements an interface for reading, but not writing, objects in an
// object database.
type multiStorage struct {
	impls []Storage
}

func MultiStorage(args ...Storage) Storage {
	return &multiStorage{impls: args}
}

// Open returns a handle on an existing object keyed by the given object
// ID.  It returns an error if that file does not already exist.
func (m *multiStorage) Open(oid []byte) (f io.ReadCloser, err error) {
	for _, s := range m.impls {
		f, err := s.Open(oid)
		if err != nil {
			if errors.IsNoSuchObject(err) {
				continue
			}
			return nil, err
		}
		if s.IsCompressed() {
			return newDecompressingReadCloser(f)
		}
		return f, nil
	}
	return nil, errors.NoSuchObject(oid)
}

// Close closes the filesystem, after which no more operations are
// allowed.
func (m *multiStorage) Close() error {
	for _, s := range m.impls {
		if err := s.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Compressed indicates whether data read from this storage source will
// be zlib-compressed.
func (m *multiStorage) IsCompressed() bool {
	// To ensure we can read from any Storage type, we automatically
	// decompress items if they need it.
	return false
}
