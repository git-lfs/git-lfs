package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
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

func TestFastWalkBasic(t *testing.T) {
	rootDir, err := ioutil.TempDir(os.TempDir(), "GitLfsTestFastWalkBasic")
	if err != nil {
		assert.FailNow(t, "Unable to get temp dir: %v", err)
	}
	defer os.RemoveAll(rootDir)
	os.Chdir(rootDir)

	expectedEntries := createFastWalkInputData(10, 160)

	walker := fastWalkWithExcludeFiles(expectedEntries[0], "")
	gotEntries, gotErrors := collectFastWalkResults(walker.ch)

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
		"foldercontainingignored/ignoredfolder",
		"foldercontainingignored/ignoredfolder/file1.txt",
		"foldercontainingignored/ignoredfolder/file2.txt",
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
		fullPath := join(mainDir, f)
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
	ioutil.WriteFile(join(mainDir, ".gitignore"), []byte(rootGitIgnore), 0644)
	// Subfolder ignore; folder will show up but but subfolder 'thisone' won't
	subFolderIgnore := `
thisone
thisisnot.txt
`
	ioutil.WriteFile(join(mainDir, "ignoredfrominside", ".gitignore"), []byte(subFolderIgnore), 0644)

	// This dir will be walked but content won't be
	expectedEntries = append(expectedEntries, join(mainDir, "foldercontainingignored"))
	// This dir will be walked and some of its content but has its own gitignore
	expectedEntries = append(expectedEntries, join(mainDir, "ignoredfrominside"))
	expectedEntries = append(expectedEntries, join(mainDir, "ignoredfrominside", "thisisok.txt"))
	// Also gitignores
	expectedEntries = append(expectedEntries, join(mainDir, ".gitignore"))
	expectedEntries = append(expectedEntries, join(mainDir, "ignoredfrominside", ".gitignore"))
	// nothing else should be there

	gotEntries := make([]string, 0, 1000)
	gotErrors := make([]error, 0, 5)
	FastWalkGitRepo(mainDir, func(parent string, info os.FileInfo, err error) {
		if err != nil {
			gotErrors = append(gotErrors, err)
		} else {
			if len(parent) == 0 {
				gotEntries = append(gotEntries, info.Name())
			} else {
				gotEntries = append(gotEntries, join(parent, info.Name()))
			}
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
