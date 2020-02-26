package tools

import (
	"bytes"
	"io"
	"os"
)

type CopyCallback func(totalSize int64, readSoFar int64, readSinceLast int) error

type BodyWithCallback struct {
	c         CopyCallback
	totalSize int64
	readSize  int64
	ReadSeekCloser
}

func NewByteBodyWithCallback(by []byte, totalSize int64, cb CopyCallback) *BodyWithCallback {
	return NewBodyWithCallback(NewByteBody(by), totalSize, cb)
}

func NewFileBodyWithCallback(f *os.File, totalSize int64, cb CopyCallback) *BodyWithCallback {
	return NewBodyWithCallback(NewFileBody(f), totalSize, cb)
}

func NewBodyWithCallback(body ReadSeekCloser, totalSize int64, cb CopyCallback) *BodyWithCallback {
	return &BodyWithCallback{
		c:              cb,
		totalSize:      totalSize,
		ReadSeekCloser: body,
	}
}

// Read wraps the underlying Reader's "Read" method. It also captures the number
// of bytes read, and calls the callback.
func (r *BodyWithCallback) Read(p []byte) (int, error) {
	n, err := r.ReadSeekCloser.Read(p)

	if n > 0 {
		r.readSize += int64(n)

		if (err == nil || err == io.EOF) && r.c != nil {
			err = r.c(r.totalSize, r.readSize, n)
		}
	}
	return n, err
}

// Seek wraps the underlying Seeker's "Seek" method, updating the number of
// bytes that have been consumed by this reader.
func (r *BodyWithCallback) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.readSize = offset
	case io.SeekCurrent:
		r.readSize += offset
	case io.SeekEnd:
		r.readSize = r.totalSize + offset
	}

	return r.ReadSeekCloser.Seek(offset, whence)
}

// ResetProgress calls the callback with a negative read size equal to the
// total number of bytes read so far, effectively "resetting" the progress.
func (r *BodyWithCallback) ResetProgress() error {
	return r.c(r.totalSize, r.readSize, -int(r.readSize))
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

		if (err == nil || err == io.EOF) && w.C != nil {
			err = w.C(w.TotalSize, w.ReadSize, n)
		}
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

func NewFileBody(f *os.File) ReadSeekCloser {
	return &closingFileReader{File: f}
}

type closingFileReader struct {
	*os.File
}

func (r *closingFileReader) Close() error {
	return nil
}
