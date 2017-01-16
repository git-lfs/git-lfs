package progress

import (
	"bytes"
	"io"
)

type CopyCallback func(totalSize int64, readSoFar int64, readSinceLast int) error

type bodyWithCallback struct {
	c         CopyCallback
	totalSize int64
	readSize  int64
	ReadSeekCloser
}

func NewByteBodyWithCallback(by []byte, totalSize int64, cb CopyCallback) ReadSeekCloser {
	return NewBodyWithCallback(NewByteBody(by), totalSize, cb)
}

func NewBodyWithCallback(body ReadSeekCloser, totalSize int64, cb CopyCallback) ReadSeekCloser {
	return &bodyWithCallback{
		c:              cb,
		totalSize:      totalSize,
		ReadSeekCloser: body,
	}
}

func (r *bodyWithCallback) Read(p []byte) (int, error) {
	n, err := r.ReadSeekCloser.Read(p)

	if n > 0 {
		r.readSize += int64(n)
	}

	if err == nil && r.c != nil {
		err = r.c(r.totalSize, r.readSize, n)
	}

	return n, err
}

type CallbackReader struct {
	C         CopyCallback
	TotalSize int64
	ReadSize  int64
	io.Reader
}

func (w *CallbackReader) Read(p []byte) (int, error) {
	n, err := w.Reader.Read(p)

	if n > 0 {
		w.ReadSize += int64(n)
	}

	if err == nil && w.C != nil {
		err = w.C(w.TotalSize, w.ReadSize, n)
	}

	return n, err
}

// prevent import cycle
type ReadSeekCloser interface {
	io.Seeker
	io.ReadCloser
}

func NewByteBody(by []byte) ReadSeekCloser {
	return &closingByteReader{Reader: bytes.NewReader(by)}
}

type closingByteReader struct {
	*bytes.Reader
}

func (r *closingByteReader) Close() error {
	return nil
}
