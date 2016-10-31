package tools_test

import (
	"runtime"
	"strings"
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
