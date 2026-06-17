package tools

import (
	"io"
	"sync/atomic"
	"testing"

	"github.com/git-lfs/git-lfs/v3/internal/testutil"
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
		Reader:    testutil.NewEagerEOFByteReader(buf),
	}

	p := make([]byte, len(buf)+1)
	n, err := r.Read(p)

	assert.Equal(t, 1, n)
	assert.Nil(t, err)

	assert.EqualValues(t, 1, calls, "expected 1 call(s) to callback, got %d", calls)
	assert.EqualValues(t, 3, actualTotalSize)
	assert.EqualValues(t, 2+1, actualReadSoFar)
	assert.EqualValues(t, 1, actualReadSinceLast)
}

func TestBodyCallbackReaderCountsReads(t *testing.T) {
	br := NewByteBodyWithCallback([]byte{0x1, 0x2, 0x3, 0x4}, 4, nil)

	assert.EqualValues(t, 0, br.readSize)

	p := make([]byte, 8)
	n, err := br.Read(p)

	assert.Equal(t, 4, n)
	assert.Nil(t, err)
	assert.EqualValues(t, 4, br.readSize)
}

func TestBodyCallbackReaderUpdatesOffsetOnSeek(t *testing.T) {
	br := NewByteBodyWithCallback([]byte{0x1, 0x2, 0x3, 0x4}, 4, nil)

	br.Seek(1, io.SeekStart)
	assert.EqualValues(t, 1, br.readSize)

	br.Seek(1, io.SeekCurrent)
	assert.EqualValues(t, 2, br.readSize)

	br.Seek(-1, io.SeekEnd)
	assert.EqualValues(t, 3, br.readSize)
}
