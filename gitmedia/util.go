package gitmedia

import (
	"io"
)

type CallbackWriter struct {
	C           func(int64, int64) error
	TotalSize   int64
	WrittenSize int64
	io.Writer
}

type CopyCallback func(int64, int64) error

func (w *CallbackWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)

	if n > 0 {
		w.WrittenSize += int64(n)
	}

	if err == nil && w.C != nil {
		err = w.C(w.TotalSize, w.WrittenSize)
	}

	return n, err
}

func CopyWithCallback(writer io.Writer, reader io.Reader, totalSize int64, cb CopyCallback) (int64, error) {
	if cb == nil {
		return io.Copy(writer, reader)
	}

	cbWriter := &CallbackWriter{
		C:         cb,
		TotalSize: totalSize,
		Writer:    writer,
	}
	return io.Copy(cbWriter, reader)
}
