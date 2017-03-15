package lfs

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
)

type Status rune

const (
	StatusAddition     Status = 'A'
	StatusCopy         Status = 'C'
	StatusDeletion     Status = 'D'
	StatusModification Status = 'M'
	StatusRename       Status = 'R'
	StatusTypeChange   Status = 'T'
	StatusUnmerged     Status = 'U'
	StatusUnknown      Status = 'X'
)

func (s Status) String() string {
	switch s {
	case StatusAddition:
		return "addition"
	case StatusCopy:
		return "copy"
	case StatusDeletion:
		return "deletion"
	case StatusModification:
		return "modification"
	case StatusRename:
		return "rename"
	case StatusTypeChange:
		return "change"
	case StatusUnmerged:
		return "unmerged"
	case StatusUnknown:
		return "unknown"
	}
	return "<unknown>"
}

type DiffIndexEntry struct {
	SrcMode     string
	DstMode     string
	SrcSha      string
	DstSha      string
	Status      Status
	StatusScore int
	SrcName     string
	DstName     string
}

type DiffIndexScanner struct {
	next *DiffIndexEntry
	err  error

	from *bufio.Scanner
}

func NewDiffIndexScanner(ref string, cached bool) (*DiffIndexScanner, *wrappedCmd, error) {
	cmd, err := startCommand("git", diffIndexCmdArgs(ref, cached)...)
	if err != nil {
		return nil, cmd, errors.Wrap(err, "diff-index")
	}

	if err = cmd.Stdin.Close(); err != nil {
		return nil, cmd, errors.Wrap(err, "diff-index: close")
	}

	return &DiffIndexScanner{
		from: bufio.NewScanner(cmd.Stdout),
	}, cmd, nil
}

func diffIndexCmdArgs(ref string, cached bool) []string {
	args := []string{"diff-index", "-M"}
	if cached {
		args = append(args, "--cached")
	}
	args = append(args, ref)

	return args
}

func (s *DiffIndexScanner) Scan() bool {
	if !s.prepareScan() {
		return false
	}

	s.next, s.err = s.scan(s.from.Text())

	return s.err == nil
}

func (s *DiffIndexScanner) Entry() *DiffIndexEntry { return s.next }
func (s *DiffIndexScanner) Err() error             { return s.err }

func (s *DiffIndexScanner) prepareScan() bool {
	s.next, s.err = nil, nil
	if !s.from.Scan() {
		s.err = s.from.Err()
		return false
	}

	return true
}

func (s *DiffIndexScanner) scan(line string) (*DiffIndexEntry, error) {
	parts := strings.Split(line, "\t")
	if len(parts) < 2 {
		return nil, errors.Errorf("invalid line: %s", line)
	}

	desc := strings.Fields(parts[0])
	if len(desc) < 5 {
		return nil, errors.Errorf("invalid description: %s", parts[0])
	}

	entry := &DiffIndexEntry{
		SrcMode: strings.TrimPrefix(desc[0], ":"),
		DstMode: desc[1],
		SrcSha:  desc[2],
		DstSha:  desc[3],
		Status:  Status(rune(desc[4][0])),
		SrcName: parts[1],
	}

	if score, err := strconv.Atoi(desc[4][1:]); err != nil {
		entry.StatusScore = score
	}

	if len(parts) > 2 {
		entry.DstName = parts[2]
	}

	return entry, nil
}
