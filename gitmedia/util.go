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

func CopyWithCallback(writer io.Writer, reader io.Reader, totalSize int64, cb func(int64, int64) error) (int64, error) {
	cbWriter := &CallbackWriter{
		C:         cb,
		TotalSize: totalSize,
		Writer:    writer,
	}
	return io.Copy(cbWriter, reader)
}
