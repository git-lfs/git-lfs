package lfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb"
	"github.com/rubyist/tracerx"
)

func PointerSmudge(writer io.Writer, ptr *Pointer, workingfile string, cb CopyCallback) error {
	mediafile, err := LocalMediaPath(ptr.Oid)
	if err != nil {
		return err
	}

	stat, statErr := os.Stat(mediafile)
	if statErr == nil && stat != nil {
		fileSize := stat.Size()
		if fileSize == 0 || fileSize != ptr.Size {
			tracerx.Printf("Removing %s, size %d is invalid", mediafile, fileSize)
			os.RemoveAll(mediafile)
			stat = nil
		}
	}

	var wErr *WrappedError
	if statErr != nil || stat == nil {
		wErr = downloadFile(writer, ptr, workingfile, mediafile, cb)
	} else {
		wErr = readLocalFile(writer, ptr, mediafile, cb)
	}

	if wErr != nil {
		return &SmudgeError{ptr.Oid, mediafile, wErr}
	}

	return nil
}

func downloadFile(writer io.Writer, ptr *Pointer, workingfile, mediafile string, cb CopyCallback) *WrappedError {
	fmt.Fprintf(os.Stderr, "Downloading %s (%s)\n", workingfile, pb.FormatBytes(ptr.Size))
	reader, size, wErr := Download(filepath.Base(mediafile))
	if reader != nil {
		defer reader.Close()
	}

	if wErr != nil {
		wErr.Errorf("Error downloading %s.", mediafile)
		return wErr
	}

	if ptr.Size == 0 {
		ptr.Size = size
	}

	mediaFile, err := contentaddressable.NewFile(mediafile)
	if err != nil {
		return Errorf(err, "Error opening media file buffer.")
	}

	_, err = CopyWithCallback(mediaFile, reader, ptr.Size, cb)
	if err == nil {
		err = mediaFile.Accept()
	}
	mediaFile.Close()

	if err != nil {
		return Errorf(err, "Error buffering media file.")
	}

	return readLocalFile(writer, ptr, mediafile, nil)
}

func readLocalFile(writer io.Writer, ptr *Pointer, mediafile string, cb CopyCallback) *WrappedError {
	reader, err := os.Open(mediafile)
	if err != nil {
		return Errorf(err, "Error opening media file.")
	}
	defer reader.Close()

	if ptr.Size == 0 {
		if stat, _ := os.Stat(mediafile); stat != nil {
			ptr.Size = stat.Size()
		}
	}

	_, err = CopyWithCallback(writer, reader, ptr.Size, cb)
	return Errorf(err, "Error reading from media file.")
}

type SmudgeError struct {
	Oid      string
	Filename string
	*WrappedError
}
