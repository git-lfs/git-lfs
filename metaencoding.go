package gitmedia

import (
	"bytes"
	"io"
)

var MediaWarning = []byte("#!/usr/bin/env git media smudge\n# This is a placeholder for large media, please install git-media to retrieve content\n# It is also possible you did not have the media locally, run 'git media sync' to retrieve it\n")

func Encode(writer io.Writer, sha string) (int, error) {
	written, err := writer.Write(MediaWarning)
	if err != nil {
		return written, err
	}

	written2, err := writer.Write([]byte(sha))
	return written + written2, err
}

func Decode(reader io.Reader) (string, error) {
	buf := make([]byte, 1024)
	written, err := reader.Read(buf)
	if err != nil {
		return "", err
	}

	return lastNonEmpty(bytes.Split(buf[0:written], []byte("\n"))), nil
}

func lastNonEmpty(parts [][]byte) string {
	idx := len(parts)
	var part []byte
	for len(part) == 0 {
		idx -= 1
		part = parts[idx]
	}
	return string(part)
}
