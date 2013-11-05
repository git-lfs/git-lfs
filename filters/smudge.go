package gitmediafilters

import (
	".."
	"../client"
	"io"
	"os"
)

func Smudge(writer io.Writer, sha string) error { // stdout, sha
	mediafile := gitmedia.LocalMediaPath(sha)
	reader, err := gitmediaclient.Get(mediafile)
	if err != nil {
		return &SmudgeError{sha, mediafile, err.Error()}
	}

	defer reader.Close()

	mediaWriter, err := os.Create(mediafile)
	defer mediaWriter.Close()

	if err != nil {
		return &SmudgeError{sha, mediafile, err.Error()}
	}

	multiWriter := io.MultiWriter(writer, mediaWriter)

	_, err = io.Copy(multiWriter, reader)
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
