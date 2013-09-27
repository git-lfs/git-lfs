package gitmediaclean

import (
	".."
	"fmt"
	"io"
	"os"
)

type ExistingWriter struct {
	tempfile *os.File
	writer   io.Writer
}

func NewExistingWriter(asset *gitmedia.LargeAsset, tmp *os.File) io.WriteCloser {
	return &ExistingWriter{tmp, os.Stdout}
}

func (w *ExistingWriter) Write(buf []byte) (int, error) {
	return w.writer.Write(buf)
}

func (w *ExistingWriter) Close() error {
	return os.Remove(w.tempfile.Name())
}

type Writer struct {
	metafile *os.File
	*ExistingWriter
}

func NewWriter(asset *gitmedia.LargeAsset, tmp *os.File) io.WriteCloser {
	mediafile := gitmedia.LocalMediaPath(asset.SHA1)
	metafile := mediafile + ".json"

	if err := os.Rename(tmp.Name(), mediafile); err != nil {
		fmt.Printf("Unable to move %s to %s\n", tmp.Name(), mediafile)
		panic(err)
	}

	file, err := os.Create(metafile)
	if err != nil {
		fmt.Printf("Unable to create meta data file: %s\n", metafile)
		panic(err)
	}

	return &Writer{
		metafile:       file,
		ExistingWriter: &ExistingWriter{tmp, io.MultiWriter(os.Stdout, file)},
	}
}

func (w *Writer) Write(buf []byte) (int, error) {
	return w.writer.Write(buf)
}

func (w *Writer) Close() error {
	w.ExistingWriter.Close()
	return w.metafile.Close()
}
