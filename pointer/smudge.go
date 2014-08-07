package pointer

import (
	"errors"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/gitmediaclient"
	"io"
	"os"
)

func Smudge(writer io.Writer, ptr *Pointer, cb gitmedia.CopyCallback) error {
	mediafile, err := gitmedia.LocalMediaPath(ptr.Oid)
	if err != nil {
		return err
	}

	if stat, statErr := os.Stat(mediafile); statErr != nil || stat == nil {
		err = downloadFile(writer, ptr, mediafile, cb)
	} else {
		err = readLocalFile(writer, ptr, mediafile, cb)
	}

	if err != nil {
		return &SmudgeError{ptr.Oid, mediafile, err.Error()}
	} else {
		return nil
	}
}

func downloadFile(writer io.Writer, ptr *Pointer, mediafile string, cb gitmedia.CopyCallback) error {
	reader, size, wErr := gitmediaclient.Get(mediafile)
	if wErr != nil {
		wErr.Errorf("Error downloading %s", mediafile)
		return wErr
	}
	defer reader.Close()

	if ptr.Size == 0 {
		ptr.Size = size
	}

	mediaWriter, err := newFile(mediafile, ptr.Oid)
	if err != nil {
		return errors.New("open: " + err.Error())
	}

	_, copyErr := io.Copy(mediaWriter, reader)
	closeErr := mediaWriter.Close()

	if copyErr != nil {
		return errors.New("write: " + copyErr.Error())
	}

	if closeErr != nil {
		return errors.New("close: " + closeErr.Error())
	}

	file, err := os.Open(mediaWriter.Path)
	if err != nil {
		return err
	}

	_, err = gitmedia.CopyWithCallback(writer, file, ptr.Size, cb)
	return err
}

func readLocalFile(writer io.Writer, ptr *Pointer, mediafile string, cb gitmedia.CopyCallback) error {
	reader, err := os.Open(mediafile)
	if err != nil {
		return err
	}
	defer reader.Close()

	if ptr.Size == 0 {
		if stat, _ := os.Stat(mediafile); stat != nil {
			ptr.Size = stat.Size()
		}
	}

	_, err = gitmedia.CopyWithCallback(writer, reader, ptr.Size, cb)
	return err
}

type SmudgeError struct {
	Oid          string
	Filename     string
	ErrorMessage string
}

func (e *SmudgeError) Error() string {
	return e.ErrorMessage
}
