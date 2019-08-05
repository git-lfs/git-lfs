// Package git contains various commands that shell out to git
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package git

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	RefTypeHEAD         = RefType(iota) // current checkout
	RefTypeOther        = RefType(iota) // stash or unknown

	// A ref which can be used as a placeholder for before the first commit
	// Equivalent to git mktree < /dev/null, useful for diffing before first commit
	RefBeforeFirstCommit = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
)

type IndexStage int

const (
	IndexStageDefault IndexStage = iota
	IndexStageBase
	IndexStageOurs
	IndexStageTheirs
)

// Prefix returns the given RefType's prefix, "refs/heads", "ref/remotes",
// etc. It returns an additional value of either true/false, whether or not this
// given ref type has a prefix.
//
// If the RefType is unrecognized, Prefix() will panic.
func (t RefType) Prefix() (string, bool) {
	switch t {
	case RefTypeLocalBranch:
		return "refs/heads", true
	case RefTypeRemoteBranch:
		return "refs/remotes", true
	case RefTypeLocalTag:
		return "refs/tags", true
	default:
		return "", false
	}
}

func ParseRef(absRef, sha string) *Ref {
	r := &Ref{Sha: sha}
	if strings.HasPrefix(absRef, "refs/heads/") {
		r.Name = absRef[11:]
		r.Type = RefTypeLocalBranch
	} else if strings.HasPrefix(absRef, "refs/tags/") {
		r.Name = absRef[10:]
		r.Type = RefTypeLocalTag
	} else if strings.HasPrefix(absRef, "refs/remotes/") {
		r.Name = absRef[13:]
		r.Type = RefTypeRemoteBranch
	} else {
		r.Name = absRef
		if absRef == "HEAD" {
			r.Type = RefTypeHEAD
		} else {
			r.Type = RefTypeOther
		}
	}
	return r
}

// A git reference (branch, tag etc)
type Ref struct {
	Name string
	Type RefType
	Sha  string
}

// Refspec returns the fully-qualified reference name (including remote), i.e.,
// for a remote branch called 'my-feature' on remote 'origin', this function
// will return:
//
//   refs/remotes/origin/my-feature
func (r *Ref) Refspec() string {
	if r == nil {
		return ""
	}

	prefix, ok := r.Type.Prefix()
	if ok {
		return fmt.Sprintf("%s/%s", prefix, r.Name)
	}

	return r.Name
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

// Prepend Git config instructions to disable Git LFS filter
func gitConfigNoLFS(args ...string) []string {
	// Before git 2.8, setting filters to blank causes lots of warnings, so use cat instead (slightly slower)
	// Also pre 2.2 it failed completely. We used to use it anyway in git 2.2-2.7 and
	// suppress the messages in stderr, but doing that with standard StderrPipe suppresses
	// the git clone output (git thinks it's not a terminal) and makes it look like it's
	// not working. You can get around that with https://github.com/kr/pty but that
	// causes difficult issues with passing through Stdin for login prompts
	// This way is simpler & more practical.
	filterOverride := ""
	if !IsGitVersionAtLeast("2.8.0") {
		filterOverride = "cat"
	}

	return append([]string{
		"-c", fmt.Sprintf("filter.lfs.smudge=%v", filterOverride),
		"-c", fmt.Sprintf("filter.lfs.clean=%v", filterOverride),
		"-c", "filter.lfs.process=",
		"-c", "filter.lfs.required=false",
	}, args...)
}

// Invoke Git with disabled LFS filters
func gitNoLFS(args ...string) *subprocess.Cmd {
	return subprocess.ExecCommand("git", gitConfigNoLFS(args...)...)
}

func gitNoLFSSimple(args ...string) (string, error) {
	return subprocess.SimpleExec("git", gitConfigNoLFS(args...)...)
}

func gitNoLFSBuffered(args ...string) (*subprocess.BufferedCmd, error) {
	return subprocess.BufferedExec("git", gitConfigNoLFS(args...)...)
}

// Invoke Git with enabled LFS filters
func git(args ...string) *subprocess.Cmd {
	return subprocess.ExecCommand("git", args...)
}

func gitSimple(args ...string) (string, error) {
	return subprocess.SimpleExec("git", args...)
}

func gitBuffered(args ...string) (*subprocess.BufferedCmd, error) {
	return subprocess.BufferedExec("git", args...)
}

func CatFile() (*subprocess.BufferedCmd, error) {
	return gitNoLFSBuffered("cat-file", "--batch-check")
}

func DiffIndex(ref string, cached bool) (*bufio.Scanner, error) {
	args := []string{"diff-index", "-M"}
	if cached {
		args = append(args, "--cached")
	}
	args = append(args, ref)

	cmd, err := gitBuffered(args...)
	if err != nil {
		return nil, err
	}
	if err = cmd.Stdin.Close(); err != nil {
		return nil, err
	}

	return bufio.NewScanner(cmd.Stdout), nil
}

func HashObject(r io.Reader) (string, error) {
	cmd := gitNoLFS("hash-object", "--stdin")
	cmd.Stdin = r
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error building Git blob OID: %s", err)
	}

	return string(bytes.TrimSpace(out)), nil
}

func Log(args ...string) (*subprocess.BufferedCmd, error) {
	logArgs := append([]string{"log"}, args...)
	return gitNoLFSBuffered(logArgs...)
}

func LsRemote(remote, remoteRef string) (string, error) {
	if remote == "" {
		return "", errors.New("remote required")
	}
	if remoteRef == "" {
		return gitNoLFSSimple("ls-remote", remote)

	}
	return gitNoLFSSimple("ls-remote", remote, remoteRef)
}

func LsTree(ref string) (*subprocess.BufferedCmd, error) {
	return gitNoLFSBuffered(
		"ls-tree",
		"-r",          // recurse
		"-l",          // report object size (we'll need this)
		"-z",          // null line termination
		"--full-tree", // start at the root regardless of where we are in it
		ref,
	)
}

func ResolveRef(ref string) (*Ref, error) {
	outp, err := gitNoLFSSimple("rev-parse", ref, "--symbolic-full-name", ref)
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
		fullref.Name = lines[0]
		fullref.Sha = lines[0]
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

func (c *Configuration) CurrentRemoteRef() (*Ref, error) {
	remoteref, err := c.RemoteRefNameForCurrentBranch()
	if err != nil {
		return nil, err
	}

	return ResolveRef(remoteref)
}

// RemoteRefForCurrentBranch returns the full remote ref (refs/remotes/{remote}/{remotebranch})
// that the current branch is tracking.
func (c *Configuration) RemoteRefNameForCurrentBranch() (string, error) {
	ref, err := CurrentRef()
	if err != nil {
		return "", err
	}

	if ref.Type == RefTypeHEAD || ref.Type == RefTypeOther {
		return "", errors.New("not on a branch")
	}

	remote := c.RemoteForBranch(ref.Name)
	if remote == "" {
		return "", fmt.Errorf("remote not found for branch %q", ref.Name)
	}

	remotebranch := c.RemoteBranchForLocalBranch(ref.Name)

	return fmt.Sprintf("refs/remotes/%s/%s", remote, remotebranch), nil
}

// RemoteForBranch returns the remote name that a given local branch is tracking (blank if none)
func (c *Configuration) RemoteForBranch(localBranch string) string {
	return c.Find(fmt.Sprintf("branch.%s.remote", localBranch))
}

// RemoteBranchForLocalBranch returns the name (only) of the remote branch that the local branch is tracking
// If no specific branch is configured, returns local branch name
func (c *Configuration) RemoteBranchForLocalBranch(localBranch string) string {
	// get remote ref to track, may not be same name
	merge := c.Find(fmt.Sprintf("branch.%s.merge", localBranch))
	if strings.HasPrefix(merge, "refs/heads/") {
		return merge[11:]
	} else {
		return localBranch
	}
}

func RemoteList() ([]string, error) {
	cmd := gitNoLFS("remote")

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
	cmd := gitNoLFS("show-ref", "--heads", "--tags")

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

// UpdateRef moves the given ref to a new sha with a given reason (and creates a
// reflog entry, if a "reason" was provided). It returns an error if any were
// encountered.
func UpdateRef(ref *Ref, to []byte, reason string) error {
	return UpdateRefIn("", ref, to, reason)
}

// UpdateRef moves the given ref to a new sha with a given reason (and creates a
// reflog entry, if a "reason" was provided). It operates within the given
// working directory "wd". It returns an error if any were encountered.
func UpdateRefIn(wd string, ref *Ref, to []byte, reason string) error {
	args := []string{"update-ref", ref.Refspec(), hex.EncodeToString(to)}
	if len(reason) > 0 {
		args = append(args, "-m", reason)
	}

	cmd := gitNoLFS(args...)
	cmd.Dir = wd

	return cmd.Run()
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
	u, _ := url.Parse(remote)
	if u == nil || u.Scheme == "" {
		// This is either an invalid remote name (maybe the user made a typo
		// when selecting a named remote) or a bare SSH URL like
		// "x@y.com:path/to/resource.git". Guess that this is a URL in the latter
		// form if the string contains a colon ":", and an invalid remote if it
		// does not.
		if strings.Contains(remote, ":") {
			return nil
		} else {
			return fmt.Errorf("Invalid remote name: %q", remote)
		}
	}

	switch u.Scheme {
	case "ssh", "http", "https", "git", "file":
		return nil
	default:
		return fmt.Errorf("Invalid remote url protocol %q in %q", u.Scheme, remote)
	}
}

func UpdateIndexFromStdin() *subprocess.Cmd {
	return git("update-index", "-q", "--refresh", "--stdin")
}

// RecentBranches returns branches with commit dates on or after the given date/time
// Return full Ref type for easier detection of duplicate SHAs etc
// since: refs with commits on or after this date will be included
// includeRemoteBranches: true to include refs on remote branches
// onlyRemote: set to non-blank to only include remote branches on a single remote
func RecentBranches(since time.Time, includeRemoteBranches bool, onlyRemote string) ([]*Ref, error) {
	cmd := gitNoLFS("for-each-ref",
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
			if reftype == RefTypeRemoteBranch {
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
	cmd := gitNoLFS("show", "-s",
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
	cmd := gitNoLFS("rev-parse", "--git-dir", "--show-toplevel")
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

	absGitDir, err := canonicalizeDir(paths[0])
	if err != nil {
		return "", "", fmt.Errorf("Error converting %q to absolute: %s", paths[0], err)
	}

	if pathLen == 1 || len(paths[1]) == 0 {
		return absGitDir, "", nil
	}

	absRootDir, err := canonicalizeDir(paths[1])
	return absGitDir, absRootDir, err
}

func canonicalizeDir(path string) (string, error) {
	if len(path) > 0 {
		path, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		return filepath.EvalSymlinks(path)
	}
	return "", nil
}

func RootDir() (string, error) {
	cmd := gitNoLFS("rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to call git rev-parse --show-toplevel: %v %v", err, string(out))
	}

	path := strings.TrimSpace(string(out))
	path, err = tools.TranslateCygwinPath(path)
	if err != nil {
		return "", err
	}
	return canonicalizeDir(path)
}

func GitDir() (string, error) {
	cmd := gitNoLFS("rev-parse", "--git-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to call git rev-parse --git-dir: %v %v", err, string(out))
	}
	path := strings.TrimSpace(string(out))
	path, err = tools.TranslateCygwinPath(path)
	if err != nil {
		return "", err
	}
	return canonicalizeDir(path)
}

func GitCommonDir() (string, error) {
	// Versions before 2.5.0 don't have the --git-common-dir option, since
	// it came in with worktrees, so just fall back to the main Git
	// directory.
	if !IsGitVersionAtLeast("2.5.0") {
		return GitDir()
	}

	cmd := gitNoLFS("rev-parse", "--git-common-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to call git rev-parse --git-dir: %v %v", err, string(out))
	}
	path := strings.TrimSpace(string(out))
	path, err = tools.TranslateCygwinPath(path)
	if err != nil {
		return "", err
	}
	return canonicalizeDir(path)
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

// IsBare returns whether or not a repository is bare. It requires that the
// current working directory is a repository.
//
// If there was an error determining whether or not the repository is bare, it
// will be returned.
func IsBare() (bool, error) {
	s, err := subprocess.SimpleExec(
		"git", "rev-parse", "--is-bare-repository")

	if err != nil {
		return false, err
	}

	return strconv.ParseBool(s)
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
	// --reference-if-able <repository>
	ReferenceIfAble string
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
	// --shallow-since <date>
	ShallowSince string
	// --shallow-since <date>
	ShallowExclude string
	// --shallow-submodules
	ShallowSubmodules bool
	// --no-shallow-submodules
	NoShallowSubmodules bool
	// jobs <n>
	Jobs int64
}

// CloneWithoutFilters clones a git repo but without the smudge filter enabled
// so that files in the working copy will be pointers and not real LFS data
func CloneWithoutFilters(flags CloneFlags, args []string) error {

	cmdargs := []string{"clone"}

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
	if len(flags.ReferenceIfAble) > 0 {
		cmdargs = append(cmdargs, "--reference-if-able", flags.ReferenceIfAble)
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
	if len(flags.ShallowSince) > 0 {
		cmdargs = append(cmdargs, "--shallow-since", flags.ShallowSince)
	}
	if len(flags.ShallowExclude) > 0 {
		cmdargs = append(cmdargs, "--shallow-exclude", flags.ShallowExclude)
	}
	if flags.ShallowSubmodules {
		cmdargs = append(cmdargs, "--shallow-submodules")
	}
	if flags.NoShallowSubmodules {
		cmdargs = append(cmdargs, "--no-shallow-submodules")
	}
	if flags.Jobs > -1 {
		cmdargs = append(cmdargs, "--jobs", strconv.FormatInt(flags.Jobs, 10))
	}

	// Now args
	cmdargs = append(cmdargs, args...)
	cmd := gitNoLFS(cmdargs...)

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

// Checkout performs an invocation of `git-checkout(1)` applying the given
// treeish, paths, and force option, if given.
//
// If any error was encountered, it will be returned immediately. Otherwise, the
// checkout has occurred successfully.
func Checkout(treeish string, paths []string, force bool) error {
	args := []string{"checkout"}
	if force {
		args = append(args, "--force")
	}

	if len(treeish) > 0 {
		args = append(args, treeish)
	}

	if len(paths) > 0 {
		args = append(args, append([]string{"--"}, paths...)...)
	}

	_, err := gitNoLFSSimple(args...)
	return err
}

// CachedRemoteRefs returns the list of branches & tags for a remote which are
// currently cached locally. No remote request is made to verify them.
func CachedRemoteRefs(remoteName string) ([]*Ref, error) {
	var ret []*Ref
	cmd := gitNoLFS("show-ref")

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

// Fetch performs a fetch with no arguments against the given remotes.
func Fetch(remotes ...string) error {
	if len(remotes) == 0 {
		return nil
	}

	var args []string
	if len(remotes) > 1 {
		args = []string{"--multiple", "--"}
	}
	args = append(args, remotes...)

	_, err := gitNoLFSSimple(append([]string{"fetch"}, args...)...)
	return err
}

// RemoteRefs returns a list of branches & tags for a remote by actually
// accessing the remote vir git ls-remote
func RemoteRefs(remoteName string) ([]*Ref, error) {
	var ret []*Ref
	cmd := gitNoLFS("ls-remote", "--heads", "--tags", "-q", remoteName)

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
			ret = append(ret, &Ref{name, RefTypeRemoteBranch, sha})
		}
	}
	return ret, cmd.Wait()
}

// AllRefs returns a slice of all references in a Git repository in the current
// working directory, or an error if those references could not be loaded.
func AllRefs() ([]*Ref, error) {
	return AllRefsIn("")
}

// AllRefs returns a slice of all references in a Git repository located in a
// the given working directory "wd", or an error if those references could not
// be loaded.
func AllRefsIn(wd string) ([]*Ref, error) {
	cmd := gitNoLFS(
		"for-each-ref", "--format=%(objectname)%00%(refname)")
	cmd.Dir = wd

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, lfserrors.Wrap(err, "cannot open pipe")
	}
	cmd.Start()

	refs := make([]*Ref, 0)

	scanner := bufio.NewScanner(outp)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "\x00", 2)
		if len(parts) != 2 {
			return nil, lfserrors.Errorf(
				"git: invalid for-each-ref line: %q", scanner.Text())
		}

		sha := parts[0]
		typ, name := ParseRefToTypeAndName(parts[1])

		refs = append(refs, &Ref{
			Name: name,
			Type: typ,
			Sha:  sha,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return refs, nil
}

// GetTrackedFiles returns a list of files which are tracked in Git which match
// the pattern specified (standard wildcard form)
// Both pattern and the results are relative to the current working directory, not
// the root of the repository
func GetTrackedFiles(pattern string) ([]string, error) {
	safePattern := sanitizePattern(pattern)
	rootWildcard := len(safePattern) < len(pattern) && strings.ContainsRune(safePattern, '*')

	var ret []string
	cmd := gitNoLFS(
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

	cmd := gitNoLFS(args...)
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
	cmd := git(args...)
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

// IsWorkingCopyDirty returns true if and only if the working copy in which the
// command was executed is dirty as compared to the index.
//
// If the status of the working copy could not be determined, an error will be
// returned instead.
func IsWorkingCopyDirty() (bool, error) {
	bare, err := IsBare()
	if bare || err != nil {
		return false, err
	}

	out, err := gitSimple("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(out) != 0, nil
}
