package pointer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	ini "github.com/glacjay/goini"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var (
	MediaWarning  = []byte("# git-media\n")
	alpha         = "http://git-media.io/v/1"
	latest        = "http://git-media.io/v/2"
	oidType       = "sha256"
	alphaHeaderRE = regexp.MustCompile(`\A# (.*git-media|external)`)
	template      = `[git-media]
version=%s
oid=sha256:%s
size=%d
`
)

type Pointer struct {
	Version string
	Oid     string
	Size    int64
	OidType string
}

func NewPointer(oid string, size int64) *Pointer {
	return &Pointer{latest, oid, size, oidType}
}

func (p *Pointer) Smudge(writer io.Writer) error {
	return Smudge(writer, p)
}

func (p *Pointer) Encode(writer io.Writer) (int, error) {
	return Encode(writer, p)
}

func Encode(writer io.Writer, pointer *Pointer) (int, error) {
	return writer.Write([]byte(fmt.Sprintf(template,
		latest, pointer.Oid, pointer.Size)))
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
		return decodeIni(data)
	}
}

func decodeIni(data []byte) (*Pointer, error) {
	dict, err := ini.LoadReader(bufio.NewReader(bytes.NewReader(data)))
	if err != nil {
		return nil, err
	}

	v, ok := dict.GetString("git-media", "version")
	if !ok || v != latest {
		if len(v) == 0 {
			v = "--"
		}

		return nil, errors.New("Invalid version: " + v)
	}

	oidValue, ok := dict.GetString("git-media", "oid")
	if !ok {
		return nil, errors.New("Invalid Oid")
	}

	oidParts := strings.SplitN(oidValue, ":", 2)
	if len(oidParts) != 2 {
		return nil, errors.New("Invalid Oid type in" + oidValue)
	}
	if oidParts[0] != oidType {
		return nil, errors.New("Invalid Oid type: " + oidParts[0])
	}
	oid := oidParts[1]

	var size int64
	sizeStr, ok := dict.GetString("git-media", "size")
	if !ok {
		return nil, errors.New("Invalid Oid")
	} else {
		size, err = strconv.ParseInt(sizeStr, 10, 0)
		if err != nil {
			return nil, errors.New("Invalid size: " + sizeStr)
		}
	}

	return NewPointer(oid, size), nil
}

func decodeAlpha(data []byte) (*Pointer, error) {
	lines := bytes.Split(data, []byte("\n"))
	last := len(lines) - 1
	if last == 0 {
		return nil, errors.New("No sha in pointer file")
	}

	return &Pointer{alpha, string(lines[last]), 0, oidType}, nil
}
