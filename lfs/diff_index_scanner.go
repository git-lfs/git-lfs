package lfs

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
)

// Status represents the status of a file that appears in the output of `git
// diff-index`.
//
// More information about each of its valid instances can be found:
// https://git-scm.com/docs/git-diff-index
type DiffIndexStatus rune

const (
	StatusAddition     DiffIndexStatus = 'A'
	StatusCopy         DiffIndexStatus = 'C'
	StatusDeletion     DiffIndexStatus = 'D'
	StatusModification DiffIndexStatus = 'M'
	StatusRename       DiffIndexStatus = 'R'
	StatusTypeChange   DiffIndexStatus = 'T'
	StatusUnmerged     DiffIndexStatus = 'U'
	StatusUnknown      DiffIndexStatus = 'X'
)

// String implements fmt.Stringer by returning a huamn-readable name for each
// status.
func (s DiffIndexStatus) String() string {
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

// DiffIndexEntry holds information about a single item in the results of a `git
// diff-index` command.
type DiffIndexEntry struct {
	// SrcMode is the file mode of the "src" file, stored as a string-based
	// octal.
	SrcMode string
	// DstMode is the file mode of the "dst" file, stored as a string-based
	// octal.
	DstMode string
	// SrcSha is the Git blob ID of the "src" file.
	SrcSha string
	// DstSha is the Git blob ID of the "dst" file.
	DstSha string
	// Status is the status of the file in the index.
	Status DiffIndexStatus
	// StatusScore is the optional "score" assosicated with a particular
	// status.
	StatusScore int
	// SrcName is the name of the file in it's "src" state as it appears in
	// the index.
	SrcName string
	// DstName is the name of the file in it's "dst" state as it appears in
	// the index.
	DstName string
}

// DiffIndexScanner scans the output of the `git diff-index` command.
type DiffIndexScanner struct {
	// next is the next entry scanned by the Scanner.
	next *DiffIndexEntry
	// err is any error that the Scanner encountered while scanning.
	err error

	// from is the underlying scanner, scanning the `git diff-index`
	// command's stdout.
	from *bufio.Scanner
}

// NewDiffIndexScanner initializes a new `DiffIndexScanner` scanning at the
// given ref, "ref".
//
// If "cache" is given, the DiffIndexScanner will scan for differences between
// the given ref and the index. If "cache" is _not_ given, DiffIndexScanner will
// scan for differences between the given ref and the currently checked out
// tree.
//
// If any error was encountered in starting the command or closing its `stdin`,
// that error will be returned immediately. Otherwise, a `*DiffIndexScanner`
// will be returned with a `nil` error.
func NewDiffIndexScanner(ref string, cached bool) (*DiffIndexScanner, error) {
	cmd, err := startCommand("git", diffIndexCmdArgs(ref, cached)...)
	if err != nil {
		return nil, errors.Wrap(err, "diff-index")
	}

	if err = cmd.Stdin.Close(); err != nil {
		return nil, errors.Wrap(err, "diff-index: close")
	}

	return &DiffIndexScanner{
		from: bufio.NewScanner(cmd.Stdout),
	}, nil
}

// diffIndexCmdArgs returns a string slice containing the arguments necessary
// to run the diff-index command.
func diffIndexCmdArgs(ref string, cached bool) []string {
	args := []string{"diff-index", "-M"}
	if cached {
		args = append(args, "--cached")
	}
	args = append(args, ref)

	return args
}

// Scan advances the scan line and yields either a new value for Entry(), or an
// Err(). It returns true or false, whether or not it can continue scanning for
// more entries.
func (s *DiffIndexScanner) Scan() bool {
	if !s.prepareScan() {
		return false
	}

	s.next, s.err = s.scan(s.from.Text())
	if s.err != nil {
		s.err = errors.Wrap(s.err, "diff-index scan")
	}

	return s.err == nil
}

// Entry returns the last entry that was Scan()'d by the DiffIndexScanner.
func (s *DiffIndexScanner) Entry() *DiffIndexEntry { return s.next }

// Entry returns the last error that was encountered by the DiffIndexScanner.
func (s *DiffIndexScanner) Err() error { return s.err }

// prepareScan clears out the results from the last Scan() loop, and advances
// the internal scanner to fetch a new line of Text().
func (s *DiffIndexScanner) prepareScan() bool {
	s.next, s.err = nil, nil
	if !s.from.Scan() {
		s.err = s.from.Err()
		return false
	}

	return true
}

// scan parses the given line and returns a `*DiffIndexEntry` or an error,
// depending on whether or not the parse was successful.
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
		Status:  DiffIndexStatus(rune(desc[4][0])),
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
