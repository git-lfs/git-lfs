package pointer

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/github/git-lfs/lfs"
	"github.com/technoweenie/go-contentaddressable"
	"io"
	"os"
	"path/filepath"
)

func Smudge(writer io.Writer, ptr *Pointer, workingfile string, cb lfs.CopyCallback) error {
	mediafile, err := lfs.LocalMediaPath(ptr.Oid)
	if err != nil {
		return err
	}

	var wErr *lfs.WrappedError
	if stat, statErr := os.Stat(mediafile); statErr != nil || stat == nil {
		wErr = downloadFile(writer, ptr, workingfile, mediafile, cb)
	} else {
		wErr = readLocalFile(writer, ptr, mediafile, cb)
	}

	if wErr != nil {
		return &SmudgeError{ptr.Oid, mediafile, wErr}
	} else {
		return nil
	}
}

func downloadFile(writer io.Writer, ptr *Pointer, workingfile, mediafile string, cb lfs.CopyCallback) *lfs.WrappedError {
	fmt.Fprintf(os.Stderr, "Downloading %s (%s)\n", workingfile, pb.FormatBytes(ptr.Size))
	reader, size, wErr := lfs.Download(filepath.Base(mediafile))
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
		return lfs.Errorf(err, "Error opening media file buffer.")
	}

	_, err = lfs.CopyWithCallback(mediaFile, reader, ptr.Size, cb)
	if err == nil {
		err = mediaFile.Accept()
	}
	mediaFile.Close()

	if err != nil {
		return lfs.Errorf(err, "Error buffering media file.")
	}

	return readLocalFile(writer, ptr, mediafile, nil)
}

func readLocalFile(writer io.Writer, ptr *Pointer, mediafile string, cb lfs.CopyCallback) *lfs.WrappedError {
	reader, err := os.Open(mediafile)
	if err != nil {
		return lfs.Errorf(err, "Error opening media file.")
	}
	defer reader.Close()

	if ptr.Size == 0 {
		if stat, _ := os.Stat(mediafile); stat != nil {
			ptr.Size = stat.Size()
		}
	}

	_, err = lfs.CopyWithCallback(writer, reader, ptr.Size, cb)
	return lfs.Errorf(err, "Error reading from media file.")
}

type SmudgeError struct {
	Oid      string
	Filename string
	*lfs.WrappedError
}
