package gitmedia

import (
	"bytes"
	"io"
	"regexp"
)

var MediaWarning = []byte("# git-media\n")

func Encode(writer io.Writer, sha string) (int, error) {
	written, err := writer.Write(MediaWarning)
	if err != nil {
		return written, err
	}

	written2, err := writer.Write([]byte(sha + "\n"))
	return written + written2, err
}

func Decode(reader io.Reader) (string, error) {
	buf := make([]byte, 100)
	written, err := reader.Read(buf)
	if err != nil {
		return "", err
	}

	lines := bytes.Split(buf[0:written], []byte("\n"))
	matched, err := regexp.Match("# (.*git-media|external)", lines[0])
	if err != nil {
		return "", err
	}

	if matched {
		return string(lines[1]), nil
	}

	return "", nil // error?
}
