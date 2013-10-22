package gitmediafilters

import (
	".."
	"../client"
	"io"
)

func Smudge(writer io.Writer, sha string) error {
	mediafile := gitmedia.LocalMediaPath(sha)
	reader, err := gitmediaclient.Get(mediafile)
	if err != nil {
		return &SmudgeError{sha, mediafile, err.Error()}
	}

	defer reader.Close()

	_, err = io.Copy(writer, reader)
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
