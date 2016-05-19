package tools

import (
	"io"

	"github.com/github/git-lfs/progress"
)

type readSeekCloserWrapper struct {
	readSeeker io.ReadSeeker
}

func (r *readSeekCloserWrapper) Read(p []byte) (n int, err error) {
	return r.readSeeker.Read(p)
}

func (r *readSeekCloserWrapper) Seek(offset int64, whence int) (int64, error) {
	return r.readSeeker.Seek(offset, whence)
}

func (r *readSeekCloserWrapper) Close() error {
	return nil
}

// NewReadSeekCloserWrapper wraps an io.ReadSeeker and implements a no-op Close() function
// to make it an io.ReadCloser
func NewReadSeekCloserWrapper(r io.ReadSeeker) io.ReadCloser {
	return &readSeekCloserWrapper{r}
}

// CopyWithCallback copies reader to writer while performing a progress callback
func CopyWithCallback(writer io.Writer, reader io.Reader, totalSize int64, cb progress.CopyCallback) (int64, error) {
	if success, _ := CloneFile(writer, reader); success {
		if cb != nil {
			cb(totalSize, totalSize, 0)
		}
		return totalSize, nil
	}
	if cb == nil {
		return io.Copy(writer, reader)
	}

	cbReader := &progress.CallbackReader{
		C:         cb,
		TotalSize: totalSize,
		Reader:    reader,
	}
	return io.Copy(writer, cbReader)
}
