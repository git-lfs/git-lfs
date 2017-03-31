// Package git contains various commands that shell out to git
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	lfserrors "github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

type RefType int

const (
	RefTypeLocalBranch  = RefType(iota)
	RefTypeRemoteBranch = RefType(iota)
	RefTypeLocalTag     = RefType(iota)
	RefTypeRemoteTag    = RefType(iota)
	RefTypeHEAD         = RefType(iota) // current checkout
	RefTypeOther        = RefType(iota) // stash or unknown

	// A ref which can be used as a placeholder for before the first commit
	// Equivalent to git mktree < /dev/null, useful for diffing before first commit
	RefBeforeFirstCommit = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
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

func CommitTree(tree, parent, message string) (string, error) {
	args := []string{
		"commit-tree", "-m", message,
	}
	if len(parent) > 0 {
		args = append(args, "-p", parent)
	}

	return subprocess.SimpleExec("git", append(args, tree)...)
}

func UpdateRef(from, to string) error {
	args := []string{
		"update-ref", from, to,
	}

	_, err := subprocess.SimpleExec("git", args...)
	return err
}

func ReadTree(treeish string) error {
	_, err := subprocess.SimpleExec("git", "read-tree", treeish)
	return err
}

func WriteTree(prefix string) (string, error) {
	if len(prefix) == 0 {
		prefix = "/"
	}

	args := []string{
		"write-tree", // fmt.Sprintf("--prefix=%s", prefix),
	}

	return subprocess.SimpleExec("git", args...)
}

func CheckoutIndex(files ...string) error {
	args := []string{
		"checkout-index", "-u",
	}

	_, err := subprocess.SimpleExec("git", append(args, files...)...)
	return err
}

func LsRemote(remote, remoteRef string) (string, error) {
	if remote == "" {
		return "", errors.New("remote required")
	}
	if remoteRef == "" {
		return subprocess.SimpleExec("git", "ls-remote", remote)

	}
	return subprocess.SimpleExec("git", "ls-remote", remote, remoteRef)
}

func ResolveRef(ref string) (*Ref, error) {
	outp, err := subprocess.SimpleExec("git", "rev-parse", ref, "--symbolic-full-name", ref)
	if err != nil {
		return nil, fmt.Errorf("Git can't resolve ref: %q", ref)
	}
	if outp == "" {
		return nil, fmt.Errorf("Git can't resolve ref: %q", ref)
	}

	lines := strings.Split(outp, "\n")
	fullref := &Ref{Sha: lines[0]}

	if len(lines) == 1 {
		// ref is a sha1 and has no symbolic-full-name
		fullref.Name = lines[0] // fullref.Sha
		fullref.Type = RefTypeOther
		return fullref, nil
	}

	// parse the symbolic-full-name
	fullref.Type, fullref.Name = ParseRefToTypeAndName(lines[1])
	return fullref, nil
}

func ResolveRefs(refnames []string) ([]*Ref, error) {
	refs := make([]*Ref, len(refnames))
	for i, name := range refnames {
		ref, err := ResolveRef(name)
		if err != nil {
			return refs, err
		}

		refs[i] = ref
	}
	return refs, nil
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

// RemoteRefForCurrentBranch returns the full remote ref (refs/remotes/{remote}/{remotebranch})
// that the current branch is tracking.
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

	return fmt.Sprintf("refs/remotes/%s/%s", remote, remotebranch), nil
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
	cmd := subprocess.ExecCommand("git", "remote")

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
	cmd := subprocess.ExecCommand("git", "show-ref", "--heads", "--tags")

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to call git show-ref: %v", err)
	}

	var refs []*Ref

	if err := cmd.Start(); err != nil {
		return refs, err
	}

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

	return refs, cmd.Wait()
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

	if err = ValidateRemoteURL(remote); err == nil {
		return nil
	}

	return fmt.Errorf("Invalid remote name: %q", remote)
}

// ValidateRemoteURL checks that a string is a valid Git remote URL
func ValidateRemoteURL(remote string) error {
	u, err := url.Parse(remote)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "ssh", "http", "https", "git":
		return nil
	case "":
		// This is either an invalid remote name (maybe the user made a typo
		// when selecting a named remote) or a bare SSH URL like
		// "x@y.com:path/to/resource.git". Guess that this is a URL in the latter
		// form if the string contains a colon ":", and an invalid remote if it
		// does not.
		if strings.Contains(remote, ":") {
			return nil
		}
		return fmt.Errorf("Invalid remote name: %q", remote)
	default:
		return fmt.Errorf("Invalid remote url protocol %q in %q", u.Scheme, remote)
	}
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
	_, err := subprocess.SimpleExec("git", "update-index", "--add", "-q", "--refresh", file)
	return err
}

type gitConfig struct {
	gitVersion string
	mu         sync.Mutex
}

var Config = &gitConfig{}

// Find returns the git config value for the key
func (c *gitConfig) Find(val string) string {
	output, _ := subprocess.SimpleExec("git", "config", val)
	return output
}

// FindGlobal returns the git config value global scope for the key
func (c *gitConfig) FindGlobal(val string) string {
	output, _ := subprocess.SimpleExec("git", "config", "--global", val)
	return output
}

// FindSystem returns the git config value in system scope for the key
func (c *gitConfig) FindSystem(val string) string {
	output, _ := subprocess.SimpleExec("git", "config", "--system", val)
	return output
}

// Find returns the git config value for the key
func (c *gitConfig) FindLocal(val string) string {
	output, _ := subprocess.SimpleExec("git", "config", "--local", val)
	return output
}

// SetGlobal sets the git config value for the key in the global config
func (c *gitConfig) SetGlobal(key, val string) (string, error) {
	return subprocess.SimpleExec("git", "config", "--global", key, val)
}

// SetSystem sets the git config value for the key in the system config
func (c *gitConfig) SetSystem(key, val string) (string, error) {
	return subprocess.SimpleExec("git", "config", "--system", key, val)
}

// UnsetGlobal removes the git config value for the key from the global config
func (c *gitConfig) UnsetGlobal(key string) (string, error) {
	return subprocess.SimpleExec("git", "config", "--global", "--unset", key)
}

// UnsetSystem removes the git config value for the key from the system config
func (c *gitConfig) UnsetSystem(key string) (string, error) {
	return subprocess.SimpleExec("git", "config", "--system", "--unset", key)
}

// UnsetGlobalSection removes the entire named section from the global config
func (c *gitConfig) UnsetGlobalSection(key string) (string, error) {
	return subprocess.SimpleExec("git", "config", "--global", "--remove-section", key)
}

// UnsetSystemSection removes the entire named section from the system config
func (c *gitConfig) UnsetSystemSection(key string) (string, error) {
	return subprocess.SimpleExec("git", "config", "--system", "--remove-section", key)
}

// UnsetLocalSection removes the entire named section from the system config
func (c *gitConfig) UnsetLocalSection(key string) (string, error) {
	return subprocess.SimpleExec("git", "config", "--local", "--remove-section", key)
}

// SetLocal sets the git config value for the key in the specified config file
func (c *gitConfig) SetLocal(file, key, val string) (string, error) {
	args := make([]string, 1, 5)
	args[0] = "config"
	if len(file) > 0 {
		args = append(args, "--file", file)
	}
	args = append(args, key, val)
	return subprocess.SimpleExec("git", args...)
}

// UnsetLocalKey removes the git config value for the key from the specified config file
func (c *gitConfig) UnsetLocalKey(file, key string) (string, error) {
	args := make([]string, 1, 5)
	args[0] = "config"
	if len(file) > 0 {
		args = append(args, "--file", file)
	}
	args = append(args, "--unset", key)
	return subprocess.SimpleExec("git", args...)
}

// List lists all of the git config values
func (c *gitConfig) List() (string, error) {
	return subprocess.SimpleExec("git", "config", "-l")
}

// ListFromFile lists all of the git config values in the given config file
func (c *gitConfig) ListFromFile(f string) (string, error) {
	return subprocess.SimpleExec("git", "config", "-l", "-f", f)
}

// Version returns the git version
func (c *gitConfig) Version() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.gitVersion) == 0 {
		v, err := subprocess.SimpleExec("git", "version")
		if err != nil {
			return v, err
		}
		c.gitVersion = v
	}

	return c.gitVersion, nil
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

// RecentBranches returns branches with commit dates on or after the given date/time
// Return full Ref type for easier detection of duplicate SHAs etc
// since: refs with commits on or after this date will be included
// includeRemoteBranches: true to include refs on remote branches
// onlyRemote: set to non-blank to only include remote branches on a single remote
func RecentBranches(since time.Time, includeRemoteBranches bool, onlyRemote string) ([]*Ref, error) {
	cmd := subprocess.ExecCommand("git", "for-each-ref",
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
	cmd := subprocess.ExecCommand("git", "show", "-s",
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
	cmd := subprocess.ExecCommand("git", "rev-parse", "--git-dir", "--show-toplevel")
	buf := &bytes.Buffer{}
	cmd.Stderr = buf

	out, err := cmd.Output()
	output := string(out)
	if err != nil {
		return "", "", fmt.Errorf("Failed to call git rev-parse --git-dir --show-toplevel: %q", buf.String())
	}

	paths := strings.Split(output, "\n")
	pathLen := len(paths)

	for i := 0; i < pathLen; i++ {
		paths[i], err = tools.TranslateCygwinPath(paths[i])
	}

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

	absRootDir := paths[1]
	return absGitDir, absRootDir, nil
}

func RootDir() (string, error) {
	cmd := subprocess.ExecCommand("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to call git rev-parse --show-toplevel: %v %v", err, string(out))
	}

	path := strings.TrimSpace(string(out))
	path, err = tools.TranslateCygwinPath(path)
	if len(path) > 0 {
		return filepath.Abs(path)
	}
	return "", nil

}

func GitDir() (string, error) {
	cmd := subprocess.ExecCommand("git", "rev-parse", "--git-dir")
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

// For compatibility with git clone we must mirror all flags in CloneWithoutFilters
type CloneFlags struct {
	// --template <template_directory>
	TemplateDirectory string
	// -l --local
	Local bool
	// -s --shared
	Shared bool
	// --no-hardlinks
	NoHardlinks bool
	// -q --quiet
	Quiet bool
	// -n --no-checkout
	NoCheckout bool
	// --progress
	Progress bool
	// --bare
	Bare bool
	// --mirror
	Mirror bool
	// -o <name> --origin <name>
	Origin string
	// -b <name> --branch <name>
	Branch string
	// -u <upload-pack> --upload-pack <pack>
	Upload string
	// --reference <repository>
	Reference string
	// --dissociate
	Dissociate bool
	// --separate-git-dir <git dir>
	SeparateGit string
	// --depth <depth>
	Depth string
	// --recursive
	Recursive bool
	// --recurse-submodules
	RecurseSubmodules bool
	// -c <value> --config <value>
	Config string
	// --single-branch
	SingleBranch bool
	// --no-single-branch
	NoSingleBranch bool
	// --verbose
	Verbose bool
	// --ipv4
	Ipv4 bool
	// --ipv6
	Ipv6 bool
}

// CloneWithoutFilters clones a git repo but without the smudge filter enabled
// so that files in the working copy will be pointers and not real LFS data
func CloneWithoutFilters(flags CloneFlags, args []string) error {

	// Before git 2.8, setting filters to blank causes lots of warnings, so use cat instead (slightly slower)
	// Also pre 2.2 it failed completely. We used to use it anyway in git 2.2-2.7 and
	// suppress the messages in stderr, but doing that with standard StderrPipe suppresses
	// the git clone output (git thinks it's not a terminal) and makes it look like it's
	// not working. You can get around that with https://github.com/kr/pty but that
	// causes difficult issues with passing through Stdin for login prompts
	// This way is simpler & more practical.
	filterOverride := ""
	if !Config.IsGitVersionAtLeast("2.8.0") {
		filterOverride = "cat"
	}
	// Disable the LFS filters while cloning to speed things up
	// this is especially effective on Windows where even calling git-lfs at all
	// with --skip-smudge is costly across many files in a checkout
	cmdargs := []string{
		"-c", fmt.Sprintf("filter.lfs.smudge=%v", filterOverride),
		"-c", "filter.lfs.process=",
		"-c", "filter.lfs.required=false",
		"clone"}

	// flags
	if flags.Bare {
		cmdargs = append(cmdargs, "--bare")
	}
	if len(flags.Branch) > 0 {
		cmdargs = append(cmdargs, "--branch", flags.Branch)
	}
	if len(flags.Config) > 0 {
		cmdargs = append(cmdargs, "--config", flags.Config)
	}
	if len(flags.Depth) > 0 {
		cmdargs = append(cmdargs, "--depth", flags.Depth)
	}
	if flags.Dissociate {
		cmdargs = append(cmdargs, "--dissociate")
	}
	if flags.Ipv4 {
		cmdargs = append(cmdargs, "--ipv4")
	}
	if flags.Ipv6 {
		cmdargs = append(cmdargs, "--ipv6")
	}
	if flags.Local {
		cmdargs = append(cmdargs, "--local")
	}
	if flags.Mirror {
		cmdargs = append(cmdargs, "--mirror")
	}
	if flags.NoCheckout {
		cmdargs = append(cmdargs, "--no-checkout")
	}
	if flags.NoHardlinks {
		cmdargs = append(cmdargs, "--no-hardlinks")
	}
	if flags.NoSingleBranch {
		cmdargs = append(cmdargs, "--no-single-branch")
	}
	if len(flags.Origin) > 0 {
		cmdargs = append(cmdargs, "--origin", flags.Origin)
	}
	if flags.Progress {
		cmdargs = append(cmdargs, "--progress")
	}
	if flags.Quiet {
		cmdargs = append(cmdargs, "--quiet")
	}
	if flags.Recursive {
		cmdargs = append(cmdargs, "--recursive")
	}
	if flags.RecurseSubmodules {
		cmdargs = append(cmdargs, "--recurse-submodules")
	}
	if len(flags.Reference) > 0 {
		cmdargs = append(cmdargs, "--reference", flags.Reference)
	}
	if len(flags.SeparateGit) > 0 {
		cmdargs = append(cmdargs, "--separate-git-dir", flags.SeparateGit)
	}
	if flags.Shared {
		cmdargs = append(cmdargs, "--shared")
	}
	if flags.SingleBranch {
		cmdargs = append(cmdargs, "--single-branch")
	}
	if len(flags.TemplateDirectory) > 0 {
		cmdargs = append(cmdargs, "--template", flags.TemplateDirectory)
	}
	if len(flags.Upload) > 0 {
		cmdargs = append(cmdargs, "--upload-pack", flags.Upload)
	}
	if flags.Verbose {
		cmdargs = append(cmdargs, "--verbose")
	}

	// Now args
	cmdargs = append(cmdargs, args...)
	cmd := subprocess.ExecCommand("git", cmdargs...)

	// Assign all streams direct
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("Failed to start git clone: %v", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("git clone failed: %v", err)
	}

	return nil
}

// CachedRemoteRefs returns the list of branches & tags for a remote which are
// currently cached locally. No remote request is made to verify them.
func CachedRemoteRefs(remoteName string) ([]*Ref, error) {
	var ret []*Ref
	cmd := subprocess.ExecCommand("git", "show-ref")

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
	return ret, cmd.Wait()
}

// RemoteRefs returns a list of branches & tags for a remote by actually
// accessing the remote vir git ls-remote
func RemoteRefs(remoteName string) ([]*Ref, error) {
	var ret []*Ref
	cmd := subprocess.ExecCommand("git", "ls-remote", "--heads", "--tags", "-q", remoteName)

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
	return ret, cmd.Wait()
}

// GetTrackedFiles returns a list of files which are tracked in Git which match
// the pattern specified (standard wildcard form)
// Both pattern and the results are relative to the current working directory, not
// the root of the repository
func GetTrackedFiles(pattern string) ([]string, error) {
	safePattern := sanitizePattern(pattern)
	rootWildcard := len(safePattern) < len(pattern) && strings.ContainsRune(safePattern, '*')

	var ret []string
	cmd := subprocess.ExecCommand("git",
		"-c", "core.quotepath=false", // handle special chars in filenames
		"ls-files",
		"--cached", // include things which are staged but not committed right now
		"--",       // no ambiguous patterns
		safePattern)

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to call git ls-files: %v", err)
	}
	cmd.Start()
	scanner := bufio.NewScanner(outp)
	for scanner.Scan() {
		line := scanner.Text()

		// If the given pattern is a root wildcard, skip all files which
		// are not direct descendants of the repository's root.
		//
		// This matches the behavior of how .gitattributes performs
		// filename matches.
		if rootWildcard && filepath.Dir(line) != "." {
			continue
		}

		ret = append(ret, strings.TrimSpace(line))
	}
	return ret, cmd.Wait()
}

func sanitizePattern(pattern string) string {
	if strings.HasPrefix(pattern, "/") {
		return pattern[1:]
	}

	return pattern
}

// GetFilesChanged returns a list of files which were changed, either between 2
// commits, or at a single commit if you only supply one argument and a blank
// string for the other
func GetFilesChanged(from, to string) ([]string, error) {
	var files []string
	args := []string{
		"-c", "core.quotepath=false", // handle special chars in filenames
		"diff-tree",
		"--no-commit-id",
		"--name-only",
		"-r",
	}

	if len(from) > 0 {
		args = append(args, from)
	}
	if len(to) > 0 {
		args = append(args, to)
	}
	args = append(args, "--") // no ambiguous patterns

	cmd := subprocess.ExecCommand("git", args...)
	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to call git diff: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("Failed to start git diff: %v", err)
	}
	scanner := bufio.NewScanner(outp)
	for scanner.Scan() {
		files = append(files, strings.TrimSpace(scanner.Text()))
	}
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("Git diff failed: %v", err)
	}

	return files, err
}

// IsFileModified returns whether the filepath specified is modified according
// to `git status`. A file is modified if it has uncommitted changes in the
// working copy or the index. This includes being untracked.
func IsFileModified(filepath string) (bool, error) {

	args := []string{
		"-c", "core.quotepath=false", // handle special chars in filenames
		"status",
		"--porcelain",
		"--", // separator in case filename ambiguous
		filepath,
	}
	cmd := subprocess.ExecCommand("git", args...)
	outp, err := cmd.StdoutPipe()
	if err != nil {
		return false, lfserrors.Wrap(err, "Failed to call git status")
	}
	if err := cmd.Start(); err != nil {
		return false, lfserrors.Wrap(err, "Failed to start git status")
	}
	matched := false
	for scanner := bufio.NewScanner(outp); scanner.Scan(); {
		line := scanner.Text()
		// Porcelain format is "<I><W> <filename>"
		// Where <I> = index status, <W> = working copy status
		if len(line) > 3 {
			// Double-check even though should be only match
			if strings.TrimSpace(line[3:]) == filepath {
				matched = true
				// keep consuming output to exit cleanly
				// will typically fall straight through anyway due to 1 line output
			}
		}
	}
	if err := cmd.Wait(); err != nil {
		return false, lfserrors.Wrap(err, "Git status failed")
	}

	return matched, nil
}
