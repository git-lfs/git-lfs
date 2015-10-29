package lfs

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

func TestWriterWithCallback(t *testing.T) {
	called := 0
	calledRead := make([]int64, 0, 2)

	reader := &CallbackReader{
		TotalSize: 5,
		Reader:    bytes.NewBufferString("BOOYA"),
		C: func(total int64, read int64, current int) error {
			called += 1
			calledRead = append(calledRead, read)
			assert.Equal(t, 5, int(total))
			return nil
		},
	}

	readBuf := make([]byte, 3)
	n, err := reader.Read(readBuf)
	assert.Equal(t, nil, err)
	assert.Equal(t, "BOO", string(readBuf[0:n]))

	n, err = reader.Read(readBuf)
	assert.Equal(t, nil, err)
	assert.Equal(t, "YA", string(readBuf[0:n]))

	assert.Equal(t, 2, called)
	assert.Equal(t, 2, len(calledRead))
	assert.Equal(t, 3, int(calledRead[0]))
	assert.Equal(t, 5, int(calledRead[1]))
}

func TestCopyWithCallback(t *testing.T) {
	buf := bytes.NewBufferString("BOOYA")

	called := 0
	calledWritten := make([]int64, 0, 2)

	n, err := CopyWithCallback(ioutil.Discard, buf, 5, func(total int64, written int64, current int) error {
		called += 1
		calledWritten = append(calledWritten, written)
		assert.Equal(t, 5, int(total))
		return nil
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, 5, int(n))

	assert.Equal(t, 1, called)
	assert.Equal(t, 1, len(calledWritten))
	assert.Equal(t, 5, int(calledWritten[0]))
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
		assert.Equal(t, c.expectedResult, FilenamePassesIncludeExcludeFilter("test/filename.dat", c.includes, c.excludes), c)
		if IsWindows() {
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
