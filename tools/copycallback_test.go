package tools

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/git-lfs/git-lfs/v3/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestBothCallbackReadersInvokeCallbackOnEagerEOF(t *testing.T) {
	var (
		calls               int
		actualTotalSize     int64
		actualReadSoFar     int64
		actualReadSinceLast int
	)

	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		calls++
		actualTotalSize = totalSize
		actualReadSoFar = readSoFar
		actualReadSinceLast = readSinceLast

		return nil
	}

	// We simulate a larger buffer which has not yet been fully read.
	buf := []byte{0x1}
	bufSize := len(buf)
	initialTotalSize := 3 * int64(bufSize)
	initialReadSize := initialTotalSize - int64(bufSize)

	r := &CallbackReader{
		C:         cb,
		TotalSize: initialTotalSize,
		ReadSize:  initialReadSize,
		Reader:    testutil.NewEagerEOFByteReader(buf),
	}
	br := NewBodyWithCallback(testutil.NewEagerEOFByteReader(buf), initialTotalSize, cb)
	br.readSize = initialReadSize

	p := make([]byte, bufSize+1)

	for _, reader := range []io.Reader{r, br} {
		t.Logf("testing with reader: %T", reader)

		n, err := reader.Read(p)

		assert.Equal(t, bufSize, n)
		assert.Nil(t, err)

		assert.Equal(t, 1, calls, "expected 1 call to callback, got %d", calls)
		assert.EqualValues(t, initialTotalSize, actualTotalSize)
		assert.EqualValues(t, initialTotalSize, actualReadSoFar)
		assert.Equal(t, bufSize, actualReadSinceLast)

		// Read again and check that no callback is made after last
		// byte has been read (since the simulated initial total
		// matched the simulated number of bytes read).
		calls = 0

		n, err = reader.Read(p)

		assert.Zero(t, n)
		assert.Equal(t, io.EOF, err)

		assert.Zero(t, calls, "expected no call to callback, got %d", calls)

		calls = 0
		actualTotalSize = 0
		actualReadSoFar = 0
		actualReadSinceLast = 0
	}
}

func TestBothCallbackReadersCountReads(t *testing.T) {
	var actualReadSoFar int64

	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		actualReadSoFar = readSoFar

		return nil
	}

	buf := []byte{0x1, 0x2, 0x3, 0x4}
	bufSize := len(buf)

	r := &CallbackReader{
		C:         cb,
		TotalSize: int64(bufSize),
		Reader:    bytes.NewReader(buf),
	}
	br := NewByteBodyWithCallback(buf, int64(bufSize), cb)

	p := make([]byte, 1)

	for _, reader := range []io.Reader{r, br} {
		t.Logf("testing with reader: %T", reader)

		for i := 1; i <= bufSize; i++ {
			n, err := reader.Read(p)

			// The underlying bytes.Reader should always return
			// a nil error when the last byte is read.
			assert.Equal(t, 1, n)
			assert.Nil(t, err)

			assert.EqualValues(t, i, actualReadSoFar)
		}

		actualReadSoFar = 0
	}
}

func TestBothCallbackReadersPreferCallbackErrorOverEOF(t *testing.T) {
	cbErr := errors.New("callback error")
	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		return cbErr
	}

	buf := []byte{0x1, 0x2}
	bufSize := len(buf)

	r := &CallbackReader{
		C:         cb,
		TotalSize: int64(bufSize),
		Reader:    testutil.NewEagerEOFByteReader(buf),
	}
	br := NewBodyWithCallback(testutil.NewEagerEOFByteReader(buf), int64(bufSize), cb)

	p := make([]byte, bufSize-1)

	for _, reader := range []io.Reader{r, br} {
		t.Logf("testing with reader: %T", reader)

		n, err := reader.Read(p)

		assert.Equal(t, bufSize-1, n)
		assert.Equal(t, cbErr, err)

		n, err = reader.Read(p)

		// We expect the EOF from EagerEOFByteReader's Read() method
		// to be replaced with the callback function's own error.
		assert.Equal(t, 1, n)
		assert.Equal(t, cbErr, err)

		// We expect no callback to be performed when no bytes are
		// available to be read, so no error should be returned.
		n, err = reader.Read(p)

		assert.Zero(t, n)
		assert.Equal(t, io.EOF, err)
	}
}

func TestBothCallbackReadersSkipCallbackAfterReadError(t *testing.T) {
	var calls int

	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		calls++

		return nil
	}

	buf := []byte{0x1, 0x2}
	bufSize := len(buf)

	r := &CallbackReader{
		C:         cb,
		TotalSize: int64(bufSize),
		Reader:    testutil.NewErrReader(io.ErrUnexpectedEOF),
	}
	br := NewBodyWithCallback(testutil.NewErrReader(io.ErrUnexpectedEOF), int64(bufSize), cb)

	p := make([]byte, bufSize-1)

	for _, reader := range []io.Reader{r, br} {
		t.Logf("testing with reader: %T", reader)

		n, err := reader.Read(p)

		assert.Zero(t, n)
		assert.Equal(t, io.ErrUnexpectedEOF, err)

		assert.Zero(t, calls, "expected no call to callback, got %d", calls)
	}
}

func TestBodyCallbackReaderUpdatesOffsetOnSeek(t *testing.T) {
	var calls int

	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		calls++

		return nil
	}

	buf := []byte{0x1, 0x2, 0x3, 0x4}
	bufSize := len(buf)

	br := NewByteBodyWithCallback(buf, int64(bufSize), cb)

	offset := 1
	br.Seek(int64(offset), io.SeekStart)
	assert.EqualValues(t, offset, br.readSize)

	offset++
	br.Seek(1, io.SeekCurrent)
	assert.EqualValues(t, offset, br.readSize)

	p := make([]byte, bufSize)

	n, err := br.Read(p)

	// The underlying bytes.Reader should always return
	// a nil error when the last byte is read.
	assert.Equal(t, bufSize-offset, n)
	assert.Nil(t, err)
	assert.Equal(t, buf[offset:], p[:bufSize-offset])

	assert.Equal(t, 1, calls, "expected 1 call to callback, got %d", calls)

	br.Seek(-1, io.SeekEnd)
	assert.EqualValues(t, bufSize-1, br.readSize)

	n, err = br.Read(p)

	assert.Equal(t, 1, n)
	assert.Nil(t, err)
	assert.Equal(t, buf[bufSize-1], p[0])

	assert.Equal(t, 2, calls, "expected 2 calls to callback, got %d", calls)
}
