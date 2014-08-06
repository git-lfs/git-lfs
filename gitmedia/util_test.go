package gitmedia

import (
	"bytes"
	"github.com/bmizerany/assert"
	"io/ioutil"
	"testing"
)

func TestWriterWithCallback(t *testing.T) {
	buf := &bytes.Buffer{}

	called := 0
	calledWritten := make([]int64, 0, 2)

	writer := &CallbackWriter{
		TotalSize: 5,
		Writer:    buf,
		C: func(total int64, written int64) error {
			called += 1
			calledWritten = append(calledWritten, written)
			assert.Equal(t, 5, int(total))
			return nil
		},
	}

	writer.Write([]byte("BOO"))
	writer.Write([]byte("YA"))

	assert.Equal(t, 2, called)
	assert.Equal(t, 2, len(calledWritten))
	assert.Equal(t, 3, int(calledWritten[0]))
	assert.Equal(t, 5, int(calledWritten[1]))
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
