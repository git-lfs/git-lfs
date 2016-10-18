package scanner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
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
	z40            = regexp.MustCompile(`\^?0{40}`)
	ambiguousRegex = regexp.MustCompile(`warning: refname (.*) is ambiguous`)
)

type RevListScanner struct {
	ScanMode         ScanningMode
	RemoteName       string
	SkipDeletedBlobs bool

	revCache *RevCache
}

func NewRevListScanner(opts *ScanRefsOptions) *RevListScanner {
	return &RevListScanner{
		ScanMode:         opts.ScanMode,
		RemoteName:       opts.RemoteName,
		SkipDeletedBlobs: opts.SkipDeletedBlobs,

		// TODO(taylor): fix dependency on having mutable data in "opts"
		revCache: NewRevCache(opts.nameMap),
	}
}

func (r *RevListScanner) Scan(left, right string) (<-chan string, <-chan error, error) {
	// ~
	subArgs, stdin, err := r.refArgs(left)
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

func (r *RevListScanner) scanAndReport(cmd *wrappedCmd, revs chan<- string, errs chan<- error) {
	for scanner := bufio.NewScanner(cmd.Stdout); scanner.Scan(); {
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 40 {
			continue
		}

		sha1 := line[:40]
		if len(line) > 40 {
			name := line[41:]

			r.revCache.Cache(sha1, name)
		}

		revs <- sha1
	}

	var err error
	stderr, _ := ioutil.ReadAll(cmd.Stderr)
	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error in `git rev-list --objects`: %v %s", err, stderr)
	} else {
		if match := ambiguousRegex.FindStringSubmatch(string(stderr)); len(match) > 0 {
			err = fmt.Errorf("Error: ref %s is ambiguous", match[1])
		}
	}

	if err != nil {
		errs <- err
	}

	close(revs)
	close(errs)
}

func (r *RevListScanner) refArgs(fromSha string) ([]string, []string, error) {
	var args, stdin []string

	switch r.ScanMode {
	case ScanRefsMode:
		if r.SkipDeletedBlobs {
			args = append(args, "--no-walk")
		} else {
			args = append(args, "--do-walk")
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
				if cachedRef.MatchesNameAndType(realRemoteRef) {
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

type UnknownScanModeError struct {
	Mode ScanningMode
}

func (e UnknownScanModeError) Error() string {
	return fmt.Sprintf("scanner: unknown scan type: %d", int(e.Mode))
}

// RevListShas uses git rev-list to return the list of object sha1s
// for the given ref. If all is true, ref is ignored. It returns a
// channel from which sha1 strings can be read.
func RevListShas(refLeft, refRight string, opt *ScanRefsOptions) (<-chan string, <-chan error, error) {
	var stdin []string

	refArgs := []string{"rev-list", "--objects"}

	switch opt.ScanMode {
	case ScanRefsMode:
		if opt.SkipDeletedBlobs {
			refArgs = append(refArgs, "--no-walk")
		} else {
			refArgs = append(refArgs, "--do-walk")
		}

		refArgs = append(refArgs, refLeft)
		if refRight != "" && !z40.MatchString(refRight) {
			refArgs = append(refArgs, refRight)
		}
	case ScanAllMode:
		refArgs = append(refArgs, "--all")
	case ScanLeftToRemoteMode:
		args, commits := revListArgsRefVsRemote(refLeft, opt.RemoteName)
		refArgs = append(refArgs, args...)
		if len(commits) > 0 {
			stdin = commits
		}
	default:
		return nil, nil, errors.New("scanner: unknown scan type: " + strconv.Itoa(int(opt.ScanMode)))
	}

	// Use "--" at the end of the command to disambiguate arguments as refs,
	// so Git doesn't complain about ambiguity if you happen to also have a
	// file named "master".
	refArgs = append(refArgs, "--")

	cmd, err := startCommand("git", refArgs...)
	if err != nil {
		return nil, nil, err
	}

	if len(stdin) > 0 {
		cmd.Stdin.Write([]byte(strings.Join(stdin, "\n")))
	}
	cmd.Stdin.Close()

	revs := make(chan string, chanBufSize)
	errchan := make(chan error, 5) // may be multiple errors

	go func() {
		for scanner := bufio.NewScanner(cmd.Stdout); scanner.Scan(); {
			line := strings.TrimSpace(scanner.Text())
			if len(line) < 40 {
				continue
			}

			sha1 := line[:40]
			if len(line) > 40 {
				opt.SetName(sha1, line[41:len(line)])
			}

			revs <- sha1
		}

		// TODO(taylor): move this logic into an `*exec.Command` wrapper
		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		if err := cmd.Wait(); err != nil {
			errchan <- fmt.Errorf("Error in git rev-list --objects: %v %v", err, string(stderr))
		} else {
			// Special case detection of ambiguous refs; lower level commands like
			// git rev-list do not return non-zero exit codes in this case, just warn
			ambiguousRegex := regexp.MustCompile(`warning: refname (.*) is ambiguous`)
			if match := ambiguousRegex.FindStringSubmatch(string(stderr)); match != nil {
				// Promote to fatal & exit
				errchan <- fmt.Errorf("Error: ref %s is ambiguous", match[1])
			}
		}

		close(revs)
		close(errchan)
	}()

	return revs, errchan, nil
}

// revListArgsRefVsRemote gets additional arguments needed to limit 'git
// rev-list' to just the changes in revTo that are also not on remoteName.
//
// Returns a slice of string command arguments, and a slice of string git
// commits to pass to `git rev-list` via STDIN.
func revListArgsRefVsRemote(refTo, remoteName string) ([]string, []string) {
	// We need to check that the locally cached versions of remote refs are still
	// present on the remote before we use them as a 'from' point. If the
	// server implements garbage collection and a remote branch had been deleted
	// since we last did 'git fetch --prune', then the objects in that branch may
	// have also been deleted on the server if unreferenced.
	// If some refs are missing on the remote, use a more explicit diff

	cachedRemoteRefs, _ := git.CachedRemoteRefs(remoteName)
	actualRemoteRefs, _ := git.RemoteRefs(remoteName)

	// Only check for missing refs on remote; if the ref is different it has moved
	// forward probably, and if not and the ref has changed to a non-descendant
	// (force push) then that will cause a re-evaluation in a subsequent command anyway
	missingRefs := tools.NewStringSet()
	for _, cachedRef := range cachedRemoteRefs {
		found := false
		for _, realRemoteRef := range actualRemoteRefs {
			if cachedRef.Type == realRemoteRef.Type && cachedRef.Name == realRemoteRef.Name {
				found = true
				break
			}
		}
		if !found {
			missingRefs.Add(cachedRef.Name)
		}
	}

	if len(missingRefs) > 0 {
		// Use only the non-missing refs as 'from' points
		commits := make([]string, 1, len(cachedRemoteRefs)+1)
		commits[0] = refTo
		for _, cachedRef := range cachedRemoteRefs {
			if !missingRefs.Contains(cachedRef.Name) {
				commits = append(commits, "^"+cachedRef.Sha)
			}
		}
		return []string{"--stdin"}, commits
	} else {
		// Safe to use cached
		return []string{refTo, "--not", "--remotes=" + remoteName}, nil
	}
}
