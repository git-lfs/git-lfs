package odb

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// MemoryStorer is an implementation of the Storer interface that holds data for
// the object database in memory.
type MemoryStorer struct {
	// mu guards reads and writes to the map "fs" below.
	mu *sync.Mutex
	// fs maps a hex-encoded SHA to a bytes.Buffer wrapped in a no-op closer
	// type.
	fs map[string]*bufCloser
}

var _ Storer = (*MemoryStorer)(nil)

// NewMemoryStorer initializes a new MemoryStorer instance with the given
// initial set.
//
// A value of "nil" is acceptable and indicates that no entries shall be added
// to the memory storer at/during construction time.
func NewMemoryStorer(m map[string]io.ReadWriter) *MemoryStorer {
	fs := make(map[string]*bufCloser, len(m))
	for n, rw := range m {
		fs[n] = &bufCloser{rw}
	}

	return &MemoryStorer{
		mu: new(sync.Mutex),
		fs: fs,
	}
}

// Store implements the Storer.Store function and copies the data given in "r"
// into an object entry in the memory. If an object given by that SHA "sha" is
// already indexed in the database, Store will panic().
func (ms *MemoryStorer) Store(sha []byte, r io.Reader) (n int64, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := fmt.Sprintf("%x", sha)
	if _, ok := ms.fs[key]; ok {
		panic(fmt.Sprintf("git/odb: memory storage create %x, already exists", sha))
	}

	ms.fs[key] = &bufCloser{new(bytes.Buffer)}
	return io.Copy(ms.fs[key], r)
}

// Open implements the Storer.Open function, and returns a io.ReadWriteCloser
// for the given SHA. If a reader for the given SHA does not exist an error will
// be returned.
func (ms *MemoryStorer) Open(sha []byte) (f io.ReadWriteCloser, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := fmt.Sprintf("%x", sha)
	if _, ok := ms.fs[key]; !ok {
		panic(fmt.Sprintf("git/odb: memory storage cannot open %x, doesn't exist", sha))
	}
	return ms.fs[key], nil
}

// bufCloser wraps a type satisfying the io.ReadWriter interface with a no-op
// Close() function, thus implementing the io.ReadWriteCloser composite
// interface.
type bufCloser struct {
	io.ReadWriter
}

var _ io.ReadWriteCloser = (*bufCloser)(nil)

// Close implements io.Closer, and returns nothing.
func (b *bufCloser) Close() error { return nil }
