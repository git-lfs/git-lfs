package filepathfilter

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatternMatch(t *testing.T) {
	for _, wildcard := range []string{"*", "*.*"} {
		assertPatternMatch(t, wildcard,
			"a",
			"a/",
			"a.a",
			"a/b",
			"a/b/",
			"a/b.b",
			"a/b/c",
			"a/b/c/",
			"a/b/c.c",
		)
	}

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
	refutePatternMatch(t, "sub/*.txt",
		"top/sub/filename.txt",
		"sub/filename.dat",
		"other/filename.txt",
	)

	// Needs wildcard for exact filename
	assertPatternMatch(t, "**/filename.txt", "sub/sub/sub/filename.txt")

	// Should not match dots to subparts
	refutePatternMatch(t, "*.ign", "sub/shouldignoreme.txt")

	// Path specific
	assertPatternMatch(t, "sub",
		"sub/",
		"sub",
		"sub/filename.txt",
		"top/sub/",
		"top/sub",
		"top/sub/filename.txt",
	)

	assertPatternMatch(t, "sub/", "sub/filename.txt", "top/sub/filename.txt")
	assertPatternMatch(t, "/sub", "sub/", "sub", "sub/filename.txt")
	assertPatternMatch(t, "/sub/", "sub/filename.txt")
	refutePatternMatch(t, "/sub", "subfilename.txt", "top/sub/", "top/sub", "top/sub/filename.txt")
	refutePatternMatch(t, "sub", "subfilename.txt")
	refutePatternMatch(t, "sub/", "subfilename.txt")
	refutePatternMatch(t, "/sub/", "subfilename.txt", "top/sub/filename.txt")

	// nested path
	assertPatternMatch(t, "top/sub",
		"top/sub/filename.txt",
		"top/sub/",
		"top/sub",
		"root/top/sub/filename.txt",
		"root/top/sub/",
		"root/top/sub",
	)
	assertPatternMatch(t, "top/sub/", "top/sub/filename.txt")
	assertPatternMatch(t, "top/sub/", "root/top/sub/filename.txt")

	assertPatternMatch(t, "/top/sub", "top/sub/", "top/sub", "top/sub/filename.txt")
	assertPatternMatch(t, "/top/sub/", "top/sub/filename.txt")

	refutePatternMatch(t, "top/sub", "top/subfilename.txt")
	refutePatternMatch(t, "top/sub/", "top/subfilename.txt")
	refutePatternMatch(t, "/top/sub",
		"top/subfilename.txt",
		"root/top/sub/filename.txt",
		"root/top/sub/",
		"root/top/sub",
	)

	refutePatternMatch(t, "/top/sub/",
		"root/top/sub/filename.txt",
		"top/subfilename.txt",
	)

	// Absolute
	assertPatternMatch(t, "*.dat", "/path/to/sub/.git/test.dat")
	assertPatternMatch(t, "**/.git", "/path/to/sub/.git")

	// Match anything
	assertPatternMatch(t, ".", "path.txt")
	assertPatternMatch(t, "./", "path.txt")
	assertPatternMatch(t, ".\\", "path.txt")
}

func assertPatternMatch(t *testing.T, pattern string, filenames ...string) {
	p := NewPattern(pattern)
	for _, filename := range toWindowsPaths(filenames) {
		assert.True(t, p.Match(filename), "%q should match pattern %q", filename, pattern)
	}
}

func refutePatternMatch(t *testing.T, pattern string, filenames ...string) {
	p := NewPattern(pattern)
	for _, filename := range toWindowsPaths(filenames) {
		assert.False(t, p.Match(filename), "%q should not match pattern %q", filename, pattern)
	}
}

type filterTest struct {
	expectedResult  bool
	expectedPattern string
	includes        []string
	excludes        []string
}

type filterPrefixTest struct {
	expected bool
	prefixes []string
	includes []string
	excludes []string
}

func (c *filterPrefixTest) Assert(t *testing.T) {
	f := New(c.platformIncludes(), c.platformExcludes())

	prefixes := c.prefixes
	if runtime.GOOS == "windows" {
		prefixes = toWindowsPaths(prefixes)
	}

	for _, prefix := range prefixes {
		assert.Equal(t, c.expected, f.HasPrefix(prefix),
			"expected=%v, prefix=%s", c.expected, prefix)
	}

}

func (c *filterPrefixTest) platformIncludes() []string {
	return toWindowsPaths(c.includes)
}

func (c *filterPrefixTest) platformExcludes() []string {
	return toWindowsPaths(c.excludes)
}

func toWindowsPaths(paths []string) []string {
	if runtime.GOOS != "windows" {
		return paths
	}

	out := make([]string, len(paths))
	for i, path := range paths {
		out[i] = strings.Replace(path, "/", "\\", -1)
	}

	return out
}

func TestFilterHasPrefix(t *testing.T) {
	prefixes := []string{"foo", "foo/", "foo/bar", "foo/bar/baz", "foo/bar/baz/"}
	for desc, c := range map[string]*filterPrefixTest{
		"empty filter":              {true, prefixes, nil, nil},
		"path prefix pattern":       {true, prefixes, []string{"/foo/bar/baz"}, nil},
		"path pattern":              {true, prefixes, []string{"foo/bar/baz"}, nil},
		"simple ext pattern":        {true, prefixes, []string{"*.dat"}, nil},
		"pathless wildcard pattern": {true, prefixes, []string{"foo*.dat"}, nil},
		"double wildcard pattern":   {true, prefixes, []string{"foo/**/baz"}, nil},
		"include other dir":         {false, prefixes, []string{"other"}, nil},

		"exclude pattern":                   {true, prefixes, nil, []string{"other"}},
		"exclude simple ext pattern":        {true, prefixes, nil, []string{"*.dat"}},
		"exclude pathless wildcard pattern": {true, prefixes, nil, []string{"foo*.dat"}},
	} {
		t.Run(desc, c.Assert)
	}

	prefixes = []string{"foo", "foo/", "foo/bar"}
	for desc, c := range map[string]*filterPrefixTest{
		"exclude path prefix pattern":     {true, prefixes, nil, []string{"/foo/bar/baz"}},
		"exclude path pattern":            {true, prefixes, nil, []string{"foo/bar/baz"}},
		"exclude double wildcard pattern": {true, prefixes, nil, []string{"foo/**/baz"}},
	} {
		t.Run(desc, c.Assert)
	}

	prefixes = []string{"foo/bar/baz", "foo/bar/baz/"}
	for desc, c := range map[string]*filterPrefixTest{
		"exclude path prefix pattern": {false, prefixes, nil, []string{"/foo/bar/baz"}},
		"exclude path pattern":        {false, prefixes, nil, []string{"foo/bar/baz"}},
	} {
		t.Run(desc, c.Assert)
	}

	prefixes = []string{"foo/bar/baz", "foo/test/baz"}
	for desc, c := range map[string]*filterPrefixTest{
		"exclude double wildcard pattern": {false, prefixes, nil, []string{"foo/**/baz"}},
	} {
		t.Run(desc, c.Assert)
	}
}

func TestFilterAllows(t *testing.T) {
	cases := []filterTest{
		// Null case
		filterTest{true, "", nil, nil},
		// Inclusion
		filterTest{true, "*.dat", []string{"*.dat"}, nil},
		filterTest{true, "file*.dat", []string{"file*.dat"}, nil},
		filterTest{true, "file*", []string{"file*"}, nil},
		filterTest{true, "*name.dat", []string{"*name.dat"}, nil},
		filterTest{false, "", []string{"/*.dat"}, nil},
		filterTest{false, "", []string{"otherfolder/*.dat"}, nil},
		filterTest{false, "", []string{"*.nam"}, nil},
		filterTest{true, "test/filename.dat", []string{"test/filename.dat"}, nil},
		filterTest{true, "test/filename.dat", []string{"test/filename.dat"}, nil},
		filterTest{false, "", []string{"blank", "something", "foo"}, nil},
		filterTest{false, "", []string{"test/notfilename.dat"}, nil},
		filterTest{true, "test", []string{"test"}, nil},
		filterTest{true, "test/*", []string{"test/*"}, nil},
		filterTest{false, "", []string{"nottest"}, nil},
		filterTest{false, "", []string{"nottest/*"}, nil},
		filterTest{true, "test/fil*", []string{"test/fil*"}, nil},
		filterTest{false, "", []string{"test/g*"}, nil},
		filterTest{true, "tes*/*", []string{"tes*/*"}, nil},
		filterTest{true, "[Tt]est/[Ff]ilename.dat", []string{"[Tt]est/[Ff]ilename.dat"}, nil},
		// Exclusion
		filterTest{false, "*.dat", nil, []string{"*.dat"}},
		filterTest{false, "file*.dat", nil, []string{"file*.dat"}},
		filterTest{false, "file*", nil, []string{"file*"}},
		filterTest{false, "*name.dat", nil, []string{"*name.dat"}},
		filterTest{true, "", nil, []string{"/*.dat"}},
		filterTest{true, "", nil, []string{"otherfolder/*.dat"}},
		filterTest{false, "test/filename.dat", nil, []string{"test/filename.dat"}},
		filterTest{false, "test/filename.dat", nil, []string{"blank", "something", "test/filename.dat", "foo"}},
		filterTest{true, "", nil, []string{"blank", "something", "foo"}},
		filterTest{true, "", nil, []string{"test/notfilename.dat"}},
		filterTest{false, "test", nil, []string{"test"}},
		filterTest{false, "test/*", nil, []string{"test/*"}},
		filterTest{true, "", nil, []string{"nottest"}},
		filterTest{true, "", nil, []string{"nottest/*"}},
		filterTest{false, "test/fil*", nil, []string{"test/fil*"}},
		filterTest{true, "", nil, []string{"test/g*"}},
		filterTest{false, "tes*/*", nil, []string{"tes*/*"}},
		filterTest{false, "[Tt]est/[Ff]ilename.dat", nil, []string{"[Tt]est/[Ff]ilename.dat"}},

		// // Both
		filterTest{true, "test/filename.dat", []string{"test/filename.dat"}, []string{"test/notfilename.dat"}},
		filterTest{false, "test/filename.dat", []string{"test"}, []string{"test/filename.dat"}},
		filterTest{true, "test/*", []string{"test/*"}, []string{"test/notfile*"}},
		filterTest{false, "test/file*", []string{"test/*"}, []string{"test/file*"}},
		filterTest{false, "test/filename.dat", []string{"another/*", "test/*"}, []string{"test/notfilename.dat", "test/filename.dat"}},
	}

	for _, c := range cases {
		if runtime.GOOS == "windows" {
			c.expectedPattern = strings.Replace(c.expectedPattern, "/", "\\", -1)
		}

		filter := New(c.includes, c.excludes)

		r1 := filter.Allows("test/filename.dat")
		assert.Equal(t, c.expectedResult, r1, "includes: %v excludes: %v", c.includes, c.excludes)

		if runtime.GOOS == "windows" {
			// also test with \ path separators, tolerate mixed separators
			for i, inc := range c.includes {
				c.includes[i] = strings.Replace(inc, "/", "\\", -1)
			}
			for i, ex := range c.excludes {
				c.excludes[i] = strings.Replace(ex, "/", "\\", -1)
			}

			filter = New(c.includes, c.excludes)

			r1 = filter.Allows("test/filename.dat")

			assert.Equal(t, c.expectedResult, r1, c)
		}
	}
}

func TestFilterReportsIncludePatterns(t *testing.T) {
	filter := New([]string{"*.foo", "*.bar"}, nil)

	assert.Equal(t, []string{"*.foo", "*.bar"}, filter.Include())
}

func TestFilterReportsExcludePatterns(t *testing.T) {
	filter := New(nil, []string{"*.baz", "*.quux"})

	assert.Equal(t, []string{"*.baz", "*.quux"}, filter.Exclude())
}
