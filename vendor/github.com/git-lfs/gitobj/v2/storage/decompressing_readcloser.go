package storage

import (
	"compress/zlib"
	"io"
)

// decompressingReadCloser wraps zlib.NewReader to ensure that both the zlib
// reader and its underlying type are closed.
type decompressingReadCloser struct {
	r  io.ReadCloser
	zr io.ReadCloser
}

// newDecompressingReadCloser creates a new wrapped zlib reader
func newDecompressingReadCloser(r io.ReadCloser) (io.ReadCloser, error) {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &decompressingReadCloser{r: r, zr: zr}, nil
}

// Read implements io.ReadCloser.
func (d *decompressingReadCloser) Read(b []byte) (int, error) {
	return d.zr.Read(b)
}

// Close implements io.ReadCloser.
func (d *decompressingReadCloser) Close() error {
	if err := d.zr.Close(); err != nil {
		return err
	}
	return d.r.Close()
}
