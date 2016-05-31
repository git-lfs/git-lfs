package tools

import "io"

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
