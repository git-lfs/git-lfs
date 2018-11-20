package gitobj

import (
	"bufio"
	"io"
	"os"
	"path"

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

	storage, err := findAllBackends(fsobj, packs, root)
	if err != nil {
		return nil, err
	}

	return &filesystemBackend{
		fs:       fsobj,
		backends: storage,
	}, nil
}

func findAllBackends(mainLoose *fileStorer, mainPacked *pack.Storage, root string) ([]storage.Storage, error) {
	storage := make([]storage.Storage, 2)
	storage[0] = mainLoose
	storage[1] = mainPacked
	f, err := os.Open(path.Join(root, "info", "alternates"))
	if err != nil {
		// No alternates file, no problem.
		if err != os.ErrNotExist {
			return storage, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		storage = append(storage, newFileStorer(scanner.Text(), ""))
		pack, err := pack.NewStorage(scanner.Text())
		if err != nil {
			return nil, err
		}
		storage = append(storage, pack)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return storage, nil
}

// NewMemoryBackend initializes a new memory-based backend.
//
// A value of "nil" is acceptable and indicates that no entries should be added
// to the memory backend at construction time.
func NewMemoryBackend(m map[string]io.ReadWriter) (storage.Backend, error) {
	return &memoryBackend{ms: newMemoryStorer(m)}, nil
}

type filesystemBackend struct {
	fs       *fileStorer
	backends []storage.Storage
}

func (b *filesystemBackend) Storage() (storage.Storage, storage.WritableStorage) {
	return storage.MultiStorage(b.backends...), b.fs
}

type memoryBackend struct {
	ms *memoryStorer
}

func (b *memoryBackend) Storage() (storage.Storage, storage.WritableStorage) {
	return b.ms, b.ms
}
