package gitobj

import (
	"io"

	"github.com/git-lfs/gitobj/pack"
	"github.com/git-lfs/gitobj/storage"
)

// NewFilesystemBackend initializes a new filesystem-based backend.
func NewFilesystemBackend(root, tmp string) (storage.Backend, error) {
	fsobj := newFileStorer(root, tmp)
	packs, err := pack.NewStorage(root)
	if err != nil {
		return nil, err
	}

	return &filesystemBackend{fs: fsobj, packs: packs}, nil
}

// NewMemoryBackend initializes a new memory-based backend.
//
// A value of "nil" is acceptable and indicates that no entries should be added
// to the memory backend at construction time.
func NewMemoryBackend(m map[string]io.ReadWriter) (storage.Backend, error) {
	return &memoryBackend{ms: newMemoryStorer(m)}, nil
}

type filesystemBackend struct {
	fs    *fileStorer
	packs *pack.Storage
}

func (b *filesystemBackend) Storage() (storage.Storage, storage.WritableStorage) {
	return storage.MultiStorage(b.fs, b.packs), b.fs
}

type memoryBackend struct {
	ms *memoryStorer
}

func (b *memoryBackend) Storage() (storage.Storage, storage.WritableStorage) {
	return b.ms, b.ms
}
