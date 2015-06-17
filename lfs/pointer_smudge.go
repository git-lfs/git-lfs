package lfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/vendor/_nuts/github.com/cheggaaa/pb"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
	contentaddressable "github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/go-contentaddressable"
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
		sendApiEvent(apiEventSuccess)
		wErr = readLocalFile(writer, ptr, mediafile, cb)
	}

	if wErr != nil {
		return &SmudgeError{ptr.Oid, mediafile, wErr}
	}

	return nil
}

// PointerSmudgeObject uses a Pointer and ObjectResource to download the object to the
// media directory. It does not write the file to the working directory.
func PointerSmudgeObject(ptr *Pointer, obj *ObjectResource, cb CopyCallback) error {
	mediafile, err := LocalMediaPath(obj.Oid)
	if err != nil {
		return err
	}

	stat, statErr := os.Stat(mediafile)
	if statErr == nil && stat != nil {
		fileSize := stat.Size()
		if fileSize == 0 || fileSize != obj.Size {
			tracerx.Printf("Removing %s, size %d is invalid", mediafile, fileSize)
			os.RemoveAll(mediafile)
			stat = nil
		}
	}

	if statErr != nil || stat == nil {
		wErr := downloadObject(ptr, obj, mediafile, cb)

		if wErr != nil {
			sendApiEvent(apiEventFail)
			return &SmudgeError{obj.Oid, mediafile, wErr}
		}
	}

	sendApiEvent(apiEventSuccess)
	return nil
}

func downloadObject(ptr *Pointer, obj *ObjectResource, mediafile string, cb CopyCallback) *WrappedError {
	reader, size, wErr := DownloadObject(obj)
	if reader != nil {
		defer reader.Close()
	}

	// TODO this can be unified with the same code in downloadFile
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
