package pack

import (
	"hash"
	"io"
)

// Storage implements the storage.Storage interface.
type Storage struct {
	packs *Set
}

// NewStorage returns a new storage object based on a pack set.
func NewStorage(root string, algo hash.Hash) (*Storage, error) {
	packs, err := NewSet(root, algo)
	if err != nil {
		return nil, err
	}
	return &Storage{packs: packs}, nil
}

// Open implements the storage.Storage.Open interface.
func (f *Storage) Open(oid []byte) (r io.ReadCloser, err error) {
	obj, err := f.packs.Object(oid)
	if err != nil {
		return nil, err
	}
	return &delayedObjectReader{obj: obj}, nil
}

// Open implements the storage.Storage.Open interface.
func (f *Storage) Close() error {
	return f.packs.Close()
}

// IsCompressed returns false, because data returned is already decompressed.
func (f *Storage) IsCompressed() bool {
	return false
}
