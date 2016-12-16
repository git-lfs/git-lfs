package lfs

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

var z40 = regexp.MustCompile(`\^?0{40}`)

// scanRefsToChan takes a ref and returns a channel of WrappedPointer objects
// for all Git LFS pointers it finds for that ref.
// Reports unique oids once only, not multiple times if >1 file uses the same content
func scanRefsToChan(cb GitScannerCallback, refLeft, refRight string, opt *ScanRefsOptions) error {
	if opt == nil {
		panic("no scan ref options")
	}

	revs, err := revListShas(refLeft, refRight, opt)
	if err != nil {
		return err
	}

	smallShas, err := catFileBatchCheck(revs)
	if err != nil {
		return err
	}

	pointers, err := catFileBatch(smallShas)
	if err != nil {
		return err
	}

	for p := range pointers.Results {
		if name, ok := opt.GetName(p.Sha1); ok {
			p.Name = name
		}
		cb(p, nil)
	}

	if err := pointers.Wait(); err != nil {
		cb(nil, err)
	}

	return nil
}

// revListShas uses git rev-list to return the list of object sha1s
// for the given ref. If all is true, ref is ignored. It returns a
// channel from which sha1 strings can be read.
func revListShas(refLeft, refRight string, opt *ScanRefsOptions) (*StringChannelWrapper, error) {
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
		args, commits := revListArgsRefVsRemote(refLeft, opt.RemoteName, opt.skippedRefs)
		refArgs = append(refArgs, args...)
		if len(commits) > 0 {
			stdin = commits
		}
	default:
		return nil, errors.New("scanner: unknown scan type: " + strconv.Itoa(int(opt.ScanMode)))
	}

	// Use "--" at the end of the command to disambiguate arguments as refs,
	// so Git doesn't complain about ambiguity if you happen to also have a
	// file named "master".
	refArgs = append(refArgs, "--")

	cmd, err := startCommand("git", refArgs...)
	if err != nil {
		return nil, err
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

	return NewStringChannelWrapper(revs, errchan), nil
}

// Get additional arguments needed to limit 'git rev-list' to just the changes
// in refTo that are also not on remoteName.
//
// Returns a slice of string command arguments, and a slice of string git
// commits to pass to `git rev-list` via STDIN.
func revListArgsRefVsRemote(refTo, remoteName string, skippedRefs []string) ([]string, []string) {
	if len(skippedRefs) < 1 {
		// Safe to use cached
		return []string{refTo, "--not", "--remotes=" + remoteName}, nil
	}

	// Use only the non-missing refs as 'from' points
	commits := make([]string, 1, len(skippedRefs)+1)
	commits[0] = refTo
	return []string{"--stdin"}, append(commits, skippedRefs...)
}
