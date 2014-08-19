package pointer

import (
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/gitmediaclient"
	"github.com/technoweenie/go-contentaddressable"
	"io"
	"os"
)

func Smudge(writer io.Writer, ptr *Pointer, cb gitmedia.CopyCallback) error {
	mediafile, err := gitmedia.LocalMediaPath(ptr.Oid)
	if err != nil {
		return err
	}

	var wErr *gitmedia.WrappedError
	if stat, statErr := os.Stat(mediafile); statErr != nil || stat == nil {
		wErr = downloadFile(writer, ptr, mediafile, cb)
	} else {
		wErr = readLocalFile(writer, ptr, mediafile, cb)
	}

	if wErr != nil {
		return &SmudgeError{ptr.Oid, mediafile, wErr}
	} else {
		return nil
	}
}

func downloadFile(writer io.Writer, ptr *Pointer, mediafile string, cb gitmedia.CopyCallback) *gitmedia.WrappedError {
	reader, size, wErr := gitmediaclient.Get(mediafile)
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
		return gitmedia.Errorf(err, "Error opening media file buffer.")
	}

	_, err = gitmedia.CopyWithCallback(mediaFile, reader, ptr.Size, cb)
	if err == nil {
		err = mediaFile.Accept()
	}
	closeErr := mediaFile.Close()

	if err != nil {
		return gitmedia.Errorf(err, "Error buffering media file.")
	}

	if closeErr != nil {
		return gitmedia.Errorf(closeErr, "Error closing saved media file buffer.")
	}

	return readLocalFile(writer, ptr, mediafile, nil)
}

func readLocalFile(writer io.Writer, ptr *Pointer, mediafile string, cb gitmedia.CopyCallback) *gitmedia.WrappedError {
	reader, err := os.Open(mediafile)
	if err != nil {
		return gitmedia.Errorf(err, "Error opening media file.")
	}
	defer reader.Close()

	if ptr.Size == 0 {
		if stat, _ := os.Stat(mediafile); stat != nil {
			ptr.Size = stat.Size()
		}
	}

	_, err = gitmedia.CopyWithCallback(writer, reader, ptr.Size, cb)
	return gitmedia.Errorf(err, "Error reading from media file.")
}

type SmudgeError struct {
	Oid      string
	Filename string
	*gitmedia.WrappedError
}
