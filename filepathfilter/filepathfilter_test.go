package filepathfilter

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatternMatch(t *testing.T) {
	assertPatternMatch(t, "filename.txt", "filename.txt")
	assertPatternMatch(t, "*.txt", "filename.txt")
	refutePatternMatch(t, "*.tx", "filename.txt")
	assertPatternMatch(t, "f*.txt", "filename.txt")
	refutePatternMatch(t, "g*.txt", "filename.txt")
	assertPatternMatch(t, "file*", "filename.txt")
	refutePatternMatch(t, "file", "filename.txt")

	// With no path separators, should match in subfolders
	assertPatternMatch(t, "*.txt", "sub/filename.txt")
	refutePatternMatch(t, "*.tx", "sub/filename.txt")
	assertPatternMatch(t, "f*.txt", "sub/filename.txt")
	refutePatternMatch(t, "g*.txt", "sub/filename.txt")
	assertPatternMatch(t, "file*", "sub/filename.txt")
	refutePatternMatch(t, "file", "sub/filename.txt")

	// matches only in subdir
	assertPatternMatch(t, "sub/*.txt", "sub/filename.txt")
	refutePatternMatch(t, "sub/*.txt", "top/sub/filename.txt")
	refutePatternMatch(t, "sub/*.txt", "sub/filename.dat")
	refutePatternMatch(t, "sub/*.txt", "other/filename.txt")

	// Needs wildcard for exact filename
	assertPatternMatch(t, "**/filename.txt", "sub/sub/sub/filename.txt")

	// Should not match dots to subparts
	refutePatternMatch(t, "*.ign", "sub/shouldignoreme.txt")

	// Path specific
	assertPatternMatch(t, "sub", "sub/")
	assertPatternMatch(t, "sub", "sub")
	assertPatternMatch(t, "sub", "sub/filename.txt")
	assertPatternMatch(t, "sub/", "sub/filename.txt")
	assertPatternMatch(t, "sub", "top/sub/filename.txt")
	assertPatternMatch(t, "sub/", "top/sub/filename.txt")
	assertPatternMatch(t, "sub", "top/sub/")
	assertPatternMatch(t, "sub", "top/sub")
	refutePatternMatch(t, "sub", "subfilename.txt")
	refutePatternMatch(t, "sub/", "subfilename.txt")

	// nested path
	assertPatternMatch(t, "top/sub", "top/sub/filename.txt")
	assertPatternMatch(t, "top/sub/", "top/sub/filename.txt")
	assertPatternMatch(t, "top/sub", "top/sub/")
	assertPatternMatch(t, "top/sub", "top/sub")
	assertPatternMatch(t, "top/sub", "root/top/sub/filename.txt")
	assertPatternMatch(t, "top/sub/", "root/top/sub/filename.txt")
	assertPatternMatch(t, "top/sub", "root/top/sub/")
	assertPatternMatch(t, "top/sub", "root/top/sub")

	// Absolute
	assertPatternMatch(t, "*.dat", "/path/to/sub/.git/test.dat")
	assertPatternMatch(t, "**/.git", "/path/to/sub/.git")

	// Match anything
	assertPatternMatch(t, ".", "path.txt")
	assertPatternMatch(t, "./", "path.txt")
	assertPatternMatch(t, ".\\", "path.txt")
}

func assertPatternMatch(t *testing.T, pattern, filename string) {
	assert.True(t, patternMatch(pattern, filename), "%q should match pattern %q", filename, pattern)
}

func refutePatternMatch(t *testing.T, pattern, filename string) {
	assert.False(t, patternMatch(pattern, filename), "%q should not match pattern %q", filename, pattern)
}

func patternMatch(pattern, filename string) bool {
	return NewPattern(pattern).Match(filepath.Clean(filename))
}

type filterTest struct {
	expectedResult bool
	includes       []string
	excludes       []string
}

func TestFilterAllows(t *testing.T) {
	cases := []filterTest{
		// Null case
		filterTest{true, nil, nil},
		// Inclusion
		filterTest{true, []string{"*.dat"}, nil},
		filterTest{true, []string{"file*.dat"}, nil},
		filterTest{true, []string{"file*"}, nil},
		filterTest{true, []string{"*name.dat"}, nil},
		filterTest{false, []string{"/*.dat"}, nil},
		filterTest{false, []string{"otherfolder/*.dat"}, nil},
		filterTest{false, []string{"*.nam"}, nil},
		filterTest{true, []string{"test/filename.dat"}, nil},
		filterTest{true, []string{"test/filename.dat"}, nil},
		filterTest{false, []string{"blank", "something", "foo"}, nil},
		filterTest{false, []string{"test/notfilename.dat"}, nil},
		filterTest{true, []string{"test"}, nil},
		filterTest{true, []string{"test/*"}, nil},
		filterTest{false, []string{"nottest"}, nil},
		filterTest{false, []string{"nottest/*"}, nil},
		filterTest{true, []string{"test/fil*"}, nil},
		filterTest{false, []string{"test/g*"}, nil},
		filterTest{true, []string{"tes*/*"}, nil},
		filterTest{true, []string{"[Tt]est/[Ff]ilename.dat"}, nil},
		// Exclusion
		filterTest{false, nil, []string{"*.dat"}},
		filterTest{false, nil, []string{"file*.dat"}},
		filterTest{false, nil, []string{"file*"}},
		filterTest{false, nil, []string{"*name.dat"}},
		filterTest{true, nil, []string{"/*.dat"}},
		filterTest{true, nil, []string{"otherfolder/*.dat"}},
		filterTest{false, nil, []string{"test/filename.dat"}},
		filterTest{false, nil, []string{"blank", "something", "test/filename.dat", "foo"}},
		filterTest{true, nil, []string{"blank", "something", "foo"}},
		filterTest{true, nil, []string{"test/notfilename.dat"}},
		filterTest{false, nil, []string{"test"}},
		filterTest{false, nil, []string{"test/*"}},
		filterTest{true, nil, []string{"nottest"}},
		filterTest{true, nil, []string{"nottest/*"}},
		filterTest{false, nil, []string{"test/fil*"}},
		filterTest{true, nil, []string{"test/g*"}},
		filterTest{false, nil, []string{"tes*/*"}},
		filterTest{false, nil, []string{"[Tt]est/[Ff]ilename.dat"}},

		// Both
		filterTest{true, []string{"test/filename.dat"}, []string{"test/notfilename.dat"}},
		filterTest{false, []string{"test"}, []string{"test/filename.dat"}},
		filterTest{true, []string{"test/*"}, []string{"test/notfile*"}},
		filterTest{false, []string{"test/*"}, []string{"test/file*"}},
		filterTest{false, []string{"another/*", "test/*"}, []string{"test/notfilename.dat", "test/filename.dat"}},
	}

	for _, c := range cases {
		result := New(c.includes, c.excludes).Allows("test/filename.dat")
		assert.Equal(t, c.expectedResult, result, "includes: %v excludes: %v", c.includes, c.excludes)
		if runtime.GOOS == "windows" {
			// also test with \ path separators, tolerate mixed separators
			for i, inc := range c.includes {
				c.includes[i] = strings.Replace(inc, "/", "\\", -1)
			}
			for i, ex := range c.excludes {
				c.excludes[i] = strings.Replace(ex, "/", "\\", -1)
			}
			assert.Equal(t, c.expectedResult, New(c.includes, c.excludes).Allows("test/filename.dat"), c)
		}
	}
}
