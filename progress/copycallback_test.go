package progress

import (
	"io"
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
		Reader:    &EOFReader{b: buf},
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

type EOFReader struct {
	b []byte
	i int
}

var _ io.Reader = (*EOFReader)(nil)

func (r *EOFReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.b[r.i:])
	r.i += n

	if r.i == len(r.b) {
		err = io.EOF
	}
	return
}

func TestEOFReaderReturnsEOFs(t *testing.T) {
	r := EOFReader{[]byte{0x1}, 0}

	p := make([]byte, 2)
	n, err := r.Read(p)

	assert.Equal(t, 1, n)
	assert.Equal(t, io.EOF, err)
}
