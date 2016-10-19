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
	z40            = regexp.MustCompile(`\^?0{40}`)
	ambiguousRegex = regexp.MustCompile(`warning: refname (.*) is ambiguous`)
)

func RevListShas(refLeft, refRight string, opt *ScanRefsOptions) (<-chan string, <-chan error, error) {
	return NewRevListScanner(opt).Scan(refLeft, refRight)
}

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
