package tools

import (
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/tr"
)

var (
	errInvalidDir = errors.New(tr.Tr.Get("invalid directory"))
	errNotDir     = errors.New(tr.Tr.Get("not a directory"))
)

type DirWalker struct {
	parentPath string
	path       string
	config     repositoryPermissionFetcher
}

// The parentPath parameter is assumed to be a valid path to a directory
// in the filesystem.
//
// The filePath parameter must be a relative file path as provided by Git,
// with only the "/" character as a separator and no empty or "." or ".."
// path segments.  Absolute paths are not supported.
func NewDirWalkerForFile(parentPath string, filePath string, config repositoryPermissionFetcher) *DirWalker {
	var path string
	i := strings.LastIndexByte(filePath, '/')
	if i >= 0 {
		path = filePath[0:i]
	}

	return &DirWalker{
		parentPath: parentPath,
		path:       path,
		config:     config,
	}
}

// walk() checks each directory in a relative path, starting from the
// initial parent path, and optionally creates any missing directories
// in the path.
//
// If an existing file or something else other than a directory conflicts
// with a directory in the path, walk() returns an error.
//
// If the create option is false, walk() returns ErrNotExist when a
// directory is not found.
//
// Note that for performance reasons and to be consistent with Git's
// implementation, walk() does not guard against TOCTOU (time-of-check/
// time-of-use) races, as the methods of the os.Root type do.
func (w *DirWalker) walk(create bool) error {
	currentPath := w.parentPath

	n := len(w.path)
	for n > 0 {
		currentDir := w.path
		nextDirIndex := n
		i := strings.IndexByte(w.path, '/')
		if i >= 0 {
			currentDir = w.path[0:i]
			nextDirIndex = i + 1
		}

		// These should never occur in Git paths.
		if currentDir == "" || currentDir == "." || currentDir == ".." {
			return errors.Join(errors.New(tr.Tr.Get("invalid directory %q in path: %q", currentDir, w.path)), errInvalidDir)
		}

		if currentPath == "" {
			currentPath = currentDir
		} else {
			currentPath += "/" + currentDir
		}

		stat, err := os.Lstat(currentPath)
		if err != nil {
			if !os.IsNotExist(err) || !create {
				return err
			}

			err = Mkdir(currentPath, w.config)
			if err != nil {
				return err
			}
		} else if !stat.Mode().IsDir() {
			return errors.Join(errors.New(tr.Tr.Get("not a directory: %q", currentPath)), errNotDir)
		}

		w.parentPath = currentPath
		w.path = w.path[nextDirIndex:]
		n -= nextDirIndex
	}

	return nil
}

func (w *DirWalker) Walk() error {
	return w.walk(false)
}

func (w *DirWalker) WalkAndCreate() error {
	return w.walk(true)
}
