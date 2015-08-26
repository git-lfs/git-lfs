// Package git contains various commands that shell out to git
package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

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

func ResolveRef(ref string) (string, error) {
	return simpleExec("git", "rev-parse", ref)
}

func CurrentRef() (string, error) {
	return ResolveRef("HEAD")
}

func CurrentBranch() (string, error) {
	return simpleExec("git", "rev-parse", "--abbrev-ref", "HEAD")
}

func CurrentRemoteRef() (string, error) {
	remote, err := CurrentRemote()
	if err != nil {
		return "", err
	}

	return ResolveRef(remote)
}

func CurrentRemote() (string, error) {
	branch, err := CurrentBranch()
	if err != nil {
		return "", err
	}

	if branch == "HEAD" {
		return "", errors.New("not on a branch")
	}

	remote := Config.Find(fmt.Sprintf("branch.%s.remote", branch))
	if remote == "" {
		return "", errors.New("remote not found")
	}

	return remote + "/" + branch, nil
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

// SetGlobal sets the git config value for the key in the global config
func (c *gitConfig) SetGlobal(key, val string) {
	simpleExec("git", "config", "--global", "--add", key, val)
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
	simpleExec("git", "config", "--file", file, "--add", key, val)
}

// UnsetLocalKey removes the git config value for the key from the specified config file
func (c *gitConfig) UnsetLocalKey(file, key string) {
	simpleExec("git", "config", "--file", file, "--unset", key)
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
