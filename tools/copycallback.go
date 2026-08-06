package tools

import (
	"io"
	"os"
)

type CopyCallback func(totalSize int64, readSoFar int64, readSinceLast int) error

type callbackState struct {
	callback  CopyCallback
	totalSize int64
	readSize  int64
}

func (cs *callbackState) doCallback(readSinceLast int) error {
	if cs.callback == nil {
		return nil
	}

	return cs.callback(cs.totalSize, cs.readSize, readSinceLast)
}

func (cs *callbackState) doCallbackIfRequired(n int, err error) error {
	if cs.callback == nil || (err != nil && err != io.EOF) {
		return nil
	}

	if err == io.EOF && cs.readSize != cs.totalSize {
		// If the total size was initially unknown or incorrect,
		// make a final callback with the total amount of data read
		// as the total size.
		cs.totalSize = cs.readSize
	} else if n == 0 {
		return nil
	}

	return cs.doCallback(n)
}

func (cs *callbackState) read(r io.Reader, p []byte) (int, error) {
	n, err := r.Read(p)

	cs.readSize += int64(n)

	callbackErr := cs.doCallbackIfRequired(n, err)

	// We only perform a callback if Read() returned no error or EOF,
	// so if the callback returned an error, we can return it to the
	// caller without masking a non-repeatable error condition.
	if callbackErr != nil {
		err = callbackErr
	}

	return n, err
}

type BodyWithCallback struct {
	state callbackState
	ReadSeekCloser
}

func NewFileBodyWithCallback(f *os.File, totalSize int64, cb CopyCallback) *BodyWithCallback {
	return NewBodyWithCallback(newNopClosingFile(f), totalSize, cb)
}

func NewBodyWithCallback(body ReadSeekCloser, totalSize int64, cb CopyCallback) *BodyWithCallback {
	return &BodyWithCallback{
		state: callbackState{
			callback:  cb,
			totalSize: totalSize,
		},
		ReadSeekCloser: body,
	}
}

// Read wraps the underlying Reader's "Read" method. It also captures the number
// of bytes read, and calls the callback.
func (r *BodyWithCallback) Read(p []byte) (int, error) {
	return r.state.read(r.ReadSeekCloser, p)
}

// Seek wraps the underlying Seeker's "Seek" method, updating the number of
// bytes that have been consumed by this reader.
func (r *BodyWithCallback) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.state.readSize = offset
	case io.SeekCurrent:
		r.state.readSize += offset
	case io.SeekEnd:
		r.state.readSize = r.state.totalSize + offset
	}

	return r.ReadSeekCloser.Seek(offset, whence)
}

// ResetProgress calls the callback with a negative read size equal to the
// total number of bytes read so far, effectively "resetting" the progress.
func (r *BodyWithCallback) ResetProgress() error {
	return r.state.doCallback(-int(r.state.readSize))
}

type CallbackReader struct {
	state callbackState
	io.Reader
}

func (r *CallbackReader) Read(p []byte) (int, error) {
	return r.state.read(r.Reader, p)
}

func NewCallbackReader(r io.Reader, totalSize int64, cb CopyCallback) *CallbackReader {
	return &CallbackReader{
		state: callbackState{
			callback:  cb,
			totalSize: totalSize,
		},
		Reader: r,
	}
}

// prevent import cycle
type ReadSeekCloser interface {
	io.Seeker
	io.ReadCloser
}

func newNopClosingFile(f *os.File) ReadSeekCloser {
	return &nopClosingFile{File: f}
}

type nopClosingFile struct {
	*os.File
}

func (r *nopClosingFile) Close() error {
	return nil
}
