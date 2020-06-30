package gitobj

import (
	"bufio"
	"hash"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/git-lfs/gitobj/v2/pack"
	"github.com/git-lfs/gitobj/v2/storage"
)

// NewFilesystemBackend initializes a new filesystem-based backend,
// optionally with additional alternates as specified in the
// `alternates` variable. The syntax is that of the Git environment variable
// GIT_ALTERNATE_OBJECT_DIRECTORIES.  The hash algorithm used is specified by
// the algo parameter.
func NewFilesystemBackend(root, tmp, alternates string, algo hash.Hash) (storage.Backend, error) {
	fsobj := newFileStorer(root, tmp)
	packs, err := pack.NewStorage(root, algo)
	if err != nil {
		return nil, err
	}

	storage, err := findAllBackends(fsobj, packs, root, algo)
	if err != nil {
		return nil, err
	}

	storage, err = addAlternatesFromEnvironment(storage, alternates, algo)
	if err != nil {
		return nil, err
	}

	return &filesystemBackend{
		fs:       fsobj,
		backends: storage,
	}, nil
}

func findAllBackends(mainLoose *fileStorer, mainPacked *pack.Storage, root string, algo hash.Hash) ([]storage.Storage, error) {
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
		storage, err = addAlternateDirectory(storage, scanner.Text(), algo)
		if err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return storage, nil
}

func addAlternateDirectory(s []storage.Storage, dir string, algo hash.Hash) ([]storage.Storage, error) {
	s = append(s, newFileStorer(dir, ""))
	pack, err := pack.NewStorage(dir, algo)
	if err != nil {
		return s, err
	}
	s = append(s, pack)
	return s, nil
}

func addAlternatesFromEnvironment(s []storage.Storage, env string, algo hash.Hash) ([]storage.Storage, error) {
	if len(env) == 0 {
		return s, nil
	}

	for _, dir := range splitAlternateString(env, alternatesSeparator) {
		var err error
		s, err = addAlternateDirectory(s, dir, algo)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

var (
	octalEscape  = regexp.MustCompile("\\\\[0-7]{1,3}")
	hexEscape    = regexp.MustCompile("\\\\x[0-9a-fA-F]{2}")
	replacements = []struct {
		olds string
		news string
	}{
		{`\a`, "\a"},
		{`\b`, "\b"},
		{`\t`, "\t"},
		{`\n`, "\n"},
		{`\v`, "\v"},
		{`\f`, "\f"},
		{`\r`, "\r"},
		{`\\`, "\\"},
		{`\"`, "\""},
		{`\'`, "'"},
	}
)

func splitAlternateString(env string, separator string) []string {
	dirs := strings.Split(env, separator)
	for i, s := range dirs {
		if !strings.HasPrefix(s, `"`) || !strings.HasSuffix(s, `"`) {
			continue
		}

		// Strip leading and trailing quotation marks
		s = s[1 : len(s)-1]
		for _, repl := range replacements {
			s = strings.Replace(s, repl.olds, repl.news, -1)
		}
		s = octalEscape.ReplaceAllStringFunc(s, func(inp string) string {
			val, _ := strconv.ParseUint(inp[1:], 8, 64)
			return string([]byte{byte(val)})
		})
		s = hexEscape.ReplaceAllStringFunc(s, func(inp string) string {
			val, _ := strconv.ParseUint(inp[2:], 16, 64)
			return string([]byte{byte(val)})
		})
		dirs[i] = s
	}
	return dirs
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
