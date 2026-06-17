package testutil

import (
	"io"
)

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

func NewEagerEOFByteReader(b []byte) *EagerEOFByteReader {
	return &EagerEOFByteReader{b: b}
}
