package filepathfilter

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
