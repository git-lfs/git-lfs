package tools_test

import (
	"io/ioutil"
	"os"
	"runtime"
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

	assert.Nil(t, tools.SetFileWriteFlag(filename, false))
	// should turn off all write
	assert.EqualValues(t, 0444, getFileMode(filename))
	assert.Nil(t, tools.SetFileWriteFlag(filename, true))
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
		assert.Nil(t, tools.SetFileWriteFlag(filename, false))
		// should turn off all write but not execute
		assert.EqualValues(t, 0555, getFileMode(filename))
		assert.Nil(t, tools.SetFileWriteFlag(filename, true))
		// should only add back user write (on Mac/Linux)
		if runtime.GOOS == "windows" {
			assert.EqualValues(t, 0777, getFileMode(filename))
		} else {
			assert.EqualValues(t, 0755, getFileMode(filename))
		}

		assert.Nil(t, os.Chmod(filename, 0440))
		assert.Nil(t, tools.SetFileWriteFlag(filename, false))
		assert.EqualValues(t, 0440, getFileMode(filename))
		assert.Nil(t, tools.SetFileWriteFlag(filename, true))
		// should only add back user write
		assert.EqualValues(t, 0640, getFileMode(filename))
	}

}
