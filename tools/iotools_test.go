package tools

import (
	"bytes"
	"io"
	"testing"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/stretchr/testify/assert"
)

func TestCopyWithCallback(t *testing.T) {
	var calls int
	allReadSoFar := make([]int64, 0, 2)

	buf := bytes.NewBufferString("BOOYA")
	bufSize := buf.Len()

	cb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		calls++
		allReadSoFar = append(allReadSoFar, readSoFar)

		assert.EqualValues(t, bufSize, totalSize)

		return nil
	}

	n, err := CopyWithCallback(io.Discard, buf, int64(bufSize), cb)

	// The underlying bytes.Buffer should always return a nil error
	// when the last byte is read.
	assert.Nil(t, err)
	assert.EqualValues(t, bufSize, n)

	assert.Equal(t, 1, calls)
	assert.Len(t, allReadSoFar, 1)
	assert.EqualValues(t, bufSize, allReadSoFar[0])
}

func TestClosingByteReaderNopClose(t *testing.T) {
	buf := []byte{0x1}
	bufSize := len(buf)

	r := NewClosingByteReader(buf)

	// The underlying bytes.Reader has no Close() method and so reads
	// should still proceed.
	err := r.Close()

	assert.Nil(t, err)

	p := make([]byte, bufSize+1)

	n, err := r.Read(p)

	// The underlying bytes.Reader should always return
	// a nil error when the last byte is read.
	assert.Equal(t, bufSize, n)
	assert.Nil(t, err)
}

func TestRetriableReaderReturnsSuccessfulReads(t *testing.T) {
	r := NewRetriableReader(bytes.NewBuffer([]byte{0x1, 0x2, 0x3, 0x4}))

	var buf [4]byte
	n, err := r.Read(buf[:])

	assert.Nil(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte{0x1, 0x2, 0x3, 0x4}, buf[:])
}

func TestRetriableReaderReturnsEOFs(t *testing.T) {
	r := NewRetriableReader(bytes.NewBuffer([]byte{ /* empty */ }))

	var buf [1]byte
	n, err := r.Read(buf[:])

	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

func TestRetriableReaderMakesErrorsRetriable(t *testing.T) {
	expected := errors.New("example error")

	r := NewRetriableReader(&ErrReader{expected})

	var buf [1]byte
	n, err := r.Read(buf[:])

	assert.Equal(t, 0, n)
	assert.EqualError(t, err, "LFS: "+expected.Error())
	assert.True(t, errors.IsRetriableError(err))

}

func TestRetriableReaderDoesNotRewrap(t *testing.T) {
	// expected is already "retriable", as would be the case if the
	// underlying reader was a *RetriableReader itself.
	expected := errors.NewRetriableError(errors.New("example error"))

	r := NewRetriableReader(&ErrReader{expected})

	var buf [1]byte
	n, err := r.Read(buf[:])

	assert.Equal(t, 0, n)
	// errors.NewRetriableError wraps the given error with the prefix
	// message "LFS", so these two errors should be equal, indicating that
	// the RetriableReader did not re-wrap the error it received.
	assert.EqualError(t, err, expected.Error())
	assert.True(t, errors.IsRetriableError(err))

}

// ErrReader implements io.Reader and only returns errors.
type ErrReader struct {
	// err is the error that this reader will return.
	err error
}

// Read implements io.Reader#Read, and returns (0, e.err).
func (e *ErrReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}
