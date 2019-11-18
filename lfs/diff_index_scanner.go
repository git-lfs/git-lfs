package lfs

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
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

// String implements fmt.Stringer by returning a human-readable name for each
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

// Format implements fmt.Formatter. If printed as "%+d", "%+s", or "%+v", the
// status will be written out as an English word: i.e., "addition", "copy",
// "deletion", etc.
//
// If the '+' flag is not given, the shorthand will be used instead: 'A', 'C',
// and 'D', respectively.
//
// If any other format verb is given, this function will panic().
func (s DiffIndexStatus) Format(state fmt.State, c rune) {
	switch c {
	case 'd', 's', 'v':
		if state.Flag('+') {
			state.Write([]byte(s.String()))
		} else {
			state.Write([]byte{byte(rune(s))})
		}
	default:
		panic(fmt.Sprintf("cannot format %v for DiffIndexStatus", c))
	}
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
	// StatusScore is the optional "score" associated with a particular
	// status.
	StatusScore int
	// SrcName is the name of the file in its "src" state as it appears in
	// the index.
	SrcName string
	// DstName is the name of the file in its "dst" state as it appears in
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
// If "refresh" is given, the DiffIndexScanner will refresh the index.  This is
// probably what you want in all cases except fsck, where invoking a filtering
// operation would be undesirable due to the possibility of corruption. It can
// also be disabled where another operation will have refreshed the index.
//
// If any error was encountered in starting the command or closing its `stdin`,
// that error will be returned immediately. Otherwise, a `*DiffIndexScanner`
// will be returned with a `nil` error.
func NewDiffIndexScanner(ref string, cached bool, refresh bool) (*DiffIndexScanner, error) {
	scanner, err := git.DiffIndex(ref, cached, refresh)
	if err != nil {
		return nil, err
	}
	return &DiffIndexScanner{
		from: scanner,
	}, nil
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
		s.err = errors.Wrap(s.err, "scan")
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
	// Format is:
	//   :100644 100644 c5b3d83a7542255ec7856487baa5e83d65b1624c 9e82ac1b514be060945392291b5b3108c22f6fe3 M foo.gif
	//   :<old mode> <new mode> <old sha1> <new sha1> <status>\t<file name>[\t<file name>]

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
