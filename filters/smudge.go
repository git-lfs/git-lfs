package gitmediafilters

import (
	".."
	"io"
	"os"
)

type Smudger interface {
	Smudge(writer io.Writer, sha string) error
}

type localSmudger struct{}

type LocalSmudgeError struct {
	Sha          string
	Filename     string
	ErrorMessage string
}

func (e *LocalSmudgeError) Error() string {
	return e.ErrorMessage
}

func LocalSmudger() *localSmudger {
	return &localSmudger{}
}

func (s *localSmudger) Smudge(writer io.Writer, sha string) error {
	mediafile := gitmedia.LocalMediaPath(sha)
	file, err := os.Open(mediafile)
	if err != nil {
		return &LocalSmudgeError{sha, mediafile, err.Error()}
	}

	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		return &LocalSmudgeError{sha, mediafile, err.Error()}
	}

	return nil
}
