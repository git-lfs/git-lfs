package gitmedia

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type CallbackReader struct {
	C         CopyCallback
	TotalSize int64
	ReadSize  int64
	io.Reader
}

type CopyCallback func(int64, int64) error

func (w *CallbackReader) Read(p []byte) (int, error) {
	n, err := w.Reader.Read(p)

	if n > 0 {
		w.ReadSize += int64(n)
	}

	if err == nil && w.C != nil {
		err = w.C(w.TotalSize, w.ReadSize)
	}

	return n, err
}

func CopyWithCallback(writer io.Writer, reader io.Reader, totalSize int64, cb CopyCallback) (int64, error) {
	if cb == nil {
		return io.Copy(writer, reader)
	}

	cbReader := &CallbackReader{
		C:         cb,
		TotalSize: totalSize,
		Reader:    reader,
	}
	return io.Copy(writer, cbReader)
}

func CopyCallbackFile(event, filename string, index, totalFiles int) (CopyCallback, *os.File, error) {
	cbFilename := os.Getenv("GIT_MEDIA_PROGRESS")
	if len(cbFilename) == 0 || len(filename) == 0 || len(event) == 0 {
		return nil, nil, nil
	}

	cbDir := filepath.Dir(cbFilename)
	if err := os.MkdirAll(cbDir, 0755); err != nil {
		return nil, nil, wrapProgressError(err, event, cbFilename)
	}

	file, err := os.OpenFile(cbFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, file, wrapProgressError(err, event, cbFilename)
	}

	var prevWritten int64

	cb := CopyCallback(func(total int64, written int64) error {
		if written != prevWritten {
			_, err := file.Write([]byte(fmt.Sprintf("%s %d/%d %d/%d %s\n", event, index, totalFiles, written, total, filename)))
			file.Sync()
			prevWritten = written
			return wrapProgressError(err, event, cbFilename)
		}

		return nil
	})
	file.Write([]byte(fmt.Sprintf("%s %d/%d 0/0 %s\n", event, index, totalFiles, filename)))

	return cb, file, nil
}

func wrapProgressError(err error, event, filename string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("Error writing Git Media %s progress to %s: %s", event, filename, err.Error())
}
