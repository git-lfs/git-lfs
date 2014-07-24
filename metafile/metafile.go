package metafile

import (
	"bytes"
	"errors"
	"io"
	"regexp"
)

var (
	MediaWarning = []byte("# git-media\n")
	alpha        = "http://git-media.io/v/1"
)

type Pointer struct {
	Version string
	Oid     string
	Size    int64
}

func NewPointer(oid string, size int64) *Pointer {
	return &Pointer{alpha, oid, size}
}

func Encode(writer io.Writer, pointer *Pointer) (int, error) {
	written, err := writer.Write(MediaWarning)
	if err != nil {
		return written, err
	}

	written2, err := writer.Write([]byte(pointer.Oid + "\n"))
	return written + written2, err
}

func Decode(reader io.Reader) (*Pointer, error) {
	buf := make([]byte, 100)
	written, err := reader.Read(buf)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(buf[0:written], []byte("\n"))
	matched, err := regexp.Match("# (.*git-media|external)", lines[0])
	if err != nil {
		return nil, err
	}

	if len(lines) < 2 {
		return nil, errors.New("No sha in meta file")
	}

	if matched {
		return &Pointer{alpha, string(lines[1]), 0}, nil
	}

	return nil, errors.New("Could not decode meta file")
}
