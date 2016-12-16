package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/subprocess"

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

func TestFileMatch(t *testing.T) {
	assert.True(t, FileMatch("filename.txt", "filename.txt"))
	assert.True(t, FileMatch("*.txt", "filename.txt"))
	assert.False(t, FileMatch("*.tx", "filename.txt"))
	assert.True(t, FileMatch("f*.txt", "filename.txt"))
	assert.False(t, FileMatch("g*.txt", "filename.txt"))
	assert.True(t, FileMatch("file*", "filename.txt"))
	assert.False(t, FileMatch("file", "filename.txt"))

	// With no path separators, should match in subfolders
	assert.True(t, FileMatch("*.txt", "sub/filename.txt"))
	assert.False(t, FileMatch("*.tx", "sub/filename.txt"))
	assert.True(t, FileMatch("f*.txt", "sub/filename.txt"))
	assert.False(t, FileMatch("g*.txt", "sub/filename.txt"))
	assert.True(t, FileMatch("file*", "sub/filename.txt"))
	assert.False(t, FileMatch("file", "sub/filename.txt"))
	// Needs wildcard for exact filename
	assert.True(t, FileMatch("**/filename.txt", "sub/sub/sub/filename.txt"))

	// Should not match dots to subparts
	assert.False(t, FileMatch("*.ign", "sub/shouldignoreme.txt"))

	// Path specific
	assert.True(t, FileMatch("sub", "sub/filename.txt"))
	assert.False(t, FileMatch("sub", "subfilename.txt"))

	// Absolute
	assert.True(t, FileMatch("*.dat", "/path/to/sub/.git/test.dat"))
	assert.True(t, FileMatch("**/.git", "/path/to/sub/.git"))

	// Match anything
	assert.True(t, FileMatch(".", "path.txt"))
	assert.True(t, FileMatch("./", "path.txt"))
	assert.True(t, FileMatch(".\\", "path.txt"))

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
		TestIncludeExcludeCase{false, []string{"*.nam"}, nil},
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
		result := FilenamePassesIncludeExcludeFilter("test/filename.dat", c.includes, c.excludes)
		assert.Equal(t, c.expectedResult, result, "includes: %v excludes: %v", c.includes, c.excludes)
		if runtime.GOOS == "windows" {
			// also test with \ path separators, tolerate mixed separators
			for i, inc := range c.includes {
				c.includes[i] = strings.Replace(inc, "/", "\\", -1)
			}
			for i, ex := range c.excludes {
				c.excludes[i] = strings.Replace(ex, "/", "\\", -1)
			}
			assert.Equal(t, c.expectedResult, FilenamePassesIncludeExcludeFilter("test/filename.dat", c.includes, c.excludes), c)
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

	expectedEntries := createFastWalkInputData(10, 160)

	fchan := fastWalkWithExcludeFiles(expectedEntries[0], "", nil)
	gotEntries, gotErrors := collectFastWalkResults(fchan)

	assert.Empty(t, gotErrors)

	sort.Strings(expectedEntries)
	sort.Strings(gotEntries)
	assert.Equal(t, expectedEntries, gotEntries)

}

func BenchmarkFastWalkGitRepoChannels(b *testing.B) {
	rootDir, err := ioutil.TempDir(os.TempDir(), "GitLfsBenchFastWalkGitRepoChannels")
	if err != nil {
		assert.FailNow(b, "Unable to get temp dir: %v", err)
	}
	defer os.RemoveAll(rootDir)
	os.Chdir(rootDir)
	entries := createFastWalkInputData(1000, 5000)

	for i := 0; i < b.N; i++ {
		var files, errors int
		FastWalkGitRepo(entries[0], func(parent string, info os.FileInfo, err error) {
			if err != nil {
				errors++
			} else {
				files++
			}
		})
		b.Logf("files: %d, errors: %d", files, errors)
	}
}

func BenchmarkFastWalkGitRepoCallback(b *testing.B) {
	rootDir, err := ioutil.TempDir(os.TempDir(), "GitLfsBenchFastWalkGitRepoCallback")
	if err != nil {
		assert.FailNow(b, "Unable to get temp dir: %v", err)
	}
	defer os.RemoveAll(rootDir)
	os.Chdir(rootDir)
	entries := createFastWalkInputData(1000, 5000)

	for i := 0; i < b.N; i++ {
		var files, errors int
		FastWalkGitRepo(entries[0], func(parentDir string, info os.FileInfo, err error) {
			if err != nil {
				errors++
			} else {
				files++
			}
		})

		b.Logf("files: %d, errors: %d", files, errors)
	}
}

func TestFastWalkGitRepo(t *testing.T) {
	rootDir, err := ioutil.TempDir(os.TempDir(), "GitLfsTestFastWalkGitRepo")
	if err != nil {
		assert.FailNow(t, "Unable to get temp dir: %v", err)
	}
	defer os.RemoveAll(rootDir)
	os.Chdir(rootDir)

	expectedEntries := createFastWalkInputData(3, 3)

	mainDir := expectedEntries[0]

	// Set up a git repo and add some ignored files / dirs
	subprocess.SimpleExec("git", "init", mainDir)
	ignored := []string{
		"filethatweignore.ign",
		"foldercontainingignored",
		"foldercontainingignored/notthisone.ign",
		"ignoredfolder",
		"ignoredfolder/file1.txt",
		"ignoredfolder/file2.txt",
		"ignoredfrominside",
		"ignoredfrominside/thisisok.txt",
		"ignoredfrominside/thisisnot.txt",
		"ignoredfrominside/thisone",
		"ignoredfrominside/thisone/file1.txt",
	}
	for _, f := range ignored {
		fullPath := filepath.Join(mainDir, f)
		if len(filepath.Ext(f)) > 0 {
			ioutil.WriteFile(fullPath, []byte("TEST"), 0644)
		} else {
			os.MkdirAll(fullPath, 0755)
		}
	}
	// write root .gitignore
	rootGitIgnore := `
# ignore *.ign everywhere
*.ign
# ignore folder
ignoredfolder
`
	ioutil.WriteFile(filepath.Join(mainDir, ".gitignore"), []byte(rootGitIgnore), 0644)
	// Subfolder ignore; folder will show up but but subfolder 'thisone' won't
	subFolderIgnore := `
thisone
thisisnot.txt
`
	ioutil.WriteFile(filepath.Join(mainDir, "ignoredfrominside", ".gitignore"), []byte(subFolderIgnore), 0644)

	// This dir will be walked but content won't be
	expectedEntries = append(expectedEntries, filepath.Join(mainDir, "foldercontainingignored"))
	// This dir will be walked and some of its content but has its own gitignore
	expectedEntries = append(expectedEntries, filepath.Join(mainDir, "ignoredfrominside"))
	expectedEntries = append(expectedEntries, filepath.Join(mainDir, "ignoredfrominside", "thisisok.txt"))
	// Also gitignores
	expectedEntries = append(expectedEntries, filepath.Join(mainDir, ".gitignore"))
	expectedEntries = append(expectedEntries, filepath.Join(mainDir, "ignoredfrominside", ".gitignore"))
	// nothing else should be there

	gotEntries := make([]string, 0, 1000)
	gotErrors := make([]error, 0, 5)
	FastWalkGitRepo(mainDir, func(parent string, info os.FileInfo, err error) {
		if err != nil {
			gotErrors = append(gotErrors, err)
		} else {
			gotEntries = append(gotEntries, filepath.Join(parent, info.Name()))
		}
	})

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
		expectedEntries = append(expectedEntries, filepath.Clean(dir))
		if i >= 3 && i <= 5 {
			// Bulk test to ensure works with > 1 batch
			numFiles = largeFolder
		}
		for f := 0; f < numFiles; f++ {
			filename := filepath.Join(dir, fmt.Sprintf("file%d.txt", f))
			ioutil.WriteFile(filename, []byte("TEST"), 0644)
			expectedEntries = append(expectedEntries, filepath.Clean(filename))
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
			gotEntries = append(gotEntries, filepath.Join(o.ParentDir, o.Info.Name()))
		}
	}

	return gotEntries, gotErrors
}
