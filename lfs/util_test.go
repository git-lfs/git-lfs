package lfs

import (
	"bytes"
	"io/ioutil"
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

func TestFilterIncludeExclude(t *testing.T) {

	// Inclusion
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, nil))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test/filename.dat"}, nil))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"blank", "something", "test/filename.dat", "foo"}, nil))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"blank", "something", "foo"}, nil))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test/notfilename.dat"}, nil))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test"}, nil))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test/*"}, nil))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"nottest"}, nil))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"nottest/*"}, nil))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test/fil*"}, nil))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test/g*"}, nil))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"tes*/*"}, nil))

	// Exclusion
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test/filename.dat"}))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"blank", "something", "test/filename.dat", "foo"}))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"blank", "something", "foo"}))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test/notfilename.dat"}))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test"}))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test/*"}))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"nottest"}))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"nottest/*"}))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test/fil*"}))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test/g*"}))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"tes*/*"}))

	// Both
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test/filename.dat"}, []string{"test/notfilename.dat"}))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test"}, []string{"test/filename.dat"}))
	assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test/*"}, []string{"test/notfile*"}))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test/*"}, []string{"test/file*"}))
	assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"another/*", "test/*"}, []string{"test/notfilename.dat", "test/filename.dat"}))

	if IsWindows() {
		// Extra tests because Windows git reports filenames with / separators
		// but we need to allow \ separators in include/exclude too
		// Can only test this ON Windows because of filepath behaviour
		// Inclusion
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, nil))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test\\filename.dat"}, nil))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"blank", "something", "test\\filename.dat", "foo"}, nil))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"blank", "something", "foo"}, nil))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test\\notfilename.dat"}, nil))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test"}, nil))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test\\*"}, nil))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"nottest"}, nil))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"nottest\\*"}, nil))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test\\fil*"}, nil))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test\\g*"}, nil))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"tes*\\*"}, nil))

		// Exclusion
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test\\filename.dat"}))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"blank", "something", "test\\filename.dat", "foo"}))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"blank", "something", "foo"}))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test\\notfilename.dat"}))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test"}))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test\\*"}))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"nottest"}))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"nottest\\*"}))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test\\fil*"}))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"test\\g*"}))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", nil, []string{"tes*\\*"}))

		// Both
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test\\filename.dat"}, []string{"test\\notfilename.dat"}))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test"}, []string{"test\\filename.dat"}))
		assert.Equal(t, true, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test\\*"}, []string{"test\\notfile*"}))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"test\\*"}, []string{"test\\file*"}))
		assert.Equal(t, false, FilenamePassesIncludeExcludeFilter("test/filename.dat", []string{"another\\*", "test\\*"}, []string{"test\\notfilename.dat", "test\\filename.dat"}))

	}
}
