// Package git contains various commands that shell out to git
package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/github/git-lfs/vendor/_nuts/github.com/kr/pty"
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

func CurrentRemoteRef() (*Ref, error) {
	remoteref, err := RemoteRefNameForCurrentBranch()
	if err != nil {
		return nil, err
	}

	return ResolveRef(remoteref)
}

// RemoteForCurrentBranch returns the name of the remote that the current branch is tracking
func RemoteForCurrentBranch() (string, error) {
	ref, err := CurrentRef()
	if err != nil {
		return "", err
	}
	remote := RemoteForBranch(ref.Name)
	if remote == "" {
		return "", fmt.Errorf("remote not found for branch %q", ref.Name)
	}
	return remote, nil
}

// RemoteRefForCurrentBranch returns the full remote ref (remote/remotebranch) that the current branch is tracking
func RemoteRefNameForCurrentBranch() (string, error) {
	ref, err := CurrentRef()
	if err != nil {
		return "", err
	}

	if ref.Type == RefTypeHEAD || ref.Type == RefTypeOther {
		return "", errors.New("not on a branch")
	}

	remote := RemoteForBranch(ref.Name)
	if remote == "" {
		return "", fmt.Errorf("remote not found for branch %q", ref.Name)
	}

	remotebranch := RemoteBranchForLocalBranch(ref.Name)

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
	defer cmd.Wait()

	scanner := bufio.NewScanner(outp)

	var ret []string
	for scanner.Scan() {
		ret = append(ret, strings.TrimSpace(scanner.Text()))
	}

	return ret, nil
}

// Refs returns all of the local and remote branches and tags for the current
// repository. Other refs (HEAD, refs/stash, git notes) are ignored.
func LocalRefs() ([]*Ref, error) {
	cmd := execCommand("git", "show-ref")

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to call git show-ref: %v", err)
	}
	cmd.Start()
	defer cmd.Wait()

	var refs []*Ref
	scanner := bufio.NewScanner(outp)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 || len(parts[0]) != 40 || len(parts[1]) < 1 {
			tracerx.Printf("Invalid line from git show-ref: %q", line)
			continue
		}

		rtype, name := ParseRefToTypeAndName(parts[1])
		if rtype != RefTypeLocalBranch && rtype != RefTypeLocalTag {
			continue
		}

		refs = append(refs, &Ref{name, rtype, parts[0]})
	}

	return refs, nil
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
	return simpleExec("git", "config", "-l", "-f", f)
}

// Version returns the git version
func (c *gitConfig) Version() (string, error) {
	return simpleExec("git", "version")
}

// IsVersionAtLeast returns whether the git version is the one specified or higher
// argument is plain version string separated by '.' e.g. "2.3.1" but can omit minor/patch
func (c *gitConfig) IsGitVersionAtLeast(ver string) bool {
	gitver, err := c.Version()
	if err != nil {
		tracerx.Printf("Error getting git version: %v", err)
		return false
	}
	return IsVersionAtLeast(gitver, ver)
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
	defer cmd.Wait()

	scanner := bufio.NewScanner(outp)

	// Output is like this:
	// refs/heads/master f03686b324b29ff480591745dbfbbfa5e5ac1bd5 2015-08-19 16:50:37 +0100
	// refs/remotes/origin/master ad3b29b773e46ad6870fdf08796c33d97190fe93 2015-08-13 16:50:37 +0100

	// Output is ordered by latest commit date first, so we can stop at the threshold
	regex := regexp.MustCompile(`^(refs/[^/]+/\S+)\s+([0-9A-Za-z]{40})\s+(\d{4}-\d{2}-\d{2}\s+\d{2}\:\d{2}\:\d{2}\s+[\+\-]\d{4})`)
	tracerx.Printf("RECENT: Getting refs >= %v", since)
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
			tracerx.Printf("RECENT: %v (%v)", ref, commitDate)
			ret = append(ret, &Ref{ref, reftype, sha})
		}
	}

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

func GitAndRootDirs() (string, string, error) {
	cmd := execCommand("git", "rev-parse", "--git-dir", "--show-toplevel")
	buf := &bytes.Buffer{}
	cmd.Stderr = buf

	out, err := cmd.Output()
	output := string(out)
	if err != nil {
		return "", "", fmt.Errorf("Failed to call git rev-parse --git-dir --show-toplevel: %q", buf.String())
	}

	paths := strings.Split(output, "\n")
	pathLen := len(paths)

	if pathLen == 0 {
		return "", "", fmt.Errorf("Bad git rev-parse output: %q", output)
	}

	absGitDir, err := filepath.Abs(paths[0])
	if err != nil {
		return "", "", fmt.Errorf("Error converting %q to absolute: %s", paths[0], err)
	}

	if pathLen == 1 || len(paths[1]) == 0 {
		return absGitDir, "", nil
	}

	absRootDir, err := filepath.Abs(paths[1])
	if err != nil {
		return "", "", fmt.Errorf("Error converting %q to absolute: %s", paths[1], err)
	}

	return absGitDir, absRootDir, nil
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

// GetAllWorkTreeHEADs returns the refs that all worktrees are using as HEADs
// This returns all worktrees plus the master working copy, and works even if
// working dir is actually in a worktree right now
// Pass in the git storage dir (parent of 'objects') to work from
func GetAllWorkTreeHEADs(storageDir string) ([]*Ref, error) {
	worktreesdir := filepath.Join(storageDir, "worktrees")
	dirf, err := os.Open(worktreesdir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	var worktrees []*Ref
	if err == nil {
		// There are some worktrees
		defer dirf.Close()
		direntries, err := dirf.Readdir(0)
		if err != nil {
			return nil, err
		}
		for _, dirfi := range direntries {
			if dirfi.IsDir() {
				// to avoid having to chdir and run git commands to identify the commit
				// just read the HEAD file & git rev-parse if necessary
				// Since the git repo is shared the same rev-parse will work from this location
				headfile := filepath.Join(worktreesdir, dirfi.Name(), "HEAD")
				ref, err := parseRefFile(headfile)
				if err != nil {
					tracerx.Printf("Error reading %v for worktree, skipping: %v", headfile, err)
					continue
				}
				worktrees = append(worktrees, ref)
			}
		}
	}

	// This has only established the separate worktrees, not the original checkout
	// If the storageDir contains a HEAD file then there is a main checkout
	// as well; this mus tbe resolveable whether you're in the main checkout or
	// a worktree
	headfile := filepath.Join(storageDir, "HEAD")
	ref, err := parseRefFile(headfile)
	if err == nil {
		worktrees = append(worktrees, ref)
	} else if !os.IsNotExist(err) { // ok if not exists, probably bare repo
		tracerx.Printf("Error reading %v for main checkout, skipping: %v", headfile, err)
	}

	return worktrees, nil
}

// Manually parse a reference file like HEAD and return the Ref it resolves to
func parseRefFile(filename string) (*Ref, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	contents := strings.TrimSpace(string(bytes))
	if strings.HasPrefix(contents, "ref:") {
		contents = strings.TrimSpace(contents[4:])
	}
	return ResolveRef(contents)
}

// IsVersionAtLeast compares 2 version strings (ok to be prefixed with 'git version', ignores)
func IsVersionAtLeast(actualVersion, desiredVersion string) bool {
	// Capture 1-3 version digits, optionally prefixed with 'git version' and possibly
	// with suffixes which we'll ignore (e.g. unstable builds, MinGW versions)
	verregex := regexp.MustCompile(`(?:git version\s+)?(\d+)(?:.(\d+))?(?:.(\d+))?.*`)

	var atleast uint64
	// Support up to 1000 in major/minor/patch digits
	const majorscale = 1000 * 1000
	const minorscale = 1000

	if match := verregex.FindStringSubmatch(desiredVersion); match != nil {
		// Ignore errors as regex won't match anything other than digits
		major, _ := strconv.Atoi(match[1])
		atleast += uint64(major * majorscale)
		if len(match) > 2 {
			minor, _ := strconv.Atoi(match[2])
			atleast += uint64(minor * minorscale)
		}
		if len(match) > 3 {
			patch, _ := strconv.Atoi(match[3])
			atleast += uint64(patch)
		}
	}

	var actual uint64
	if match := verregex.FindStringSubmatch(actualVersion); match != nil {
		major, _ := strconv.Atoi(match[1])
		actual += uint64(major * majorscale)
		if len(match) > 2 {
			minor, _ := strconv.Atoi(match[2])
			actual += uint64(minor * minorscale)
		}
		if len(match) > 3 {
			patch, _ := strconv.Atoi(match[3])
			actual += uint64(patch)
		}
	}

	return actual >= atleast
}

// CloneWithoutFilters clones a git repo but without the smudge filter enabled
// so that files in the working copy will be pointers and not real LFS data
func CloneWithoutFilters(args []string) error {

	// Before git 2.2, setting filters to blank fails, so use cat instead (slightly slower)
	filterOverride := ""
	if !Config.IsGitVersionAtLeast("2.2.0") {
		filterOverride = "cat"
	}
	// Disable the LFS filters while cloning to speed things up
	// this is especially effective on Windows where even calling git-lfs at all
	// with --skip-smudge is costly across many files in a checkout
	cmdargs := []string{
		"-c", fmt.Sprintf("filter.lfs.smudge=%v", filterOverride),
		"-c", "filter.lfs.required=false",
		"clone"}
	cmdargs = append(cmdargs, args...)
	cmd := execCommand("git", cmdargs...)

	// Spool stdout directly to our own
	cmd.Stdout = os.Stdout

	// Assign pty/tty so git thinks it's a real terminal
	outpty, outtty, err := pty.Open()
	cmd.Stdin = outtty
	cmd.Stdout = outtty
	errpty, errtty, err := pty.Open()
	// stderr needs filtering
	cmd.Stderr = errtty
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setctty = true
	cmd.SysProcAttr.Setsid = true

	var outputWait sync.WaitGroup
	outputWait.Add(2)
	go func() {
		io.Copy(os.Stdout, outpty)
		outputWait.Done()
	}()
	go func() {
		// Filter stderr to exclude messages caused by disabling the filters
		// As of git 2.7 it still tries to call the blank filter but required=false
		// this problem should be gone in git 2.8 https://github.com/git/git/commit/1a8630d
		scanner := bufio.NewScanner(errpty)
		for scanner.Scan() {
			s := scanner.Text()

			// Swallow all the known messages from intentionally breaking filter
			if strings.Contains(s, "error: external filter") ||
				strings.Contains(s, "error: cannot fork") ||
				// Linux / Mac messages
				strings.Contains(s, "error: cannot run : No such file or directory") ||
				strings.Contains(s, "warning: Clone succeeded, but checkout failed") ||
				strings.Contains(s, "You can inspect what was checked out with 'git status'") ||
				strings.Contains(s, "retry the checkout") ||
				// Windows messages
				strings.Contains(s, "error: cannot spawn : No such file or directory") ||
				// blank formatting
				len(strings.TrimSpace(s)) == 0 {
				// Send filtered stderr to trace in case useful
				tracerx.Printf(s)
				continue
			}
			os.Stderr.WriteString(s)
			os.Stderr.WriteString("\n") // stripped by scanner
		}
		outputWait.Done()
	}()

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Failed to start git clone: %v", err)
	}

	outtty.Close()
	errtty.Close()

	err = cmd.Wait()
	outputWait.Wait()
	if err != nil {
		return fmt.Errorf("git clone failed: %v", err)
	}

	return nil
}

// CachedRemoteRefs returns the list of branches & tags for a remote which are
// currently cached locally. No remote request is made to verify them.
func CachedRemoteRefs(remoteName string) ([]*Ref, error) {

	var ret []*Ref
	cmd := execCommand("git", "show-ref")

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to call git show-ref: %v", err)
	}
	cmd.Start()
	scanner := bufio.NewScanner(outp)

	r := regexp.MustCompile(fmt.Sprintf(`([0-9a-fA-F]{40})\s+refs/remotes/%v/(.*)`, remoteName))
	for scanner.Scan() {
		if match := r.FindStringSubmatch(scanner.Text()); match != nil {
			name := strings.TrimSpace(match[2])
			// Don't match head
			if name == "HEAD" {
				continue
			}

			sha := match[1]
			ret = append(ret, &Ref{name, RefTypeRemoteBranch, sha})
		}
	}
	return ret, nil
}

// RemoteRefs returns a list of branches & tags for a remote by actually
// accessing the remote vir git ls-remote
func RemoteRefs(remoteName string) ([]*Ref, error) {

	var ret []*Ref
	cmd := execCommand("git", "ls-remote", "--heads", "--tags", "-q", remoteName)

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to call git ls-remote: %v", err)
	}
	cmd.Start()
	scanner := bufio.NewScanner(outp)

	r := regexp.MustCompile(`([0-9a-fA-F]{40})\s+refs/(heads|tags)/(.*)`)
	for scanner.Scan() {
		if match := r.FindStringSubmatch(scanner.Text()); match != nil {
			name := strings.TrimSpace(match[3])
			// Don't match head
			if name == "HEAD" {
				continue
			}

			sha := match[1]
			if match[2] == "heads" {
				ret = append(ret, &Ref{name, RefTypeRemoteBranch, sha})
			} else {
				ret = append(ret, &Ref{name, RefTypeRemoteTag, sha})
			}
		}
	}
	return ret, nil
}

// An env for an exec.Command without GIT_TRACE
var env []string
var traceEnv = "GIT_TRACE="

func init() {
	realEnv := os.Environ()
	env = make([]string, 0, len(realEnv))

	for _, kv := range realEnv {
		if strings.HasPrefix(kv, traceEnv) {
			continue
		}
		env = append(env, kv)
	}
}
