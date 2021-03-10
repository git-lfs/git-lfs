package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanPathsCleansPaths(t *testing.T) {
	cleaned := CleanPaths("/foo/bar/,/foo/bar/baz", ",")
	assert.Equal(t, []string{"/foo/bar", "/foo/bar/baz"}, cleaned)
}

func TestCleanPathsReturnsNoResultsWhenGivenNoPaths(t *testing.T) {
	cleaned := CleanPaths("", ",")
	assert.Empty(t, cleaned)
}

type ExpandPathTestCase struct {
	Path   string
	Expand bool

	Want    string
	WantErr string

	currentUser func() (*user.User, error)
	lookupUser  func(who string) (*user.User, error)
}

func (c *ExpandPathTestCase) Assert(t *testing.T) {
	if c.currentUser != nil {
		oldCurrentUser := currentUser
		currentUser = c.currentUser
		defer func() { currentUser = oldCurrentUser }()
	}

	if c.lookupUser != nil {
		oldLookupUser := lookupUser
		lookupUser = c.lookupUser
		defer func() { lookupUser = oldLookupUser }()
	}

	got, err := ExpandPath(c.Path, c.Expand)
	if err != nil || len(c.WantErr) > 0 {
		assert.EqualError(t, err, c.WantErr)
	}
	assert.Equal(t, filepath.ToSlash(c.Want), filepath.ToSlash(got))
}

func TestExpandPath(t *testing.T) {
	for desc, c := range map[string]*ExpandPathTestCase{
		"no expand": {
			Path: "/path/to/hooks",
			Want: "/path/to/hooks",
		},
		"current": {
			Path: "~/path/to/hooks",
			Want: "/home/jane/path/to/hooks",
			currentUser: func() (*user.User, error) {
				return &user.User{
					HomeDir: "/home/jane",
				}, nil
			},
		},
		"current, slash": {
			Path: "~/",
			Want: "/home/jane",
			currentUser: func() (*user.User, error) {
				return &user.User{
					HomeDir: "/home/jane",
				}, nil
			},
		},
		"current, no slash": {
			Path: "~",
			Want: "/home/jane",
			currentUser: func() (*user.User, error) {
				return &user.User{
					HomeDir: "/home/jane",
				}, nil
			},
		},
		"non-current": {
			Path: "~other/path/to/hooks",
			Want: "/home/special/path/to/hooks",
			lookupUser: func(who string) (*user.User, error) {
				assert.Equal(t, "other", who)
				return &user.User{
					HomeDir: "/home/special",
				}, nil
			},
		},
		"non-current, no slash": {
			Path: "~other",
			Want: "/home/special",
			lookupUser: func(who string) (*user.User, error) {
				assert.Equal(t, "other", who)
				return &user.User{
					HomeDir: "/home/special",
				}, nil
			},
		},
		"non-current (missing)": {
			Path:    "~other/path/to/hooks",
			WantErr: "could not find user other: missing",
			lookupUser: func(who string) (*user.User, error) {
				assert.Equal(t, "other", who)
				return nil, fmt.Errorf("missing")
			},
		},
	} {
		t.Run(desc, c.Assert)
	}
}

type ExpandConfigPathTestCase struct {
	Path        string
	DefaultPath string

	Want    string
	WantErr string

	currentUser      func() (*user.User, error)
	lookupConfigHome func() string
}

func (c *ExpandConfigPathTestCase) Assert(t *testing.T) {
	if c.currentUser != nil {
		oldCurrentUser := currentUser
		currentUser = c.currentUser
		defer func() { currentUser = oldCurrentUser }()
	}

	if c.lookupConfigHome != nil {
		oldLookupConfigHome := lookupConfigHome
		lookupConfigHome = c.lookupConfigHome
		defer func() { lookupConfigHome = oldLookupConfigHome }()
	}

	got, err := ExpandConfigPath(c.Path, c.DefaultPath)
	if err != nil || len(c.WantErr) > 0 {
		assert.EqualError(t, err, c.WantErr)
	}
	assert.Equal(t, filepath.ToSlash(c.Want), filepath.ToSlash(got))
}

func TestExpandConfigPath(t *testing.T) {
	os.Unsetenv("XDG_CONFIG_HOME")
	for desc, c := range map[string]*ExpandConfigPathTestCase{
		"unexpanded full path": {
			Path: "/path/to/attributes",
			Want: "/path/to/attributes",
		},
		"expanded full path": {
			Path: "~/path/to/attributes",
			Want: "/home/pat/path/to/attributes",
			currentUser: func() (*user.User, error) {
				return &user.User{
					HomeDir: "/home/pat",
				}, nil
			},
		},
		"expanded default path": {
			DefaultPath: "git/attributes",
			Want:        "/home/pat/.config/git/attributes",
			currentUser: func() (*user.User, error) {
				return &user.User{
					HomeDir: "/home/pat",
				}, nil
			},
		},
		"XDG_CONFIG_HOME set": {
			DefaultPath: "git/attributes",
			Want:        "/home/pat/configpath/git/attributes",
			lookupConfigHome: func() string {
				return "/home/pat/configpath"
			},
		},
	} {
		t.Run(desc, c.Assert)
	}
}

func TestFastWalkBasic(t *testing.T) {
	rootDir, err := ioutil.TempDir(os.TempDir(), "GitLfsTestFastWalkBasic")
	if err != nil {
		assert.FailNow(t, "Unable to get temp dir: %v", err)
	}
	defer os.RemoveAll(rootDir)
	os.Chdir(rootDir)

	expectedEntries := createFastWalkInputData(10, 160)

	walker := fastWalkWithExcludeFiles(expectedEntries[0])
	gotEntries, gotErrors := collectFastWalkResults(walker.ch)

	assert.Empty(t, gotErrors)

	sort.Strings(expectedEntries)
	sort.Strings(gotEntries)
	assert.Equal(t, expectedEntries, gotEntries)
}

// Make test data - ensure you've Chdir'ed into a temp dir first
// Returns list of files/dirs that are created
// First entry is the parent dir of all others
func createFastWalkInputData(smallFolder, largeFolder int) []string {
	dirs := []string{
		"testroot",
		"testroot/folder1",
		"testroot/folder2",
		"testroot/folder2/subfolder1",
		"testroot/folder2/subfolder2",
		"testroot/folder2/subfolder3",
		"testroot/folder2/subfolder4",
		"testroot/folder2/subfolder4/subsub",
	}
	expectedEntries := make([]string, 0, 250)

	for i, dir := range dirs {
		os.MkdirAll(dir, 0755)
		numFiles := smallFolder
		expectedEntries = append(expectedEntries, dir)
		if i >= 3 && i <= 5 {
			// Bulk test to ensure works with > 1 batch
			numFiles = largeFolder
		}
		for f := 0; f < numFiles; f++ {
			filename := join(dir, fmt.Sprintf("file%d.txt", f))
			ioutil.WriteFile(filename, []byte("TEST"), 0644)
			expectedEntries = append(expectedEntries, filename)
		}
	}

	return expectedEntries
}

func collectFastWalkResults(fchan <-chan fastWalkInfo) ([]string, []error) {
	gotEntries := make([]string, 0, 1000)
	gotErrors := make([]error, 0, 5)
	for o := range fchan {
		if o.Err != nil {
			gotErrors = append(gotErrors, o.Err)
		} else {
			if len(o.ParentDir) == 0 {
				gotEntries = append(gotEntries, o.Info.Name())
			} else {
				gotEntries = append(gotEntries, join(o.ParentDir, o.Info.Name()))
			}
		}
	}

	return gotEntries, gotErrors
}

func getFileMode(filename string) os.FileMode {
	s, err := os.Stat(filename)
	if err != nil {
		return 0000
	}
	return s.Mode()
}

// uniq creates an element-wise copy of "xs" containing only unique elements in
// the same order.
func uniq(xs []string) []string {
	seen := make(map[string]struct{})
	uniq := make([]string, 0, len(xs))

	for _, x := range xs {
		if _, ok := seen[x]; !ok {
			seen[x] = struct{}{}
			uniq = append(uniq, x)
		}
	}

	return uniq
}

func TestSetWriteFlag(t *testing.T) {
	f, err := ioutil.TempFile("", "lfstestwriteflag")
	assert.Nil(t, err)
	filename := f.Name()
	defer os.Remove(filename)
	f.Close()
	// Set up with read/write bit for all but no execute
	assert.Nil(t, os.Chmod(filename, 0666))

	assert.Nil(t, SetFileWriteFlag(filename, false))
	// should turn off all write
	assert.EqualValues(t, 0444, getFileMode(filename))
	assert.Nil(t, SetFileWriteFlag(filename, true))
	// should only add back user write (on Mac/Linux)
	if runtime.GOOS == "windows" {
		assert.EqualValues(t, 0666, getFileMode(filename))
	} else {
		assert.EqualValues(t, 0644, getFileMode(filename))
	}

	// Can't run selective UGO tests on Windows as doesn't support separate roles
	// Also Golang only supports read/write but not execute on Windows
	if runtime.GOOS != "windows" {
		// Set up with read/write/execute bit for all but no execute
		assert.Nil(t, os.Chmod(filename, 0777))
		assert.Nil(t, SetFileWriteFlag(filename, false))
		// should turn off all write but not execute
		assert.EqualValues(t, 0555, getFileMode(filename))
		assert.Nil(t, SetFileWriteFlag(filename, true))
		// should only add back user write (on Mac/Linux)
		if runtime.GOOS == "windows" {
			assert.EqualValues(t, 0777, getFileMode(filename))
		} else {
			assert.EqualValues(t, 0755, getFileMode(filename))
		}

		assert.Nil(t, os.Chmod(filename, 0440))
		assert.Nil(t, SetFileWriteFlag(filename, false))
		assert.EqualValues(t, 0440, getFileMode(filename))
		assert.Nil(t, SetFileWriteFlag(filename, true))
		// should only add back user write
		assert.EqualValues(t, 0640, getFileMode(filename))
	}
}

func TestExecutablePermissions(t *testing.T) {
	assert.EqualValues(t, os.FileMode(0755), ExecutablePermissions(0644))
	assert.EqualValues(t, os.FileMode(0750), ExecutablePermissions(0640))
	assert.EqualValues(t, os.FileMode(0700), ExecutablePermissions(0600))
}
