// Package git contains various commands that shell out to git
package git

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

type RefType int

const (
	RefTypeLocalBranch  = RefType(iota)
	RefTypeRemoteBranch = RefType(iota)
	RefTypeLocalTag     = RefType(iota)
	RefTypeRemoteTag    = RefType(iota)
	RefTypeHEAD         = RefType(iota) // current checkout
	RefTypeOther        = RefType(iota) // stash or unknown
)

// A git reference (branch, tag etc)
type Ref struct {
	Name string
	Type RefType
	Sha  string
}

// Some top level information about a commit (only first line of message)
type CommitSummary struct {
	Sha            string
	ShortSha       string
	Parents        []string
	CommitDate     time.Time
	AuthorDate     time.Time
	AuthorName     string
	AuthorEmail    string
	CommitterName  string
	CommitterEmail string
	Subject        string
}

func LsRemote(remote, remoteRef string) (string, error) {
	if remote == "" {
		return "", errors.New("remote required")
	}
	if remoteRef == "" {
		return simpleExec("git", "ls-remote", remote)

	}
	return simpleExec("git", "ls-remote", remote, remoteRef)
}

func ResolveRef(ref string) (*Ref, error) {
	outp, err := simpleExec("git", "rev-parse", ref, "--symbolic-full-name", ref)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(outp, "\n")
	if len(lines) <= 1 {
		return nil, fmt.Errorf("Git can't resolve ref: %q", ref)
	}

	fullref := &Ref{Sha: lines[0]}
	fullref.Type, fullref.Name = ParseRefToTypeAndName(lines[1])
	return fullref, nil
}

func CurrentRef() (*Ref, error) {
	return ResolveRef("HEAD")
}

func CurrentBranch() (string, error) {
	return simpleExec("git", "rev-parse", "--abbrev-ref", "HEAD")
}

func CurrentRemoteRef() (*Ref, error) {
	remoteref, err := RemoteRefNameForCurrentBranch()
	if err != nil {
		return nil, err
	}

	return ResolveRef(remoteref)
}

// RemoteForCurrentBranch returns the name of the remote that the current branch is tracking
func RemoteForCurrentBranch() (string, error) {
	branch, err := CurrentBranch()
	if err != nil {
		return "", err
	}
	remote := RemoteForBranch(branch)
	if remote == "" {
		return "", errors.New("remote not found")
	}
	return remote, nil
}

// RemoteRefForCurrentBranch returns the full remote ref (remote/remotebranch) that the current branch is tracking
func RemoteRefNameForCurrentBranch() (string, error) {
	branch, err := CurrentBranch()
	if err != nil {
		return "", err
	}

	if branch == "HEAD" {
		return "", errors.New("not on a branch")
	}

	remote := RemoteForBranch(branch)
	if remote == "" {
		return "", errors.New("remote not found")
	}

	remotebranch := RemoteBranchForLocalBranch(branch)

	return remote + "/" + remotebranch, nil
}

// RemoteForBranch returns the remote name that a given local branch is tracking (blank if none)
func RemoteForBranch(localBranch string) string {
	return Config.Find(fmt.Sprintf("branch.%s.remote", localBranch))
}

// RemoteBranchForLocalBranch returns the name (only) of the remote branch that the local branch is tracking
// If no specific branch is configured, returns local branch name
func RemoteBranchForLocalBranch(localBranch string) string {
	// get remote ref to track, may not be same name
	merge := Config.Find(fmt.Sprintf("branch.%s.merge", localBranch))
	if strings.HasPrefix(merge, "refs/heads/") {
		return merge[11:]
	} else {
		return localBranch
	}

}

func RemoteList() ([]string, error) {
	cmd := execCommand("git", "remote")

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to call git remote: %v", err)
	}
	cmd.Start()
	scanner := bufio.NewScanner(outp)

	var ret []string
	for scanner.Scan() {
		ret = append(ret, strings.TrimSpace(scanner.Text()))
	}
	return ret, nil
}

// ValidateRemote checks that a named remote is valid for use
// Mainly to check user-supplied remotes & fail more nicely
func ValidateRemote(remote string) error {
	remotes, err := RemoteList()
	if err != nil {
		return err
	}
	for _, r := range remotes {
		if r == remote {
			return nil
		}
	}
	return errors.New("Invalid remote name")
}

// DefaultRemote returns the default remote based on:
// 1. The currently tracked remote branch, if present
// 2. "origin", if defined
// 3. Any other SINGLE remote defined in .git/config
// Returns an error if all of these fail, i.e. no tracked remote branch, no
// "origin", and either no remotes defined or 2+ non-"origin" remotes
func DefaultRemote() (string, error) {
	tracked, err := RemoteForCurrentBranch()
	if err == nil {
		return tracked, nil
	}

	// Otherwise, check what remotes are defined
	remotes, err := RemoteList()
	if err != nil {
		return "", err
	}
	switch len(remotes) {
	case 0:
		return "", errors.New("No remotes defined")
	case 1: // always use a single remote whether it's origin or otherwise
		return remotes[0], nil
	default:
		for _, remote := range remotes {
			// Use origin if present
			if remote == "origin" {
				return remote, nil
			}
		}
	}
	return "", errors.New("Unable to pick default remote, too ambiguous")
}

func UpdateIndex(file string) error {
	_, err := simpleExec("git", "update-index", "-q", "--refresh", file)
	return err
}

type gitConfig struct {
}

var Config = &gitConfig{}

// Find returns the git config value for the key
func (c *gitConfig) Find(val string) string {
	output, _ := simpleExec("git", "config", val)
	return output
}

// Find returns the git config value for the key
func (c *gitConfig) FindGlobal(val string) string {
	output, _ := simpleExec("git", "config", "--global", val)
	return output
}

// Find returns the git config value for the key
func (c *gitConfig) FindLocal(val string) string {
	output, _ := simpleExec("git", "config", "--local", val)
	return output
}

// SetGlobal sets the git config value for the key in the global config
func (c *gitConfig) SetGlobal(key, val string) {
	simpleExec("git", "config", "--global", key, val)
}

// UnsetGlobal removes the git config value for the key from the global config
func (c *gitConfig) UnsetGlobal(key string) {
	simpleExec("git", "config", "--global", "--unset", key)
}

func (c *gitConfig) UnsetGlobalSection(key string) {
	simpleExec("git", "config", "--global", "--remove-section", key)
}

// SetLocal sets the git config value for the key in the specified config file
func (c *gitConfig) SetLocal(file, key, val string) {
	args := make([]string, 1, 5)
	args[0] = "config"
	if len(file) > 0 {
		args = append(args, "--file", file)
	}
	args = append(args, key, val)
	simpleExec("git", args...)
}

// UnsetLocalKey removes the git config value for the key from the specified config file
func (c *gitConfig) UnsetLocalKey(file, key string) {
	args := make([]string, 1, 5)
	args[0] = "config"
	if len(file) > 0 {
		args = append(args, "--file", file)
	}
	args = append(args, "--unset", key)
	simpleExec("git", args...)
}

// List lists all of the git config values
func (c *gitConfig) List() (string, error) {
	return simpleExec("git", "config", "-l")
}

// ListFromFile lists all of the git config values in the given config file
func (c *gitConfig) ListFromFile(f string) (string, error) {
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return "", nil
	}
	return simpleExec("git", "config", "-l", "-f", f)
}

// Version returns the git version
func (c *gitConfig) Version() (string, error) {
	return simpleExec("git", "version")
}

// simpleExec is a small wrapper around os/exec.Command.
func simpleExec(name string, args ...string) (string, error) {
	tracerx.Printf("run_command: '%s' %s", name, strings.Join(args, " "))
	cmd := execCommand(name, args...)

	output, err := cmd.Output()
	if _, ok := err.(*exec.ExitError); ok {
		return "", nil
	}
	if err != nil {
		return fmt.Sprintf("Error running %s %s", name, args), err
	}

	return strings.Trim(string(output), " \n"), nil
}

// RecentBranches returns branches with commit dates on or after the given date/time
// Return full Ref type for easier detection of duplicate SHAs etc
// since: refs with commits on or after this date will be included
// includeRemoteBranches: true to include refs on remote branches
// onlyRemote: set to non-blank to only include remote branches on a single remote
func RecentBranches(since time.Time, includeRemoteBranches bool, onlyRemote string) ([]*Ref, error) {
	cmd := execCommand("git", "for-each-ref",
		`--sort=-committerdate`,
		`--format=%(refname) %(objectname) %(committerdate:iso)`,
		"refs")
	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to call git for-each-ref: %v", err)
	}
	cmd.Start()
	scanner := bufio.NewScanner(outp)

	// Output is like this:
	// refs/heads/master f03686b324b29ff480591745dbfbbfa5e5ac1bd5 2015-08-19 16:50:37 +0100
	// refs/remotes/origin/master ad3b29b773e46ad6870fdf08796c33d97190fe93 2015-08-13 16:50:37 +0100

	// Output is ordered by latest commit date first, so we can stop at the threshold
	regex := regexp.MustCompile(`^(refs/[^/]+/\S+)\s+([0-9A-Za-z]{40})\s+(\d{4}-\d{2}-\d{2}\s+\d{2}\:\d{2}\:\d{2}\s+[\+\-]\d{4})`)

	var ret []*Ref
	for scanner.Scan() {
		line := scanner.Text()
		if match := regex.FindStringSubmatch(line); match != nil {
			fullref := match[1]
			sha := match[2]
			reftype, ref := ParseRefToTypeAndName(fullref)
			if reftype == RefTypeRemoteBranch || reftype == RefTypeRemoteTag {
				if !includeRemoteBranches {
					continue
				}
				if onlyRemote != "" && !strings.HasPrefix(ref, onlyRemote+"/") {
					continue
				}
			}
			// This is a ref we might use
			// Check the date
			commitDate, err := ParseGitDate(match[3])
			if err != nil {
				return ret, err
			}
			if commitDate.Before(since) {
				// the end
				break
			}
			ret = append(ret, &Ref{ref, reftype, sha})
		}
	}
	cmd.Wait()

	return ret, nil

}

// Get the type & name of a git reference
func ParseRefToTypeAndName(fullref string) (t RefType, name string) {
	const localPrefix = "refs/heads/"
	const remotePrefix = "refs/remotes/"
	const remoteTagPrefix = "refs/remotes/tags/"
	const localTagPrefix = "refs/tags/"

	if fullref == "HEAD" {
		name = fullref
		t = RefTypeHEAD
	} else if strings.HasPrefix(fullref, localPrefix) {
		name = fullref[len(localPrefix):]
		t = RefTypeLocalBranch
	} else if strings.HasPrefix(fullref, remotePrefix) {
		name = fullref[len(remotePrefix):]
		t = RefTypeRemoteBranch
	} else if strings.HasPrefix(fullref, remoteTagPrefix) {
		name = fullref[len(remoteTagPrefix):]
		t = RefTypeRemoteTag
	} else if strings.HasPrefix(fullref, localTagPrefix) {
		name = fullref[len(localTagPrefix):]
		t = RefTypeLocalTag
	} else {
		name = fullref
		t = RefTypeOther
	}
	return
}

// Parse a Git date formatted in ISO 8601 format (%ci/%ai)
func ParseGitDate(str string) (time.Time, error) {

	// Unfortunately Go and Git don't overlap in their builtin date formats
	// Go's time.RFC1123Z and Git's %cD are ALMOST the same, except that
	// when the day is < 10 Git outputs a single digit, but Go expects a leading
	// zero - this is enough to break the parsing. Sigh.

	// Format is for 2 Jan 2006, 15:04:05 -7 UTC as per Go
	return time.Parse("2006-01-02 15:04:05 -0700", str)
}

// FormatGitDate converts a Go date into a git command line format date
func FormatGitDate(tm time.Time) string {
	// Git format is "Fri Jun 21 20:26:41 2013 +0900" but no zero-leading for day
	return tm.Format("Mon Jan 2 15:04:05 2006 -0700")
}

// Get summary information about a commit
func GetCommitSummary(commit string) (*CommitSummary, error) {
	cmd := execCommand("git", "show", "-s",
		`--format=%H|%h|%P|%ai|%ci|%ae|%an|%ce|%cn|%s`, commit)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Failed to call git show: %v %v", err, string(out))
	}

	// At most 10 substrings so subject line is not split on anything
	fields := strings.SplitN(string(out), "|", 10)
	// Cope with the case where subject is blank
	if len(fields) >= 9 {
		ret := &CommitSummary{}
		// Get SHAs from output, not commit input, so we can support symbolic refs
		ret.Sha = fields[0]
		ret.ShortSha = fields[1]
		ret.Parents = strings.Split(fields[2], " ")
		// %aD & %cD (RFC2822) matches Go's RFC1123Z format
		ret.AuthorDate, _ = ParseGitDate(fields[3])
		ret.CommitDate, _ = ParseGitDate(fields[4])
		ret.AuthorEmail = fields[5]
		ret.AuthorName = fields[6]
		ret.CommitterEmail = fields[7]
		ret.CommitterName = fields[8]
		if len(fields) > 9 {
			ret.Subject = strings.TrimRight(fields[9], "\n")
		}
		return ret, nil
	} else {
		msg := fmt.Sprintf("Unexpected output from git show: %v", string(out))
		return nil, errors.New(msg)
	}

}

func RootDir() (string, error) {
	cmd := execCommand("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to call git rev-parse --show-toplevel: %v %v", err, string(out))
	}

	path := strings.TrimSpace(string(out))
	if len(path) > 0 {
		return filepath.Abs(path)
	}
	return "", nil

}
func GitDir() (string, error) {
	cmd := execCommand("git", "rev-parse", "--git-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to call git rev-parse --git-dir: %v %v", err, string(out))
	}
	path := strings.TrimSpace(string(out))
	if len(path) > 0 {
		return filepath.Abs(path)
	}
	return "", nil

}
