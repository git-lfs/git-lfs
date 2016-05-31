package progress

import "io"

type CopyCallback func(totalSize int64, readSoFar int64, readSinceLast int) error

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
