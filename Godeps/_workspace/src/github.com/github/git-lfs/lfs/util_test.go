package lfs

import (
	"bytes"
	"github.com/bmizerany/assert"
	"io/ioutil"
	"testing"
)

func TestWriterWithCallback(t *testing.T) {
	called := 0
	calledRead := make([]int64, 0, 2)

	reader := &CallbackReader{
		TotalSize: 5,
		Reader:    bytes.NewBufferString("BOOYA"),
		C: func(total int64, read int64) error {
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

	n, err := CopyWithCallback(ioutil.Discard, buf, 5, func(total int64, written int64) error {
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
