package gitmediafilters

import (
	".."
	"fmt"
	"io"
	"os"
)

// git media clean writer that outputs the large asset meta data to stdin, and
// deletes the temp file.  This is used directly if the large asset has been
// saved to the git media directory already.
type ExistingWriter struct {
	tempfile *os.File
	writer   io.Writer
}

func NewExistingWriter(cleaned *CleanedAsset, writer io.Writer) *ExistingWriter {
	return &ExistingWriter{cleaned.File, writer}
}

func (w *ExistingWriter) Write(buf []byte) (int, error) {
	return w.writer.Write(buf)
}

func (w *ExistingWriter) Close() error {
	return os.Remove(w.tempfile.Name())
}

type Writer struct {
	*ExistingWriter
}

// a git media clean writer that writes the large asset data to the local
// git media directory.  Also embeds an ExistingWriter.
func NewWriter(cleaned *CleanedAsset, writer io.Writer) *Writer {
	mediafile := gitmedia.LocalMediaPath(cleaned.Sha)

	if err := os.Rename(cleaned.File.Name(), mediafile); err != nil {
		fmt.Printf("Unable to move %s to %s\n", cleaned.File.Name(), mediafile)
		panic(err)
	}

	return &Writer{NewExistingWriter(cleaned, writer)}
}

func (w *Writer) Write(buf []byte) (int, error) {
	return w.writer.Write(buf)
}

func (w *Writer) Close() error {
	return w.ExistingWriter.Close()
}
