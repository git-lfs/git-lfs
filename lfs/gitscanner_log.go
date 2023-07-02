package lfs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/filepathfilter"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

// When scanning diffs with parseScannerLogOutput(), the direction of diff
// to include data from, i.e., '+' or '-'.  Depending on what you're scanning
// for either might be useful.
type LogDiffDirection byte

const (
	LogDiffAdditions = LogDiffDirection('+') // include '+' diffs
	LogDiffDeletions = LogDiffDirection('-') // include '-' diffs
)

var (
	// Arguments to append to a git log call which will limit the output to
	// lfs changes and format the output suitable for parseLogOutput.. method(s)
	logLfsSearchArgs = []string{
		"--no-ext-diff",
		"--no-textconv",
		"--color=never",
		"-G", "oid sha256:", // only diffs which include an lfs file SHA change
		"-p",                             // include diff so we can read the SHA
		"-U12",                           // Make sure diff context is always big enough to support 10 extension lines to get whole pointer
		`--format=lfs-commit-sha: %H %P`, // just a predictable commit header we can detect
	}
)

type gitscannerResult struct {
	Pointer *WrappedPointer
	Err     error
}

func scanUnpushed(cb GitScannerFoundPointer, remote string) error {
	logArgs := []string{
		"--branches", "--tags", // include all locally referenced commits
		"--not"} // but exclude everything that comes after

	if len(remote) == 0 {
		logArgs = append(logArgs, "--remotes")
	} else {
		logArgs = append(logArgs, fmt.Sprintf("--remotes=%v", remote))
	}

	// Add standard search args to find lfs references
	logArgs = append(logArgs, logLfsSearchArgs...)

	cmd, err := git.Log(logArgs...)
	if err != nil {
		return err
	}

	parseScannerLogOutput(cb, LogDiffAdditions, cmd, nil)
	return nil
}

func scanStashed(cb GitScannerFoundPointer) error {
	// Stashes are actually 2-3 commits, each containing one of:
	// 1. Working copy (WIP) modified files
	// 2. Index changes
	// 3. Untracked files (but only if "git stash -u" was used)
	// The first of these, the WIP commit, is a merge whose first parent
	// is HEAD and whose other parent(s) are commits 2 and 3 above.

	// We need to get the individual diff of each of these commits to
	// ensure we have all of the LFS objects referenced by the stash,
	// so a future "git stash pop" can restore them all.

	// First we get the list of SHAs of the WIP merge commits from the
	// reflog using "git log -g --format=%h refs/stash --".  Because
	// older Git versions (at least <=2.7) don't report merge parents in
	// the reflog, we can't extract the parent SHAs from "Merge:" lines
	// in the log; we can, however, use the "git log -m" option to force
	// an individual diff with the first merge parent in a second step.
	logArgs := []string{"-g", "--format=%h", "refs/stash", "--"}

	cmd, err := git.Log(logArgs...)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmd.Stdout)

	var stashMergeShas []string
	for scanner.Scan() {
		stashMergeSha := strings.TrimSpace(scanner.Text())
		stashMergeShas = append(stashMergeShas, fmt.Sprintf("%v^..%v", stashMergeSha, stashMergeSha))
	}
	if err := scanner.Err(); err != nil {
		errors.New(tr.Tr.Get("error while scanning `git log` for stashed refs: %v", err))
	}
	err = cmd.Wait()
	if err != nil {
		// Ignore this error, it really only happens when there's no refs/stash
		return nil
	}

	// We can use the log parser if we provide the -m and --first-parent
	// options to get the first WIP merge diff shown individually, then
	// no additional options to get the second index merge diff and
	// possible third untracked files merge diff in a subsequent step.
	stashMergeLogArgs := [][]string{{"-m", "--first-parent"}, {}}

	for _, logArgs := range stashMergeLogArgs {
		// Add standard search args to find lfs references
		logArgs = append(logArgs, logLfsSearchArgs...)

		logArgs = append(logArgs, stashMergeShas...)

		cmd, err = git.Log(logArgs...)
		if err != nil {
			return err
		}

		parseScannerLogOutput(cb, LogDiffAdditions, cmd, nil)
	}

	return nil
}

func parseScannerLogOutput(cb GitScannerFoundPointer, direction LogDiffDirection, cmd *subprocess.BufferedCmd, filter *filepathfilter.Filter) {
	ch := make(chan gitscannerResult, chanBufSize)

	go func() {
		scanner := newLogScanner(direction, cmd.Stdout)
		scanner.Filter = filter
		for scanner.Scan() {
			if p := scanner.Pointer(); p != nil {
				ch <- gitscannerResult{Pointer: p}
			}
		}
		if err := scanner.Err(); err != nil {
			ioutil.ReadAll(cmd.Stdout)
			ch <- gitscannerResult{Err: errors.New(tr.Tr.Get("error while scanning `git log`: %v", err))}
		}
		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			ch <- gitscannerResult{Err: errors.New(tr.Tr.Get("error in `git log`: %v %v", err, string(stderr)))}
		}
		close(ch)
	}()

	cmd.Stdin.Close()
	for result := range ch {
		cb(result.Pointer, result.Err)
	}
}

// logPreviousVersions scans history for all previous versions of LFS pointers
// from 'since' up to (but not including) the final state at ref
func logPreviousSHAs(cb GitScannerFoundPointer, ref string, filter *filepathfilter.Filter, since time.Time) error {
	logArgs := []string{
		fmt.Sprintf("--since=%v", git.FormatGitDate(since)),
	}
	// Add standard search args to find lfs references
	logArgs = append(logArgs, logLfsSearchArgs...)
	// ending at ref
	logArgs = append(logArgs, ref)

	cmd, err := git.Log(logArgs...)
	if err != nil {
		return err
	}

	parseScannerLogOutput(cb, LogDiffDeletions, cmd, filter)
	return nil
}

// logScanner parses log output formatted as per logLfsSearchArgs & returns
// pointers.
type logScanner struct {
	// Filter will ensure file paths matching the include patterns, or not matching
	// the exclude patterns are skipped.
	Filter *filepathfilter.Filter

	r       *bufio.Reader
	err     error
	dir     LogDiffDirection
	pointer *WrappedPointer

	pointerData         *bytes.Buffer
	currentFilename     string
	currentFileIncluded bool

	commitHeaderRegex    *regexp.Regexp
	fileHeaderRegex      *regexp.Regexp
	fileMergeHeaderRegex *regexp.Regexp
	pointerDataRegex     *regexp.Regexp
}

// dir: whether to include results from + or - diffs
// r: a stream of output from git log with at least logLfsSearchArgs specified
func newLogScanner(dir LogDiffDirection, r io.Reader) *logScanner {
	return &logScanner{
		r:                   bufio.NewReader(r),
		dir:                 dir,
		pointerData:         &bytes.Buffer{},
		currentFileIncluded: true,

		// no need to compile these regexes on every `git-lfs` call, just ones that
		// use the scanner.
		commitHeaderRegex:    regexp.MustCompile(fmt.Sprintf(`^lfs-commit-sha: (%s)(?: (%s))*`, git.ObjectIDRegex, git.ObjectIDRegex)),
		fileHeaderRegex:      regexp.MustCompile(`^diff --git a\/(.+?)\s+b\/(.+)`),
		fileMergeHeaderRegex: regexp.MustCompile(`^diff --cc (.+)`),
		pointerDataRegex:     regexp.MustCompile(`^([\+\- ])(version https://git-lfs|oid sha256|size|ext-).*$`),
	}
}

func (s *logScanner) Pointer() *WrappedPointer {
	return s.pointer
}

func (s *logScanner) Err() error {
	return s.err
}

func (s *logScanner) Scan() bool {
	s.pointer = nil
	p, canScan := s.scan()
	s.pointer = p
	return canScan
}

// Utility func used at several points below (keep in narrow scope)
func (s *logScanner) finishLastPointer() *WrappedPointer {
	if s.pointerData.Len() == 0 || !s.currentFileIncluded {
		return nil
	}

	p, err := DecodePointer(s.pointerData)
	s.pointerData.Reset()

	if err == nil {
		return &WrappedPointer{Name: s.currentFilename, Pointer: p}
	} else {
		tracerx.Printf("Unable to parse pointer from log: %v", err)
		return nil
	}
}

// For each commit we'll get something like this:
/*
	lfs-commit-sha: 60fde3d23553e10a55e2a32ed18c20f65edd91e7 e2eaf1c10b57da7b98eb5d722ec5912ddeb53ea1

	diff --git a/1D_Noise.png b/1D_Noise.png
	new file mode 100644
	index 0000000..2622b4a
	--- /dev/null
	+++ b/1D_Noise.png
	@@ -0,0 +1,3 @@
	+version https://git-lfs.github.com/spec/v1
	+oid sha256:f5d84da40ab1f6aa28df2b2bf1ade2cdcd4397133f903c12b4106641b10e1ed6
	+size 1289
*/
// There can be multiple diffs per commit (multiple binaries)
// Also when a binary is changed the diff will include a '-' line for the old SHA
func (s *logScanner) scan() (*WrappedPointer, bool) {
	for {
		line, err := s.r.ReadString('\n')

		if err != nil && err != io.EOF {
			s.err = err
			return nil, false
		}

		// remove trailing newline delimiter and optional single carriage return
		line = strings.TrimSuffix(strings.TrimRight(line, "\n"), "\r")

		if match := s.commitHeaderRegex.FindStringSubmatch(line); match != nil {
			// Currently we're not pulling out commit groupings, but could if we wanted
			// This just acts as a delimiter for finishing a multiline pointer
			if p := s.finishLastPointer(); p != nil {
				return p, true
			}
		} else if match := s.fileHeaderRegex.FindStringSubmatch(line); match != nil {
			// Finding a regular file header
			p := s.finishLastPointer()

			// Pertinent file name depends on whether we're listening to additions or removals
			if s.dir == LogDiffAdditions {
				s.setFilename(match[2])
			} else {
				s.setFilename(match[1])
			}

			if p != nil {
				return p, true
			}
		} else if match := s.fileMergeHeaderRegex.FindStringSubmatch(line); match != nil {
			// Git merge file header is a little different, only one file
			p := s.finishLastPointer()

			s.setFilename(match[1])

			if p != nil {
				return p, true
			}
		} else if s.currentFileIncluded {
			if match := s.pointerDataRegex.FindStringSubmatch(line); match != nil {
				// An LFS pointer data line
				// Include only the entirety of one side of the diff
				// -U3 will ensure we always get all of it, even if only
				// the SHA changed (version & size the same)
				changeType := match[1][0]

				// Always include unchanged context lines (normally just the version line)
				if LogDiffDirection(changeType) == s.dir || changeType == ' ' {
					// Must skip diff +/- marker
					s.pointerData.WriteString(line[1:])
					s.pointerData.WriteString("\n") // newline was stripped off by scanner
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	if p := s.finishLastPointer(); p != nil {
		return p, true
	}

	return nil, false
}

func (s *logScanner) setFilename(name string) {
	s.currentFilename = name
	s.currentFileIncluded = s.Filter.Allows(name)
}
