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

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/subprocess"
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

	parseScannerLogOutput(cb, LogDiffAdditions, cmd)
	return nil
}

func scanStashed(cb GitScannerFoundPointer, s *GitScanner) error {

	// Stashes are actually 2-3 commits, each containing one of:
	// 1. Working copy modified files
	// 2. Index changes
	// 3. Untracked files, if -u was used
	// We need to get the diff of all 3 of these commits to ensure we have all
	// of the LFS objects necessary to pop the stash

	// First get the list of stashes
	// In recent version of git you can get parents directly from
	// this command, avoiding the intermediate "git show"
	// However older gits (at least <=2.7) don't report merge parents in the reflog
	// So we need to do it in 2 stages
	logArgs := []string{"-g", "--format=%h", "refs/stash", "--"}

	cmd, err := git.Log(logArgs...)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmd.Stdout)

	var allStashShas []string
	for scanner.Scan() {
		leafSha := strings.TrimSpace(scanner.Text())

		allStashShas = append(allStashShas, leafSha)

		// For each leaf, use "git show" to expand parents & thus get
		// all 2-3 shas involved in the stash
		// As mentioned newer gits could do this in the reflog output but not gteed
		showArgs := []string{"--quiet", "--format=%p", leafSha}
		showCmd, err := git.Show(showArgs...)
		if err != nil {
			return err
		}

		// gets the abbreviated parent hashes as :
		//      A B [C]
		// A = Parent commit of the stash (ignore, not part of the stash)
		// B = Index changes for the hash
		// C = Untracked files (optional, only present if -u)
		showScanner := bufio.NewScanner(showCmd.Stdout)

		for showScanner.Scan() {
			line := strings.TrimSpace(showScanner.Text())
			refs := strings.Split(line, " ")
			for i, ref := range refs {
				if i > 0 { // Extra merge parents
					allStashShas = append(allStashShas, ref)
				}
			}
		}
		err = showCmd.Wait()
		if err != nil {
			return err
		}

	}
	err = cmd.Wait()
	if err != nil {
		// Ignore this error, it really only happens when there's no refs/stash
		return nil
	}

	// Now we need to specifically use "git show" to parse results
	// We can't use "git log" because weirdly that omits the index changes
	// in the diff display, it collapses both into one diff and only shows the
	// final change (not a 3-way like show). Only "show" on all the shas
	// above displays them separately

	// The "leaf" stash actually shows both the index and working copy, like this:

	// -  oid sha256:8e1c163c2a04e25158962537cbff2540ded60d4612506a27bc04d059c7ae16dd
	//  - oid sha256:f2f84832183a0fca648c1ef49cfd32632b16b47ef5f17ac07dcfcb0ae00b86e5
	// -- size 16
	// +++oid sha256:b23f7e7314c5921e3e1cd87456d7867a51ccbe0c2c19ee4df64525c468d775df
	// +++size 30

	// The second "-" entry has a space prefix which shows this as a 3-way diff
	// However since we include all 2-3 commits explicitly in the git show,
	// We get this line as a "+" entry in the other commit
	// So we only need to care about the "+" entries
	// We can use the log parser, which can now handle 3-char +/- prefixes as well

	showArgs := logLfsSearchArgs
	showArgs = append(showArgs, allStashShas...)
	showArgs = append(showArgs, "--")

	cmd, err = git.Show(showArgs...)
	if err != nil {
		return err
	}

	parseScannerLogOutput(cb, LogDiffAdditions, cmd)
	return nil

}

func parseScannerLogOutput(cb GitScannerFoundPointer, direction LogDiffDirection, cmd *subprocess.BufferedCmd) {
	ch := make(chan gitscannerResult, chanBufSize)

	go func() {
		scanner := newLogScanner(direction, cmd.Stdout)
		for scanner.Scan() {
			if p := scanner.Pointer(); p != nil {
				ch <- gitscannerResult{Pointer: p}
			}
		}
		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			ch <- gitscannerResult{Err: fmt.Errorf("error in git log: %v %v", err, string(stderr))}
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
func logPreviousSHAs(cb GitScannerFoundPointer, ref string, since time.Time) error {
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

	parseScannerLogOutput(cb, LogDiffDeletions, cmd)
	return nil
}

func parseLogOutputToPointers(log io.Reader, dir LogDiffDirection,
	includePaths, excludePaths []string, results chan *WrappedPointer) {
	scanner := newLogScanner(dir, log)
	if len(includePaths)+len(excludePaths) > 0 {
		scanner.Filter = filepathfilter.New(includePaths, excludePaths)
	}
	for scanner.Scan() {
		if p := scanner.Pointer(); p != nil {
			results <- p
		}
	}
}

// logScanner parses log output formatted as per logLfsSearchArgs & returns
// pointers.
type logScanner struct {
	// Filter will ensure file paths matching the include patterns, or not matchin
	// the exclude patterns are skipped.
	Filter *filepathfilter.Filter

	s       *bufio.Scanner
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
		s:                   bufio.NewScanner(r),
		dir:                 dir,
		pointerData:         &bytes.Buffer{},
		currentFileIncluded: true,

		// no need to compile these regexes on every `git-lfs` call, just ones that
		// use the scanner.
		commitHeaderRegex:    regexp.MustCompile(fmt.Sprintf(`^lfs-commit-sha: (%s)(?: (%s))*`, git.ObjectIDRegex, git.ObjectIDRegex)),
		fileHeaderRegex:      regexp.MustCompile(`diff --git a\/(.+?)\s+b\/(.+)`),
		fileMergeHeaderRegex: regexp.MustCompile(`diff --cc (.+)`),
		// stash diff can have up to 3 +/- characters
		pointerDataRegex: regexp.MustCompile(`^([\+\- ]{1,3})(version https://git-lfs|oid sha256|size|ext-).*$`),
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

				if len(match[1]) > 1 {
					// A merge diff doesn't just have 1 +/- char, for example:
					// @@@@ -1,3 -1,3 -1,0 +1,3 @@@@
					//   +version https://git-lfs.github.com/spec/v1
					// -  oid sha256:8e1c163c2a04e25158962537cbff2540ded60d4612506a27bc04d059c7ae16dd
					//  - oid sha256:f2f84832183a0fca648c1ef49cfd32632b16b47ef5f17ac07dcfcb0ae00b86e5
					// -- size 16
					// +++oid sha256:b23f7e7314c5921e3e1cd87456d7867a51ccbe0c2c19ee4df64525c468d775df
					// +++size 30

					// To simplify, we're going to take "+" to mean "contains +"
					// and - will only register as the first character as before
					// We will ignore the " - " line entirely (changeIsBlank = false)
					if strings.Contains(match[1], "+") {
						changeType = '+'
					}
				}
				// merge lines can have 2-3 chars so can't just use changeType==' ' for blank
				changeIsBlank := len(strings.TrimSpace(match[1])) == 0

				// Always include unchanged context lines (normally just the version line)
				if LogDiffDirection(changeType) == s.dir || changeIsBlank {
					// Must skip diff +/- marker
					// can be 1-3 chars (3 for merge)
					s.pointerData.WriteString(line[len(match[1]):])
					s.pointerData.WriteString("\n") // newline was stripped off by scanner
				}
			}
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
