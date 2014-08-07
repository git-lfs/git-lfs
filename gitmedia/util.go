package gitmedia

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func CopyCallbackFile(event, filename string) (CopyCallback, *os.File, error) {
	cbFilename := os.Getenv("GIT_MEDIA_PROGRESS")
	if len(cbFilename) == 0 || len(filename) == 0 || len(event) == 0 {
		return nil, nil, nil
	}

	cbDir := filepath.Dir(cbFilename)
	if err := os.MkdirAll(cbDir, 0755); err != nil {
		return nil, nil, wrapProgressError(err, cbFilename)
	}

	file, err := os.OpenFile(cbFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, file, wrapProgressError(err, cbFilename)
	}

	var prevProgress int
	var progress int

	cb := CopyCallback(func(total int64, written int64) error {
		progress = 0
		if total > 0 {
			progress = int(float64(written) / float64(total) * 100)
		}

		if progress != prevProgress {
			_, err := file.Write([]byte(fmt.Sprintf("%s %d %s\n", event, progress, filename)))
			prevProgress = progress
			return wrapProgressError(err, cbFilename)
		}

		return nil
	})
	file.Write([]byte(fmt.Sprintf("%s 0 %s\n", event, filename)))

	return cb, file, nil
}

func wrapProgressError(err error, filename string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("Error writing Git Media progress to %s: %s", filename, err.Error())
}
