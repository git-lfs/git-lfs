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
