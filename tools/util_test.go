package tools

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	assert.Nil(t, err)
	assert.Equal(t, 5, int(n))

	assert.Equal(t, 1, called)
	assert.Len(t, calledWritten, 1)
	assert.Equal(t, 5, int(calledWritten[0]))
}

func TestMethodExists(t *testing.T) {
	// testing following methods exist in all platform.
	_, _ = CheckCloneFileSupported(os.TempDir())
	_, _ = CloneFile(io.Writer(nil), io.Reader(nil))
	_, _ = CloneFileByPath("", "")
}

func TestRenameNoReplaceDestExists(t *testing.T) {
	source, err := ioutil.TempFile("", "source")
	assert.NoError(t, err)
	assert.NoError(t, source.Close())
	defer os.Remove(source.Name())

	sourceData := []byte("source")
	assert.NoError(t, ioutil.WriteFile(source.Name(), sourceData, 0644))

	dest, err := ioutil.TempFile("", "dest")
	assert.NoError(t, err)
	defer os.Remove(dest.Name())
	assert.NoError(t, dest.Close())

	destData := []byte("dest")
	assert.NoError(t, ioutil.WriteFile(dest.Name(), destData, 0644))

	// Perform rename
	assert.Errorf(t, RenameNoReplace(source.Name(), dest.Name()), "file exists")

	sourceData2, err := ioutil.ReadFile(source.Name())
	assert.NoError(t, err)
	assert.Equal(t, sourceData, sourceData2)

	destData2, err := ioutil.ReadFile(dest.Name())
	assert.NoError(t, err)
	assert.Equal(t, destData, destData2)
}

func TestRenameNoReplace(t *testing.T) {
	source, err := ioutil.TempFile("", "source")
	assert.NoError(t, err)
	assert.NoError(t, source.Close())
	defer os.Remove(source.Name())

	sourceData := []byte("source")
	assert.NoError(t, ioutil.WriteFile(source.Name(), sourceData, 0644))

	dest, err := ioutil.TempFile("", "dest")
	assert.NoError(t, err)
	defer os.Remove(dest.Name())
	assert.NoError(t, dest.Close())

	// Remove destination file
	assert.NoError(t, os.Remove(dest.Name()))

	// Perform rename
	assert.NoError(t, RenameNoReplace(source.Name(), dest.Name()))

	_, err = os.Stat(source.Name())
	assert.True(t, os.IsNotExist(err))

	destData, err := ioutil.ReadFile(dest.Name())
	assert.NoError(t, err)
	assert.Equal(t, sourceData, destData)
}
