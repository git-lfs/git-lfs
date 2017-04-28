package pointer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/github/git-lfs/lfs"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var (
	MediaWarning  = []byte("# git-media\n")
	alpha         = "http://git-media.io/v/1"
	beta          = "http://git-media.io/v/2"
	latest        = "https://git-lfs.github.com/spec/v1"
	oidType       = "sha256"
	alphaHeaderRE = regexp.MustCompile(`\A# (.*git-media|external)`)
	oidRE         = regexp.MustCompile(`\A[0-9a-fA-F]{64}`)
	template      = `version %s
oid sha256:%s
size %d
`
	matcherRE   = regexp.MustCompile("git-media|hawser|git-lfs")
	pointerKeys = []string{"version", "oid", "size"}
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

func (p *Pointer) Smudge(writer io.Writer, workingfile string, cb lfs.CopyCallback) error {
	return Smudge(writer, p, workingfile, cb)
}

func (p *Pointer) Encode(writer io.Writer) (int, error) {
	return Encode(writer, p)
}

func (p *Pointer) Encoded() string {
	return fmt.Sprintf(template, latest, p.Oid, p.Size)
}

func Encode(writer io.Writer, pointer *Pointer) (int, error) {
	return writer.Write([]byte(pointer.Encoded()))
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
		return decodeKV(data)
	}
}

func decodeKV(data []byte) (*Pointer, error) {
	parsed, err := decodeKVData(data)
	if err != nil {
		return nil, err
	}

	v, ok := parsed["version"]
	if !ok || (v != latest && v != beta) {
		if len(v) == 0 {
			v = "--"
		}
		return nil, errors.New("Invalid version: " + v)
	}

	oidValue, ok := parsed["oid"]
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
	sizeStr, ok := parsed["size"]
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

func decodeKVData(data []byte) (map[string]string, error) {
	m := make(map[string]string)

	if !matcherRE.Match(data) {
		return m, fmt.Errorf("Not a valid Git LFS pointer file.")
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	line := 0
	numKeys := len(pointerKeys)
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) == 0 {
			continue
		}

		parts := strings.SplitN(text, " ", 2)
		key := parts[0]

		if numKeys <= line {
			return m, fmt.Errorf("Extra line: %s", text)
		}

		if expected := pointerKeys[line]; key != expected {
			return m, fmt.Errorf("Expected key %s, got %s", expected, key)
		}

		line += 1
		if len(parts) < 2 {
			return m, fmt.Errorf("Error reading line %d: %s", line, text)
		}

		m[key] = parts[1]
	}

	return m, scanner.Err()
}

func decodeAlpha(data []byte) (*Pointer, error) {
	lines := bytes.Split(data, []byte("\n"))
	last := len(lines) - 1
	if last == 0 {
		return nil, errors.New("No OID in pointer file")
	}
	if !oidRE.Match(lines[last]) {
		return nil, errors.New("Invalid OID in pointer file")
	}

	return &Pointer{alpha, string(lines[last]), 0, oidType}, nil
}
