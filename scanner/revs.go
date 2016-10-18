package scanner

import (
	"bufio"
	"errors"
	"fmt"
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

var z40 = regexp.MustCompile(`\^?0{40}`)

// RevListShas uses git rev-list to return the list of object sha1s
// for the given ref. If all is true, ref is ignored. It returns a
// channel from which sha1 strings can be read.
func RevListShas(refLeft, refRight string, opt *ScanRefsOptions) (<-chan string, <-chan error, error) {
	refArgs := []string{"rev-list", "--objects"}
	var stdin []string
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
		scanner := bufio.NewScanner(cmd.Stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) < 40 {
				continue
			}

			sha1 := line[0:40]
			if len(line) > 40 {
				opt.SetName(sha1, line[41:len(line)])
			}
			revs <- sha1
		}

		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
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
