package metafile

import (
	"bytes"
	"errors"
	"io"
	"regexp"
)

var (
	MediaWarning  = []byte("# git-media\n")
	alpha         = "http://git-media.io/v/1"
	latest        = "http://git-media.io/v/2"
	oidType       = "sha256"
	alphaHeaderRE = regexp.MustCompile(`\A# (.*git-media|external)`)
	linebreak     = []byte("\n")
)

type Pointer struct {
	Version string
	Oid     string
	Size    int64
	OidType string
}

func NewPointer(oid string, size int64) *Pointer {
	return &Pointer{alpha, oid, size, oidType}
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
	buf := make([]byte, 200)
	written, err := reader.Read(buf)
	if err != nil {
		return nil, err
	}

	data := bytes.TrimSpace(buf[0:written])

	if alphaHeaderRE.Match(data) {
		return decodeAlpha(data)
	} else {
		return nil, errors.New("No INI decoder yet")
	}
}

func decodeAlpha(data []byte) (*Pointer, error) {
	lines := bytes.Split(data, linebreak)
	last := len(lines) - 1
	if last == 0 {
		return nil, errors.New("No sha in pointer file")
	}

	return &Pointer{alpha, string(lines[last]), 0, oidType}, nil
}
