package filters

import (
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/gitmediaclient"
	"io"
	"os"
)

func Smudge(writer io.Writer, sha string) error {
	mediafile, err := gitmedia.LocalMediaPath(sha)
	if err != nil {
		return err
	}

	if stat, err := os.Stat(mediafile); err != nil || stat == nil {
		reader, err := gitmediaclient.Get(mediafile)
		if err != nil {
			return &SmudgeError{sha, mediafile, err.Error()}
		}
		defer reader.Close()

		mediaWriter, err := os.Create(mediafile)
		if err != nil {
			return &SmudgeError{sha, mediafile, err.Error()}
		}
		defer mediaWriter.Close()

		if err := copyFile(reader, writer, mediaWriter); err != nil {
			return &SmudgeError{sha, mediafile, err.Error()}
		}
	} else {
		reader, err := os.Open(mediafile)
		if err != nil {
			return &SmudgeError{sha, mediafile, err.Error()}
		}
		defer reader.Close()

		if err := copyFile(reader, writer); err != nil {
			return &SmudgeError{sha, mediafile, err.Error()}
		}
	}

	return nil
}

func copyFile(reader io.ReadCloser, writers ...io.Writer) error {
	multiWriter := io.MultiWriter(writers...)

	_, err := io.Copy(multiWriter, reader)
	return err
}

type SmudgeError struct {
	Sha          string
	Filename     string
	ErrorMessage string
}

func (e *SmudgeError) Error() string {
	return e.ErrorMessage
}
