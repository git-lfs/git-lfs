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

func NewExistingWriter(cleaned *CleanedAsset, writer io.Writer) io.WriteCloser {
	return &ExistingWriter{cleaned.File, writer}
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

// a git media clean writer that writes the large asset data to the local
// git media directory.  Also embeds an ExistingWriter.
func NewWriter(cleaned *CleanedAsset, writer io.Writer) io.WriteCloser {
	mediafile := gitmedia.LocalMediaPath(cleaned.Sha)
	metafile := mediafile + ".txt"

	if err := os.Rename(cleaned.File.Name(), mediafile); err != nil {
		fmt.Printf("Unable to move %s to %s\n", cleaned.File.Name(), mediafile)
		panic(err)
	}

	file, err := os.Create(metafile)
	if err != nil {
		fmt.Printf("Unable to create meta data file: %s\n", metafile)
		panic(err)
	}

	return &Writer{
		metafile:       file,
		ExistingWriter: &ExistingWriter{cleaned.File, io.MultiWriter(writer, file)},
	}
}

func (w *Writer) Write(buf []byte) (int, error) {
	return w.writer.Write(buf)
}

func (w *Writer) Close() error {
	w.ExistingWriter.Close()
	return w.metafile.Close()
}
