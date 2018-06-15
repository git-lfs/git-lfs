package lfs

import (
	"bytes"
	"testing"

	"github.com/git-lfs/git-lfs/tools"
	"github.com/stretchr/testify/assert"
)

func TestBodyWithCallback(t *testing.T) {
	called := 0
	calledRead := make([]int64, 0, 2)

	cb := func(total int64, read int64, current int) error {
		called += 1
		calledRead = append(calledRead, read)
		assert.Equal(t, 5, int(total))
		return nil
	}
	reader := tools.NewByteBodyWithCallback([]byte("BOOYA"), 5, cb)

	readBuf := make([]byte, 3)
	n, err := reader.Read(readBuf)
	assert.Nil(t, err)
	assert.Equal(t, "BOO", string(readBuf[0:n]))

	n, err = reader.Read(readBuf)
	assert.Nil(t, err)
	assert.Equal(t, "YA", string(readBuf[0:n]))

	assert.Equal(t, 2, called)
	assert.Len(t, calledRead, 2)
	assert.Equal(t, 3, int(calledRead[0]))
	assert.Equal(t, 5, int(calledRead[1]))
}

func TestReadWithCallback(t *testing.T) {
	called := 0
	calledRead := make([]int64, 0, 2)

	reader := &tools.CallbackReader{
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
	assert.Nil(t, err)
	assert.Equal(t, "BOO", string(readBuf[0:n]))

	n, err = reader.Read(readBuf)
	assert.Nil(t, err)
	assert.Equal(t, "YA", string(readBuf[0:n]))

	assert.Equal(t, 2, called)
	assert.Len(t, calledRead, 2)
	assert.Equal(t, 3, int(calledRead[0]))
	assert.Equal(t, 5, int(calledRead[1]))
}
