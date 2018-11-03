package commands

import (
	"io"
	"os"
)

type multiWriter struct {
	writer io.Writer
	fd     uintptr
}

func newMultiWriter(f *os.File, writers ...io.Writer) *multiWriter {
	return &multiWriter{
		writer: io.MultiWriter(append([]io.Writer{f}, writers...)...),
		fd:     f.Fd(),
	}
}

func (w *multiWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *multiWriter) Fd() uintptr {
	return w.fd
}
