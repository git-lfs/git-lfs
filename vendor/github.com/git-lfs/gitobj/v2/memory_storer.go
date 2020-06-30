package gitobj

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/git-lfs/gitobj/v2/errors"
)

// memoryStorer is an implementation of the storer interface that holds data for
// the object database in memory.
type memoryStorer struct {
	// mu guards reads and writes to the map "fs" below.
	mu *sync.Mutex
	// fs maps a hex-encoded SHA to a bytes.Buffer wrapped in a no-op closer
	// type.
	fs map[string]*bufCloser
}

// newMemoryStorer initializes a new memoryStorer instance with the given
// initial set.
//
// A value of "nil" is acceptable and indicates that no entries shall be added
// to the memory storer at/during construction time.
func newMemoryStorer(m map[string]io.ReadWriter) *memoryStorer {
	fs := make(map[string]*bufCloser, len(m))
	for n, rw := range m {
		fs[n] = &bufCloser{rw}
	}

	return &memoryStorer{
		mu: new(sync.Mutex),
		fs: fs,
	}
}

// Store implements the storer.Store function and copies the data given in "r"
// into an object entry in the memory. If an object given by that SHA "sha" is
// already indexed in the database, Store will panic().
func (ms *memoryStorer) Store(sha []byte, r io.Reader) (n int64, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := fmt.Sprintf("%x", sha)

	ms.fs[key] = &bufCloser{new(bytes.Buffer)}
	return io.Copy(ms.fs[key], r)
}

// Open implements the storer.Open function, and returns a io.ReadWriteCloser
// for the given SHA. If a reader for the given SHA does not exist an error will
// be returned.
func (ms *memoryStorer) Open(sha []byte) (f io.ReadCloser, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := fmt.Sprintf("%x", sha)
	if _, ok := ms.fs[key]; !ok {
		return nil, errors.NoSuchObject(sha)
	}
	return ms.fs[key], nil
}

// Close closes the memory storer.
func (ms *memoryStorer) Close() error {
	return nil
}

// IsCompressed returns true, because the memory storer returns compressed data.
func (ms *memoryStorer) IsCompressed() bool {
	return true
}

// bufCloser wraps a type satisfying the io.ReadWriter interface with a no-op
// Close() function, thus implementing the io.ReadWriteCloser composite
// interface.
type bufCloser struct {
	io.ReadWriter
}

// Close implements io.Closer, and returns nothing.
func (b *bufCloser) Close() error { return nil }
