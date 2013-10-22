package gitmediafilters

import (
	".."
	"io"
	"os"
)

func Smudge(writer io.Writer, sha string) error {
	mediafile := gitmedia.LocalMediaPath(sha)
	file, err := os.Open(mediafile)
	if err != nil {
		return &SmudgeError{sha, mediafile, err.Error()}
	}

	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		return &SmudgeError{sha, mediafile, err.Error()}
	}

	return nil
}

type SmudgeError struct {
	Sha          string
	Filename     string
	ErrorMessage string
}

func (e *SmudgeError) Error() string {
	return e.ErrorMessage
}
