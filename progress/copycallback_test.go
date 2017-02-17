package progress

import (
	"bytes"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyCallbackReaderCallsCallbackUnderfilledBuffer(t *testing.T) {
	var (
		calls               uint32
		actualTotalSize     int64
		actualReadSoFar     int64
		actualReadSinceLast int
	)

	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		atomic.AddUint32(&calls, 1)

		actualTotalSize = totalSize
		actualReadSoFar = readSoFar
		actualReadSinceLast = readSinceLast

		return nil
	}

	buf := []byte{0x1}
	r := &CallbackReader{
		C:         cb,
		TotalSize: 3,
		ReadSize:  2,
		Reader:    bytes.NewReader(buf),
	}

	p := make([]byte, len(buf)+1)
	n, err := r.Read(p)

	assert.Equal(t, 1, n)
	assert.Nil(t, err)

	assert.EqualValues(t, 1, calls)
	assert.EqualValues(t, 3, actualTotalSize)
	assert.EqualValues(t, 2+1, actualReadSoFar)
	assert.EqualValues(t, 1, actualReadSinceLast)
}
