package tools

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type newDirWalkerForFileTestCase struct {
	filePath        string
	expectedDirPath string
}

func (c *newDirWalkerForFileTestCase) Assert(t *testing.T) {
	w := NewDirWalkerForFile("", c.filePath, nil)
	assert.Equal(t, c.expectedDirPath, w.path)
}

func TestNewDirWalkerForFile(t *testing.T) {
	for desc, c := range map[string]*newDirWalkerForFileTestCase{
		"filename only":            {"foo.bin", ""},
		"path with one dir":        {"abc/foo.bin", "abc"},
		"path with two dirs":       {"abc/def/foo.bin", "abc/def"},
		"path with leading slash":  {"/foo.bin", ""},
		"path with trailing slash": {"abc/", "abc"},
		"bare slash":               {"/", ""},
		"empty path":               {"", ""},
	} {
		t.Run(desc, c.Assert)
	}
}

type dirWalkerTestConfig struct{}

func (c *dirWalkerTestConfig) RepositoryPermissions(executable bool) os.FileMode {
	return os.FileMode(0755)
}

type dirWalkerWalkTestCase struct {
	parentPath string
	path       string
	create     bool

	existsPath string
	existsFile string
	existsLink string

	expectedParentPath string
	expectedPath       string
	expectedErr        error

	walker *DirWalker
}

func (c *dirWalkerWalkTestCase) prependParentPath(path string) string {
	if path == "" {
		return c.parentPath
	} else if c.parentPath == "" {
		return path
	} else if path[0] == '/' {
		return "/" + c.parentPath + path
	} else {
		return c.parentPath + "/" + path
	}
}

func (c *dirWalkerWalkTestCase) setupPaths(t *testing.T, parentPath string) error {
	c.parentPath = parentPath

	if parentPath != "" {
		if err := os.MkdirAll(parentPath, 0755); err != nil {
			return fmt.Errorf("unable to create path: %w", err)
		}
	}

	if c.existsPath != "" {
		c.existsPath = c.prependParentPath(c.existsPath)
		if err := os.MkdirAll(c.existsPath, 0755); err != nil {
			return fmt.Errorf("unable to create path: %w", err)
		}
	}

	if c.existsFile != "" {
		c.existsFile = c.prependParentPath(c.existsFile)
		f, err := os.Create(c.existsFile)
		if err != nil {
			return fmt.Errorf("unable to create file: %w", err)
		}
		f.Close()
	}

	if c.existsLink != "" {
		c.existsLink = c.prependParentPath(c.existsLink)
		if err := os.Symlink(t.TempDir(), c.existsLink); err != nil {
			return fmt.Errorf("unable to create symbolic link: %w", err)
		}
	}

	c.expectedParentPath = c.prependParentPath(c.expectedParentPath)

	return nil
}

func (c *dirWalkerWalkTestCase) Assert(t *testing.T) {
	c.walker.parentPath = c.parentPath
	c.walker.path = c.path

	err := c.walker.walk(c.create)

	assert.Equal(t, c.expectedParentPath, c.walker.parentPath, "found path does not match")
	assert.Equal(t, c.expectedPath, c.walker.path, "missing path does not match")
	if c.expectedErr == nil {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
		assert.True(t, errors.Is(err, c.expectedErr), "wrong error type")
	}
}

func TestDirWalkerWalk(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	defer os.Chdir(wd)

	for desc, c := range map[string]*dirWalkerWalkTestCase{
		"empty path": {},
		"one extant dir": {
			path:               "abc",
			existsPath:         "abc",
			expectedParentPath: "abc",
		},
		"one missing dir": {
			path:         "abc",
			expectedPath: "abc",
			expectedErr:  os.ErrNotExist,
		},
		"two extant dirs": {
			path:               "abc/def",
			existsPath:         "abc/def",
			expectedParentPath: "abc/def",
		},
		"two missing dirs": {
			path:         "abc/def",
			expectedPath: "abc/def",
			expectedErr:  os.ErrNotExist,
		},
		"three extant dirs": {
			path:               "abc/def/ghi",
			existsPath:         "abc/def/ghi",
			expectedParentPath: "abc/def/ghi",
		},
		"three missing dirs": {
			path:         "abc/def/ghi",
			expectedPath: "abc/def/ghi",
			expectedErr:  os.ErrNotExist,
		},
		"one extant dir and one missing dir": {
			path:               "abc/def",
			existsPath:         "abc",
			expectedParentPath: "abc",
			expectedPath:       "def",
			expectedErr:        os.ErrNotExist,
		},
		"one extant dir and two missing dirs": {
			path:               "abc/def/ghi",
			existsPath:         "abc",
			expectedParentPath: "abc",
			expectedPath:       "def/ghi",
			expectedErr:        os.ErrNotExist,
		},
		"two extant dirs and one missing dir": {
			path:               "abc/def/ghi",
			existsPath:         "abc/def",
			expectedParentPath: "abc/def",
			expectedPath:       "ghi",
			expectedErr:        os.ErrNotExist,
		},
		"one missing dir with trailing slash": {
			path:         "abc/",
			expectedPath: "abc/",
			expectedErr:  os.ErrNotExist,
		},
		"one extant dir with trailing slash": {
			path:               "abc/",
			existsPath:         "abc",
			expectedParentPath: "abc",
		},
		"two extant dirs with trailing slash": {
			path:               "abc/def/",
			existsPath:         "abc/def",
			expectedParentPath: "abc/def",
		},
		"one extant dir and one missing dir with trailing slash": {
			path:               "abc/def/",
			existsPath:         "abc",
			expectedParentPath: "abc",
			expectedPath:       "def/",
			expectedErr:        os.ErrNotExist,
		},
		"one conflicting file": {
			path:         "abc",
			existsFile:   "abc",
			expectedPath: "abc",
			expectedErr:  errNotDir,
		},
		"one extant dir and one conflicting file": {
			path:               "abc/def",
			existsPath:         "abc",
			existsFile:         "abc/def",
			expectedParentPath: "abc",
			expectedPath:       "def",
			expectedErr:        errNotDir,
		},
		"two extant dirs and one conflicting file": {
			path:               "abc/def/ghi",
			existsPath:         "abc/def",
			existsFile:         "abc/def/ghi",
			expectedParentPath: "abc/def",
			expectedPath:       "ghi",
			expectedErr:        errNotDir,
		},
		"one extant dir, one conflicting file, and one missing dir": {
			path:               "abc/def/ghi",
			existsPath:         "abc",
			existsFile:         "abc/def",
			expectedParentPath: "abc",
			expectedPath:       "def/ghi",
			expectedErr:        errNotDir,
		},
		"one conflicting symlink": {
			path:         "abc",
			existsLink:   "abc",
			expectedPath: "abc",
			expectedErr:  errNotDir,
		},
		"one extant dir and one conflicting symlink": {
			path:               "abc/def",
			existsPath:         "abc",
			existsLink:         "abc/def",
			expectedParentPath: "abc",
			expectedPath:       "def",
			expectedErr:        errNotDir,
		},
		"two extant dirs and one conflicting symlink": {
			path:               "abc/def/ghi",
			existsPath:         "abc/def",
			existsLink:         "abc/def/ghi",
			expectedParentPath: "abc/def",
			expectedPath:       "ghi",
			expectedErr:        errNotDir,
		},
		"one extant dir, one conflicting symlink, and one missing dir": {
			path:               "abc/def/ghi",
			existsPath:         "abc",
			existsLink:         "abc/def",
			expectedParentPath: "abc",
			expectedPath:       "def/ghi",
			expectedErr:        errNotDir,
		},
		"one extant dir (not modified)": {
			path:               "abc",
			create:             true,
			existsPath:         "abc",
			expectedParentPath: "abc",
		},
		"one created dir": {
			path:               "abc",
			create:             true,
			expectedParentPath: "abc",
		},
		"two extant dirs (not modified)": {
			path:               "abc/def",
			create:             true,
			existsPath:         "abc/def",
			expectedParentPath: "abc/def",
		},
		"two created dirs": {
			path:               "abc/def",
			create:             true,
			expectedParentPath: "abc/def",
		},
		"three extant dirs (not modified)": {
			path:               "abc/def/ghi",
			create:             true,
			existsPath:         "abc/def/ghi",
			expectedParentPath: "abc/def/ghi",
		},
		"three created dirs": {
			path:               "abc/def/ghi",
			create:             true,
			expectedParentPath: "abc/def/ghi",
		},
		"one extant dir and one created dir": {
			path:               "abc/def",
			create:             true,
			existsPath:         "abc",
			expectedParentPath: "abc/def",
		},
		"one extant dir and two created dirs": {
			path:               "abc/def/ghi",
			create:             true,
			existsPath:         "abc",
			expectedParentPath: "abc/def/ghi",
		},
		"two extant dirs and one created dir": {
			path:               "abc/def/ghi",
			create:             true,
			existsPath:         "abc/def",
			expectedParentPath: "abc/def/ghi",
		},
		"one created dir with trailing slash": {
			path:               "abc/",
			create:             true,
			expectedParentPath: "abc",
		},
		"one extant dir with trailing slash (not modified)": {
			path:               "abc/",
			create:             true,
			existsPath:         "abc",
			expectedParentPath: "abc",
		},
		"two extant dirs with trailing slash (not modified)": {
			path:               "abc/def/",
			create:             true,
			existsPath:         "abc/def",
			expectedParentPath: "abc/def",
		},
		"one extant dir and one created dir with trailing slash": {
			path:               "abc/def/",
			create:             true,
			existsPath:         "abc",
			expectedParentPath: "abc/def",
		},
		"one conflicting file (not modified)": {
			path:         "abc",
			create:       true,
			existsFile:   "abc",
			expectedPath: "abc",
			expectedErr:  errNotDir,
		},
		"one extant dir and one conflicting file (not modified)": {
			path:               "abc/def",
			create:             true,
			existsPath:         "abc",
			existsFile:         "abc/def",
			expectedParentPath: "abc",
			expectedPath:       "def",
			expectedErr:        errNotDir,
		},
		"two extant dirs and one conflicting file (not modified)": {
			path:               "abc/def/ghi",
			create:             true,
			existsPath:         "abc/def",
			existsFile:         "abc/def/ghi",
			expectedParentPath: "abc/def",
			expectedPath:       "ghi",
			expectedErr:        errNotDir,
		},
		"one extant dir, one conflicting file, and one missing dir (not modified)": {
			path:               "abc/def/ghi",
			create:             true,
			existsPath:         "abc",
			existsFile:         "abc/def",
			expectedParentPath: "abc",
			expectedPath:       "def/ghi",
			expectedErr:        errNotDir,
		},
		"one conflicting symlink (not modified)": {
			path:         "abc",
			create:       true,
			existsLink:   "abc",
			expectedPath: "abc",
			expectedErr:  errNotDir,
		},
		"one extant dir and one conflicting symlink (not modified)": {
			path:               "abc/def",
			create:             true,
			existsPath:         "abc",
			existsLink:         "abc/def",
			expectedParentPath: "abc",
			expectedPath:       "def",
			expectedErr:        errNotDir,
		},
		"two extant dirs and one conflicting symlink (not modified)": {
			path:               "abc/def/ghi",
			create:             true,
			existsPath:         "abc/def",
			existsLink:         "abc/def/ghi",
			expectedParentPath: "abc/def",
			expectedPath:       "ghi",
			expectedErr:        errNotDir,
		},
		"one extant dir, one conflicting symlink, and one missing dir (not modified)": {
			path:               "abc/def/ghi",
			create:             true,
			existsPath:         "abc",
			existsLink:         "abc/def",
			expectedParentPath: "abc",
			expectedPath:       "def/ghi",
			expectedErr:        errNotDir,
		},
		"invalid bare slash": {
			path:         "/",
			expectedPath: "/",
			expectedErr:  errInvalidDir,
		},
		"invalid multiple slashes": {
			path:               "abc//def",
			existsPath:         "abc",
			expectedParentPath: "abc",
			expectedPath:       "/def",
			expectedErr:        errInvalidDir,
		},
		"invalid leading slash": {
			path:         "/abc",
			existsPath:   "abc",
			expectedPath: "/abc",
			expectedErr:  errInvalidDir,
		},
		"invalid bare dot component": {
			path:         ".",
			expectedPath: ".",
			expectedErr:  errInvalidDir,
		},
		"invalid dot component": {
			path:               "abc/./def",
			existsPath:         "abc/def",
			expectedParentPath: "abc",
			expectedPath:       "./def",
			expectedErr:        errInvalidDir,
		},
		"invalid bare double-dot component": {
			path:         "..",
			expectedPath: "..",
			expectedErr:  errInvalidDir,
		},
		"invalid double-dot component": {
			path:               "abc/../def",
			existsPath:         "abc",
			expectedParentPath: "abc",
			expectedPath:       "../def",
			expectedErr:        errInvalidDir,
		},
	} {
		if err := os.Chdir(t.TempDir()); err != nil {
			t.Errorf("unable to change directory: %s", err)
		}

		c.walker = &DirWalker{
			config: &dirWalkerTestConfig{},
		}

		if err := c.setupPaths(t, ""); err != nil {
			t.Error(err)
			continue
		}

		t.Run(desc, c.Assert)

		// retest with parent path; note that this alters the test case
		if err := c.setupPaths(t, "foo/bar"); err != nil {
			t.Error(err)
			continue
		}

		t.Run(desc+" with parent path", c.Assert)
	}
}
