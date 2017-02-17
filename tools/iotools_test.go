package tools_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/stretchr/testify/assert"
)

func TestRetriableReaderReturnsSuccessfulReads(t *testing.T) {
	r := tools.NewRetriableReader(bytes.NewBuffer([]byte{0x1, 0x2, 0x3, 0x4}))

	var buf [4]byte
	n, err := r.Read(buf[:])

	assert.Nil(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte{0x1, 0x2, 0x3, 0x4}, buf[:])
}

func TestRetriableReaderReturnsEOFs(t *testing.T) {
	r := tools.NewRetriableReader(bytes.NewBuffer([]byte{ /* empty */ }))

	var buf [1]byte
	n, err := r.Read(buf[:])

	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

func TestRetriableReaderMakesErrorsRetriable(t *testing.T) {
	expected := errors.New("example error")

	r := tools.NewRetriableReader(&ErrReader{expected})

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

	r := tools.NewRetriableReader(&ErrReader{expected})

	var buf [1]byte
	n, err := r.Read(buf[:])

	assert.Equal(t, 0, n)
	// errors.NewRetriableError wraps the given error with the prefix
	// message "LFS", so these two errors should be equal, indicating that
	// the RetriableReader did not re-wrap the error it received.
	assert.EqualError(t, err, expected.Error())
	assert.True(t, errors.IsRetriableError(err))

}

func TestCountingReaderCountsReads(t *testing.T) {
	cr := tools.NewCountingReadSeekCloser(NopReadSeekCloser(bytes.NewReader(
		[]byte{0x1, 0x2, 0x3, 0x4},
	)), 0)

	assert.EqualValues(t, 0, cr.N())

	p := make([]byte, 8)
	n, err := cr.Read(p)

	assert.Equal(t, 4, n)
	assert.Nil(t, err)
	assert.EqualValues(t, 4, cr.N())
}

func TestCountingReaderPassesErrors(t *testing.T) {
	expected := errors.New("some err")

	cr := tools.NewCountingReadSeekCloser(NopReadSeekCloser(&ErrReader{expected}), -1)

	p := make([]byte, 4)
	n, err := cr.Read(p)

	assert.Equal(t, 0, n)
	assert.Equal(t, expected, err)
}

func TestCountingReaderUpdatesOffsetOnSeek(t *testing.T) {
	cr := tools.NewCountingReadSeekCloser(NopReadSeekCloser(bytes.NewReader(
		[]byte{0x1, 0x2, 0x3, 0x4},
	)), 4)

	cr.Seek(1, io.SeekStart)
	assert.EqualValues(t, 1, cr.N())

	cr.Seek(1, io.SeekCurrent)
	assert.EqualValues(t, 2, cr.N())

	cr.Seek(-1, io.SeekEnd)
	assert.EqualValues(t, 3, cr.N())
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

// Seek implements io.Seeker.Seek and returns (0, e.err).
func (e *ErrReader) Seek(offset int64, whence int) (int64, error) {
	return 0, e.err
}

type readSeekCloser struct {
	io.ReadSeeker
}

func NopReadSeekCloser(r io.ReadSeeker) tools.ReadSeekCloser {
	return &readSeekCloser{r}
}

func (rsc *readSeekCloser) Close() error { return nil }
