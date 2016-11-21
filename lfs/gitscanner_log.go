package lfs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"time"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

// When scanning diffs e.g. parseLogOutputToPointers, which direction of diff to include
// data from, i.e. '+' or '-'. Depending on what you're scanning for either might be useful
type LogDiffDirection byte

const (
	LogDiffAdditions = LogDiffDirection('+') // include '+' diffs
	LogDiffDeletions = LogDiffDirection('-') // include '-' diffs
)

var (
	// Arguments to append to a git log call which will limit the output to
	// lfs changes and format the output suitable for parseLogOutput.. method(s)
	logLfsSearchArgs = []string{
		"-G", "oid sha256:", // only diffs which include an lfs file SHA change
		"-p",   // include diff so we can read the SHA
		"-U12", // Make sure diff context is always big enough to support 10 extension lines to get whole pointer
		`--format=lfs-commit-sha: %H %P`, // just a predictable commit header we can detect
	}
)

func scanUnpushed(remote string) (*PointerChannelWrapper, error) {
	logArgs := []string{"log",
		"--branches", "--tags", // include all locally referenced commits
		"--not"} // but exclude everything that comes after

	if len(remote) == 0 {
		logArgs = append(logArgs, "--remotes")
	} else {
		logArgs = append(logArgs, fmt.Sprintf("--remotes=%v", remote))
	}

	// Add standard search args to find lfs references
	logArgs = append(logArgs, logLfsSearchArgs...)

	cmd, err := startCommand("git", logArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	pchan := make(chan *WrappedPointer, chanBufSize)
	errchan := make(chan error, 1)

	go func() {
		parseLogOutputToPointers(cmd.Stdout, LogDiffAdditions, nil, nil, pchan)
		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			errchan <- fmt.Errorf("Error in git log: %v %v", err, string(stderr))
		}
		close(pchan)
		close(errchan)
	}()

	return NewPointerChannelWrapper(pchan, errchan), nil
}

// logPreviousVersions scans history for all previous versions of LFS pointers
// from 'since' up to (but not including) the final state at ref
func logPreviousSHAs(ref string, since time.Time) (*PointerChannelWrapper, error) {
	logArgs := []string{"log",
		fmt.Sprintf("--since=%v", git.FormatGitDate(since)),
	}
	// Add standard search args to find lfs references
	logArgs = append(logArgs, logLfsSearchArgs...)
	// ending at ref
	logArgs = append(logArgs, ref)

	cmd, err := startCommand("git", logArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	pchan := make(chan *WrappedPointer, chanBufSize)
	errchan := make(chan error, 1)

	// we pull out deletions, since we want the previous SHAs at commits in the range
	// this means we pick up all previous versions that could have been checked
	// out in the date range, not just if the commit which *introduced* them is in the range
	go func() {
		parseLogOutputToPointers(cmd.Stdout, LogDiffDeletions, nil, nil, pchan)
		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			errchan <- fmt.Errorf("Error in git log: %v %v", err, string(stderr))
		}
		close(pchan)
		close(errchan)
	}()

	return NewPointerChannelWrapper(pchan, errchan), nil
}

type logScanner struct {
	s            *bufio.Scanner
	dir          LogDiffDirection
	includePaths []string
	excludePaths []string
	pointer      *WrappedPointer

	pointerData         *bytes.Buffer
	currentFilename     string
	currentFileIncluded bool

	commitHeaderRegex    *regexp.Regexp
	fileHeaderRegex      *regexp.Regexp
	fileMergeHeaderRegex *regexp.Regexp
	pointerDataRegex     *regexp.Regexp
}

func newLogScanner(r io.Reader, dir LogDiffDirection, includePaths, excludePaths []string) *logScanner {
	return &logScanner{
		s:                    bufio.NewScanner(r),
		dir:                  dir,
		includePaths:         includePaths,
		excludePaths:         excludePaths,
		pointerData:          &bytes.Buffer{},
		currentFileIncluded:  true,
		commitHeaderRegex:    regexp.MustCompile(`^lfs-commit-sha: ([A-Fa-f0-9]{40})(?: ([A-Fa-f0-9]{40}))*`),
		fileHeaderRegex:      regexp.MustCompile(`diff --git a\/(.+?)\s+b\/(.+)`),
		fileMergeHeaderRegex: regexp.MustCompile(`diff --cc (.+)`),
		pointerDataRegex:     regexp.MustCompile(`^([\+\- ])(version https://git-lfs|oid sha256|size|ext-).*$`),
	}
}

func (s *logScanner) Pointer() *WrappedPointer {
	return s.pointer
}

func (s *logScanner) Err() error {
	return s.s.Err()
}

func (s *logScanner) Scan() bool {
	s.pointer = nil
	p, canScan := s.scan()
	s.pointer = p
	return canScan
}

// Utility func used at several points below (keep in narrow scope)
func (s *logScanner) finishLastPointer() *WrappedPointer {
	if s.pointerData.Len() > 0 {
		if s.currentFileIncluded {
			p, err := DecodePointer(s.pointerData)
			if err == nil {
				return &WrappedPointer{Name: s.currentFilename, Pointer: p}
			} else {
				tracerx.Printf("Unable to parse pointer from log: %v", err)
			}
		}
		s.pointerData.Reset()
	}
	return nil
}

// parseLogOutputToPointers parses log output formatted as per logLfsSearchArgs & return pointers
// log: a stream of output from git log with at least logLfsSearchArgs specified
// dir: whether to include results from + or - diffs
// includePaths, excludePaths: filter the results by filename
// results: a channel which will receive the pointers (caller must close)
func parseLogOutputToPointers(log io.Reader, dir LogDiffDirection,
	includePaths, excludePaths []string, results chan *WrappedPointer) {

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

	// Define regexes to capture commit & diff headers
	commitHeaderRegex := regexp.MustCompile(`^lfs-commit-sha: ([A-Fa-f0-9]{40})(?: ([A-Fa-f0-9]{40}))*`)
	fileHeaderRegex := regexp.MustCompile(`diff --git a\/(.+?)\s+b\/(.+)`)
	fileMergeHeaderRegex := regexp.MustCompile(`diff --cc (.+)`)
	pointerDataRegex := regexp.MustCompile(`^([\+\- ])(version https://git-lfs|oid sha256|size|ext-).*$`)
	var pointerData bytes.Buffer
	var currentFilename string
	currentFileIncluded := true

	// Utility func used at several points below (keep in narrow scope)
	finishLastPointer := func() {
		if pointerData.Len() > 0 {
			if currentFileIncluded {
				p, err := DecodePointer(&pointerData)
				if err == nil {
					results <- &WrappedPointer{Name: currentFilename, Pointer: p}
				} else {
					tracerx.Printf("Unable to parse pointer from log: %v", err)
				}
			}
			pointerData.Reset()
		}
	}

	scanner := bufio.NewScanner(log)
	for scanner.Scan() {
		line := scanner.Text()
		if match := commitHeaderRegex.FindStringSubmatch(line); match != nil {
			// Currently we're not pulling out commit groupings, but could if we wanted
			// This just acts as a delimiter for finishing a multiline pointer
			finishLastPointer()

		} else if match := fileHeaderRegex.FindStringSubmatch(line); match != nil {
			// Finding a regular file header
			finishLastPointer()
			// Pertinent file name depends on whether we're listening to additions or removals
			if dir == LogDiffAdditions {
				currentFilename = match[2]
			} else {
				currentFilename = match[1]
			}
			currentFileIncluded = tools.FilenamePassesIncludeExcludeFilter(currentFilename, includePaths, excludePaths)
		} else if match := fileMergeHeaderRegex.FindStringSubmatch(line); match != nil {
			// Git merge file header is a little different, only one file
			finishLastPointer()
			currentFilename = match[1]
			currentFileIncluded = tools.FilenamePassesIncludeExcludeFilter(currentFilename, includePaths, excludePaths)
		} else if currentFileIncluded {
			if match := pointerDataRegex.FindStringSubmatch(line); match != nil {
				// An LFS pointer data line
				// Include only the entirety of one side of the diff
				// -U3 will ensure we always get all of it, even if only
				// the SHA changed (version & size the same)
				changeType := match[1][0]
				// Always include unchanged context lines (normally just the version line)
				if LogDiffDirection(changeType) == dir || changeType == ' ' {
					// Must skip diff +/- marker
					pointerData.WriteString(line[1:])
					pointerData.WriteString("\n") // newline was stripped off by scanner
				}
			}
		}
	}
	// Final pointer if in progress
	finishLastPointer()
}

func (s *logScanner) scan() (*WrappedPointer, bool) {
	for s.s.Scan() {
		line := s.s.Text()

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
				s.currentFilename = match[2]
			} else {
				s.currentFilename = match[1]
			}
			s.currentFileIncluded = tools.FilenamePassesIncludeExcludeFilter(s.currentFilename, s.includePaths, s.excludePaths)

			if p != nil {
				return p, true
			}
		} else if match := s.fileMergeHeaderRegex.FindStringSubmatch(line); match != nil {
			// Git merge file header is a little different, only one file
			p := s.finishLastPointer()
			s.currentFilename = match[1]
			s.currentFileIncluded = tools.FilenamePassesIncludeExcludeFilter(s.currentFilename, s.includePaths, s.excludePaths)

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
	}

	return s.finishLastPointer(), false
}
