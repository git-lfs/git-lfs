package git

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
)

// An entry from ls-tree or rev-list including a blob sha and tree path
type TreeBlob struct {
	Oid      string
	Size     int64
	Mode     int32
	Filename string
}

type LsTreeScanner struct {
	s    *bufio.Scanner
	tree *TreeBlob
}

func NewLsTreeScanner(r io.Reader) *LsTreeScanner {
	s := bufio.NewScanner(r)
	s.Split(scanNullLines)
	return &LsTreeScanner{s: s}
}

func (s *LsTreeScanner) TreeBlob() *TreeBlob {
	return s.tree
}

func (s *LsTreeScanner) Err() error {
	return nil
}

func (s *LsTreeScanner) Scan() bool {
	t, hasNext := s.next()
	s.tree = t
	return hasNext
}

func (s *LsTreeScanner) next() (*TreeBlob, bool) {
	hasNext := s.s.Scan()
	line := s.s.Text()
	parts := strings.SplitN(line, "\t", 2)
	if len(parts) < 2 {
		return nil, hasNext
	}

	attrs := strings.SplitN(parts[0], " ", 4)
	if len(attrs) < 4 {
		return nil, hasNext
	}

	mode, err := strconv.ParseInt(strings.TrimSpace(attrs[0]), 8, 32)
	if err != nil {
		return nil, hasNext
	}

	if attrs[1] != "blob" {
		return nil, hasNext
	}

	sz, err := strconv.ParseInt(strings.TrimSpace(attrs[3]), 10, 64)
	if err != nil {
		return nil, hasNext
	}

	oid := attrs[2]
	filename := parts[1]
	return &TreeBlob{Oid: oid, Size: sz, Mode: int32(mode), Filename: filename}, hasNext
}

func scanNullLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\000'); i >= 0 {
		// We have a full null-terminated line.
		return i + 1, data[0:i], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
