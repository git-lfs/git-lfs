package pointer

import (
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

	if stat, err := os.Stat(mediafile); err != nil || stat == nil {
		err = downloadFile(writer, mediafile)
	} else {
		err = readLocalFile(writer, mediafile)
	}

	if err != nil {
		return &SmudgeError{oid, mediafile, err.Error()}
	} else {
		return nil
	}
}

func downloadFile(writer io.Writer, mediafile string) error {
	reader, err := gitmediaclient.Get(mediafile)
	if err != nil {
		return err
	}
	defer reader.Close()

	mediaWriter, err := os.Create(mediafile)
	if err != nil {
		return err
	}
	defer mediaWriter.Close()

	return copyFile(reader, writer, mediaWriter)
}

func readLocalFile(writer io.Writer, mediafile string) error {
	reader, err := os.Open(mediafile)
	if err != nil {
		return err
	}
	defer reader.Close()

	return copyFile(reader, writer)
}

func copyFile(reader io.ReadCloser, writers ...io.Writer) error {
	multiWriter := io.MultiWriter(writers...)

	_, err := io.Copy(multiWriter, reader)
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
