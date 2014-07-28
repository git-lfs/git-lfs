package pointer

import (
	"errors"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/gitmediaclient"
	"io"
	"os"
)

func Smudge(writer io.Writer, oid string) error {
	mediafile, err := gitmedia.LocalMediaPath(oid)
	if err != nil {
		return err
	}

	if stat, statErr := os.Stat(mediafile); statErr != nil || stat == nil {
		err = downloadFile(writer, oid, mediafile)
	} else {
		err = readLocalFile(writer, mediafile)
	}

	if err != nil {
		return &SmudgeError{oid, mediafile, err.Error()}
	} else {
		return nil
	}
}

func downloadFile(writer io.Writer, oid, mediafile string) error {
	reader, err := gitmediaclient.Get(mediafile)
	if err != nil {
		return errors.New("client: " + err.Error())
	}
	defer reader.Close()

	mediaWriter, err := newFile(mediafile, oid)
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

	_, err = io.Copy(writer, file)
	return err
}

func readLocalFile(writer io.Writer, mediafile string) error {
	reader, err := os.Open(mediafile)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(writer, reader)
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
