package scanner

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/tools"
)

const (
	// stdoutBufSize is the size of the buffers given to a sub-rocess stdout
	stdoutBufSize = 16384

	// chanBufSize is the size of the channels used to pass data from one
	// sub-process to another.
	chanBufSize = 100
)

var (
	// z40 is a regular expression used to capture blobs that begin with 40
	// zeros.
	z40 = regexp.MustCompile(`\^?0{40}`)

	// ambiguousRegex is a regular expression used to determine whether or
	// not a call to `git rev-list ...` encountered ambiguous refs.
	ambiguousRegex = regexp.MustCompile(`warning: refname (.*) is ambiguous`)
)

// RevListShas list all revisions between `refLeft` and `refRight` according to
// the rules provided by *ScanRefsOptions.
func RevListShas(refLeft, refRight string, opt *ScanRefsOptions) (<-chan string, <-chan error, error) {
	return NewRevListScanner(opt).Scan(refLeft, refRight)
}

// RevListScanner implements the behavior of scanning over a list of revisions.
type RevListScanner struct {
	// ScanMode is an optional parameter which determines how the
	// RevListScanner should process the revisions given.
	ScanMode ScanningMode
	// RemoteName is the name of the remote to compare to.
	RemoteName string
	// SkipDeletedBlobs provides an option specifying whether or not we
	// should report revs whos blobs no longer exist at the remote, or at
	// the range's end.
	SkipDeletedBlobs bool

	nc *NameCache
}

// NewRevListScanner constructs a *RevListScanner using the given opts.
func NewRevListScanner(opts *ScanRefsOptions) *RevListScanner {
	return &RevListScanner{
		ScanMode:         opts.ScanMode,
		RemoteName:       opts.RemoteName,
		SkipDeletedBlobs: opts.SkipDeletedBlobs,

		nc: opts.nameCache,
	}
}

// Scan reports a channel of revisions, and a channel of errors encountered
// while scanning those revisions, over the given range from "left" to "right".
//
// If there was an error encountered before beginning to scan, it will be
// returned by itself, with two nil channels.
func (r *RevListScanner) Scan(left, right string) (<-chan string, <-chan error, error) {
	subArgs, stdin, err := r.refArgs(left, right)
	if err != nil {
		return nil, nil, err
	}

	args := make([]string, 2, len(subArgs)+2+1)
	args[0], args[1] = "rev-list", "--objects"
	args = append(args, append(subArgs, "--")...)

	cmd, err := startCommand("git", args...)
	if err != nil {
		return nil, nil, err
	}

	if len(stdin) > 0 {
		io.WriteString(cmd.Stdin, strings.Join(stdin, "\n"))
	}
	cmd.Stdin.Close()

	revs := make(chan string, chanBufSize)
	errchan := make(chan error, 5)

	go r.scanAndReport(cmd, revs, errchan)

	return revs, errchan, nil
}

// scanAndReport inspects the output of the given "cmd", and reports results to
// `revs` and `errs` respectively. Once the command has finished executing, the
// contents of its stderr are buffered into memory.
//
// If the exit-code of the `git rev-list` command is non-zero, a well-formatted
// error message (containing that command's stdout will be reported to `errs`.
// If the exit-code was zero, but the command reported ambiguous refs, then that
// too will be converted into an error.
//
// Once the command has finished processing and scanAndReport is no longer
// inspecting the output of that command, it closes both the `revs` and `errs`
// channels.
//
// scanAndReport runs in its own goroutine.
func (r *RevListScanner) scanAndReport(cmd *wrappedCmd, revs chan<- string, errs chan<- error) {
	for scanner := bufio.NewScanner(cmd.Stdout); scanner.Scan(); {
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 40 {
			continue
		}

		sha1 := line[:40]
		if len(line) > 40 {
			name := line[41:]

			r.nc.Cache(sha1, name)
		}

		revs <- sha1
	}

	stderr, _ := ioutil.ReadAll(cmd.Stderr)
	if err := cmd.Wait(); err != nil {
		errs <- fmt.Errorf("Error in `git rev-list --objects`: %v %s", err, stderr)
	} else {
		if match := ambiguousRegex.FindStringSubmatch(string(stderr)); len(match) > 0 {
			errs <- fmt.Errorf("Error: ref %s is ambiguous", match[1])
		}
	}

	close(revs)
	close(errs)
}

func (r *RevListScanner) refArgs(fromSha, toSha string) ([]string, []string, error) {
	var args, stdin []string

	switch r.ScanMode {
	case ScanRefsMode:
		if r.SkipDeletedBlobs {
			args = append(args, "--no-walk")
		} else {
			args = append(args, "--do-walk")
		}

		args = append(args, fromSha)
		if len(toSha) > 0 && !z40.MatchString(toSha) {
			args = append(args, toSha)
		}
	case ScanAllMode:
		args = append(args, "--all")
	case ScanLeftToRemoteMode:
		cachedRemoteRefs, _ := git.CachedRemoteRefs(r.RemoteName)
		actualRemoteRefs, _ := git.RemoteRefs(r.RemoteName)

		missingRefs := tools.NewStringSet()
		for _, cachedRef := range cachedRemoteRefs {
			var found bool
			for _, realRemoteRef := range actualRemoteRefs {
				if cachedRef == realRemoteRef {
					found = true
					break
				}
			}

			if !found {
				missingRefs.Add(cachedRef.Name)
			}
		}

		if len(missingRefs) > 0 {
			commits := make([]string, 1, len(cachedRemoteRefs)+1)
			commits[0] = fromSha

			for _, cachedRef := range cachedRemoteRefs {
				if missingRefs.Contains(cachedRef.Name) {
					continue
				}

				commits = append(commits, fmt.Sprintf("^%s", cachedRef.Sha))
			}

			stdin = commits
			args = append(args, "--stdin")
		} else {
			args = append(args, fromSha, "--not", "--remotes="+r.RemoteName)
		}
	default:
		return nil, nil, &UnknownScanModeError{r.ScanMode}
	}

	return args, stdin, nil
}

// UnknownScanModeError is an error given when an unrecognized scanning mode is
// given.
type UnknownScanModeError struct {
	// Mode is the mode that was unrecognized.
	Mode ScanningMode
}

// Error implements the `error` interface on `*UnknownScanModeError`.
func (e UnknownScanModeError) Error() string {
	return fmt.Sprintf("scanner: unknown scan type: %d", int(e.Mode))
}
