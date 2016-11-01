package tools_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/github/git-lfs/tools"
	"github.com/stretchr/testify/assert"
)

func TestCleanPathsCleansPaths(t *testing.T) {
	cleaned := tools.CleanPaths("/foo/bar/,/foo/bar/baz", ",")

	assert.Equal(t, []string{"/foo/bar", "/foo/bar/baz"}, cleaned)
}

func TestCleanPathsReturnsNoResultsWhenGivenNoPaths(t *testing.T) {
	cleaned := tools.CleanPaths("", ",")

	assert.Empty(t, cleaned)
}

type TestIncludeExcludeCase struct {
	expectedResult bool
	includes       []string
	excludes       []string
}

func TestFilterIncludeExclude(t *testing.T) {

	cases := []TestIncludeExcludeCase{
		// Null case
		TestIncludeExcludeCase{true, nil, nil},
		// Inclusion
		TestIncludeExcludeCase{true, []string{"*.dat"}, nil},
		TestIncludeExcludeCase{true, []string{"file*.dat"}, nil},
		TestIncludeExcludeCase{true, []string{"file*"}, nil},
		TestIncludeExcludeCase{true, []string{"*name.dat"}, nil},
		TestIncludeExcludeCase{false, []string{"/*.dat"}, nil},
		TestIncludeExcludeCase{false, []string{"otherfolder/*.dat"}, nil},
		TestIncludeExcludeCase{true, []string{"test/filename.dat"}, nil},
		TestIncludeExcludeCase{true, []string{"test/filename.dat"}, nil},
		TestIncludeExcludeCase{false, []string{"blank", "something", "foo"}, nil},
		TestIncludeExcludeCase{false, []string{"test/notfilename.dat"}, nil},
		TestIncludeExcludeCase{true, []string{"test"}, nil},
		TestIncludeExcludeCase{true, []string{"test/*"}, nil},
		TestIncludeExcludeCase{false, []string{"nottest"}, nil},
		TestIncludeExcludeCase{false, []string{"nottest/*"}, nil},
		TestIncludeExcludeCase{true, []string{"test/fil*"}, nil},
		TestIncludeExcludeCase{false, []string{"test/g*"}, nil},
		TestIncludeExcludeCase{true, []string{"tes*/*"}, nil},
		// Exclusion
		TestIncludeExcludeCase{false, nil, []string{"*.dat"}},
		TestIncludeExcludeCase{false, nil, []string{"file*.dat"}},
		TestIncludeExcludeCase{false, nil, []string{"file*"}},
		TestIncludeExcludeCase{false, nil, []string{"*name.dat"}},
		TestIncludeExcludeCase{true, nil, []string{"/*.dat"}},
		TestIncludeExcludeCase{true, nil, []string{"otherfolder/*.dat"}},
		TestIncludeExcludeCase{false, nil, []string{"test/filename.dat"}},
		TestIncludeExcludeCase{false, nil, []string{"blank", "something", "test/filename.dat", "foo"}},
		TestIncludeExcludeCase{true, nil, []string{"blank", "something", "foo"}},
		TestIncludeExcludeCase{true, nil, []string{"test/notfilename.dat"}},
		TestIncludeExcludeCase{false, nil, []string{"test"}},
		TestIncludeExcludeCase{false, nil, []string{"test/*"}},
		TestIncludeExcludeCase{true, nil, []string{"nottest"}},
		TestIncludeExcludeCase{true, nil, []string{"nottest/*"}},
		TestIncludeExcludeCase{false, nil, []string{"test/fil*"}},
		TestIncludeExcludeCase{true, nil, []string{"test/g*"}},
		TestIncludeExcludeCase{false, nil, []string{"tes*/*"}},

		// Both
		TestIncludeExcludeCase{true, []string{"test/filename.dat"}, []string{"test/notfilename.dat"}},
		TestIncludeExcludeCase{false, []string{"test"}, []string{"test/filename.dat"}},
		TestIncludeExcludeCase{true, []string{"test/*"}, []string{"test/notfile*"}},
		TestIncludeExcludeCase{false, []string{"test/*"}, []string{"test/file*"}},
		TestIncludeExcludeCase{false, []string{"another/*", "test/*"}, []string{"test/notfilename.dat", "test/filename.dat"}},
	}

	for _, c := range cases {
		assert.Equal(t, c.expectedResult, tools.FilenamePassesIncludeExcludeFilter("test/filename.dat", c.includes, c.excludes), c)
		if runtime.GOOS == "windows" {
			// also test with \ path separators, tolerate mixed separators
			for i, inc := range c.includes {
				c.includes[i] = strings.Replace(inc, "/", "\\", -1)
			}
			for i, ex := range c.excludes {
				c.excludes[i] = strings.Replace(ex, "/", "\\", -1)
			}
			assert.Equal(t, c.expectedResult, tools.FilenamePassesIncludeExcludeFilter("test/filename.dat", c.includes, c.excludes), c)
		}
	}
}

func TestFastWalkBasic(t *testing.T) {
	rootDir, err := ioutil.TempDir(os.TempDir(), "GitLfsTestFastWalkBasic")
	if err != nil {
		assert.FailNow(t, "Unable to get temp dir: %v", err)
	}

	defer os.RemoveAll(rootDir)
	os.Chdir(rootDir)
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
		numFiles := 10
		expectedEntries = append(expectedEntries, dir)
		if i >= 3 && i <= 5 {
			// Bulk test to ensure works with > 1 batch
			numFiles = 160
		}
		for f := 0; f < numFiles; f++ {
			filename := filepath.Join(dir, fmt.Sprintf("file%d.txt", f))
			ioutil.WriteFile(filename, []byte("TEST"), 0644)
			expectedEntries = append(expectedEntries, filename)
		}
	}

	fchan, errchan := tools.FastWalk(dirs[0], nil, nil)
	gotEntries, gotErrors := collectFastWalkResults(fchan, errchan)

	assert.Equal(t, 0, len(gotErrors))

	sort.Strings(expectedEntries)
	sort.Strings(gotEntries)
	assert.Equal(t, expectedEntries, gotEntries)

}

func collectFastWalkResults(fchan <-chan tools.FastWalkInfo, errchan <-chan error) ([]string, []error) {
	gotEntries := make([]string, 0, 1000)
	gotErrors := make([]error, 0, 5)
	var waitg sync.WaitGroup
	waitg.Add(2)
	go func() {
		for o := range fchan {
			gotEntries = append(gotEntries, filepath.Join(o.ParentDir, o.Info.Name()))
		}
		waitg.Done()
	}()
	go func() {
		for err := range errchan {
			gotErrors = append(gotErrors, err)
		}
		waitg.Done()
	}()
	waitg.Wait()

	return gotEntries, gotErrors
}
