package testutil

import (
	"errors"
	"io"
)

type ErrReader struct {
	err error
}

// Read() always returns the given error, akin to the ErrReader type from the
// "testing/iotest" package but with a complete io.ReadSeekCloser interface.
func (r *ErrReader) Read(p []byte) (int, error) {
	return 0, r.err
}

// Seek() is simply a no-op to complete the io.ReadSeekCloser interface.
func (r *ErrReader) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

// Close() is simply a no-op to complete the io.ReadSeekCloser interface.
func (r *ErrReader) Close() error {
	return nil
}

func NewErrReader(err error) *ErrReader {
	return &ErrReader{err}
}

type EagerEOFByteReader struct {
	b []byte
	i int
}

// Read() always returns io.EOF as early as possible, so it will
// return (n>0,io.EOF) instead of (n>0,nil) followed by (0,io.EOF).
func (r *EagerEOFByteReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.b[r.i:])
	r.i += n

	if r.i == len(r.b) {
		err = io.EOF
	}
	return
}

// Seek() could mirror the implementation of the Seek() method
// of bytes.Reader.  However, at present we do not need this method
// to be functional for any tests, so we can just stub it out instead.
func (r *EagerEOFByteReader) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("not implemented")
}

func (r *EagerEOFByteReader) Close() error {
	return nil
}

func NewEagerEOFByteReader(b []byte) *EagerEOFByteReader {
	return &EagerEOFByteReader{b: b}
}

type DeferredEOFByteReader struct {
	b     []byte
	i     int
	atEOF bool
}

// Read() always returns io.EOF as late as possible, so it will
// return (n>0,nil) followed by (0,io.EOF) instead of (n>0,io.EOF).
func (r *DeferredEOFByteReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.b[r.i:])
	r.i += n

	if r.i == len(r.b) {
		if r.atEOF {
			err = io.EOF
		} else {
			r.atEOF = true
		}
	}
	return
}

func NewDeferredEOFByteReader(b []byte) *DeferredEOFByteReader {
	return &DeferredEOFByteReader{b: b}
}
