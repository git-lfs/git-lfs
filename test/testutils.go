package test

// Utility functions for more complex go tests
// Need to be in a separate test package so they can be imported anywhere
// Also can't add _test.go suffix to exclude from main build (import doesn't work)

// To avoid import cycles, append "_test" to the package statement of any test using
// this package and use "import . original/package/name" to get the same visibility
// as if the test was in the same package (as usual)

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
)

type RepoType int

const (
	// Normal repo with working copy
	RepoTypeNormal = RepoType(iota)
	// Bare repo (no working copy)
	RepoTypeBare = RepoType(iota)
	// Repo with working copy but git dir is separate
	RepoTypeSeparateDir = RepoType(iota)
)

type RepoCreateSettings struct {
	RepoType RepoType
}

type Repo struct {
	// Path to the repo, working copy if non-bare
	Path string
	// Path to the git dir
	GitDir string
	// Paths to remotes
	Remotes map[string]*Repo
	// Settings used to create this repo
	Settings *RepoCreateSettings
	// Previous dir for pushd
	popDir string
	// Testing context
	t *testing.T
}

// Change to repo dir but save current dir
func (r *Repo) Pushd() {
	if r.popDir != "" {
		r.t.Fatalf("Cannot Pushd twice")
	}
	oldwd, err := os.Getwd()
	if err != nil {
		r.t.Fatalf("Can't get cwd %v", err)
	}
	err = os.Chdir(r.Path)
	if err != nil {
		r.t.Fatalf("Can't chdir %v", err)
	}
	r.popDir = oldwd
}

func (r *Repo) Popd() {
	if r.popDir != "" {
		err := os.Chdir(r.Path)
		if err != nil {
			r.t.Fatalf("Can't chdir %v", err)
		}
		r.popDir = ""
	}
}

func (r *Repo) Cleanup() {

	// pop out if necessary
	r.Popd()

	// Make sure cwd isn't inside a path we're going to delete
	oldwd, err := os.Getwd()
	if err == nil {
		if strings.HasPrefix(oldwd, r.Path) ||
			strings.HasPrefix(oldwd, r.GitDir) {
			os.Chdir(os.TempDir())
		}
	}

	if r.GitDir != "" {
		os.RemoveAll(r.GitDir)
		r.GitDir = ""
	}
	if r.Path != "" {
		os.RemoveAll(r.Path)
		r.Path = ""
	}
	for _, remote := range r.Remotes {
		remote.Cleanup()
	}
	r.Remotes = nil
}

func NewRepo(t *testing.T) *Repo {
	return NewCustomRepo(t, &RepoCreateSettings{RepoType: RepoTypeNormal})
}
func NewCustomRepo(t *testing.T, settings *RepoCreateSettings) *Repo {
	ret := &Repo{
		Settings: settings,
		Remotes:  make(map[string]*Repo),
		t:        t}

	path, err := ioutil.TempDir("", "lfsRepo")
	if err != nil {
		t.Fatalf("Can't create temp dir for git repo: %v", err)
	}
	ret.Path = path
	args := []string{"init"}
	switch settings.RepoType {
	case RepoTypeBare:
		args = append(args, "--bare")
		ret.GitDir = ret.Path
	case RepoTypeSeparateDir:
		gitdir, err := ioutil.TempDir("", "lfstestgitdir")
		if err != nil {
			ret.Cleanup()
			t.Fatalf("Can't create temp dir for git repo: %v", err)
		}
		args = append(args, "--separate-dir", gitdir)
		ret.GitDir = gitdir
	default:
		ret.GitDir = filepath.Join(ret.Path, ".git")
	}
	args = append(args, path)
	cmd := exec.Command("git", args...)
	err = cmd.Run()
	if err != nil {
		ret.Cleanup()
		t.Fatalf("Unable to create git repo at %v: %v", path, err)
	}
	return ret
}

// Simplistic fire & forget running of git command - returns combined output
func RunGitCommand(t *testing.T, failureCheck bool, args ...string) string {
	outp, err := exec.Command("git", args...).CombinedOutput()
	if failureCheck && err != nil {
		t.Fatalf("Error running git command 'git %v': %v", strings.Join(args, " "), err)
	}
	return string(outp)

}

// Input data for a single file in a commit
type FileInput struct {
	// Name of file (required)
	Filename string
	// Size of file (required)
	Size int64
	// Input data (optional - if nil, placeholder data of Size will be created)
	Data io.Reader
}

// Input for defining commits for test repo
type CommitInput struct {
	// Date that we should commit on (optional, leave blank for 'now')
	CommitDate time.Time
	// List of files to include in this commit
	Files []*FileInput
	// List of parent branches (all branches must have been created in a previous NewBranch or be master)
	// Can be omitted to just use the parent of the previous commit
	ParentBranches []string
	// Name of a new branch we should create at this commit (optional - master not required)
	NewBranch string
	// Names of any tags we should create at this commit (optional)
	Tags []string
	// Name of committer
	CommitterName string
	// Email of committer
	CommitterEmail string
}

// Output struct with details of commits created for test
type CommitOutput struct {
	Sha     string
	Parents []string
	Files   []*lfs.Pointer
}

func formatGitDate(tm time.Time) string {
	// Git format is "Fri Jun 21 20:26:41 2013 +0900" but no zero-leading for day
	return tm.Format("Mon Jan 2 15:04:05 2006 -0700")
}

func commitAtDate(atDate time.Time, committerName, committerEmail, msg string) error {
	var args []string
	if committerName != "" && committerEmail != "" {
		args = append(args, "-c", fmt.Sprintf("user.name=%v", committerName))
		args = append(args, "-c", fmt.Sprintf("user.email=%v", committerEmail))
	}
	args = append(args, "commit", "--allow-empty", "-m", msg)
	cmd := exec.Command("git", args...)
	env := os.Environ()
	// set GIT_COMMITTER_DATE environment var e.g. "Fri Jun 21 20:26:41 2013 +0900"
	if atDate.IsZero() {
		env = append(env, "GIT_COMMITTER_DATE=")
	} else {
		env = append(env, fmt.Sprintf("GIT_COMMITTER_DATE=%v", formatGitDate(atDate)))
	}
	cmd.Env = env
	return cmd.Run()
}

func (repo *Repo) AddCommits(inputs []*CommitInput) []*CommitOutput {
	if repo.Settings.RepoType == RepoTypeBare {
		repo.t.Fatalf("Cannot use SetupRepo on a bare repo; clone it & push changes instead")
	}

	// Change to repo working dir
	oldwd, err := os.Getwd()
	if err != nil {
		repo.t.Fatalf("Can't get cwd %v", err)
	}
	err = os.Chdir(repo.Path)
	if err != nil {
		repo.t.Fatalf("Can't chdir to repo %v", err)
	}
	// Used to check whether we need to checkout another commit before
	lastBranch := "master"
	outputs := make([]*CommitOutput, 0, len(inputs))

	for i, input := range inputs {
		output := &CommitOutput{}
		// first, are we on the correct branch
		if len(input.ParentBranches) > 0 {
			if input.ParentBranches[0] != lastBranch {
				RunGitCommand(repo.t, true, "checkout", input.ParentBranches[0])
				lastBranch = input.ParentBranches[0]
			}
		}
		// Is this a merge?
		if len(input.ParentBranches) > 1 {
			// Always take the *other* side in a merge so we adopt changes
			// also don't automatically commit, we'll do that below
			args := []string{"merge", "--no-ff", "--no-commit", "--strategy-option=theirs"}
			args = append(args, input.ParentBranches[1:]...)
			RunGitCommand(repo.t, false, args...)
		} else if input.NewBranch != "" {
			RunGitCommand(repo.t, true, "checkout", "-b", input.NewBranch)
			lastBranch = input.NewBranch
		}
		// Any files to write?
		for fi, infile := range input.Files {
			inputData := infile.Data
			if inputData == nil {
				// Different data for each file but deterministic
				inputData = NewPlaceholderDataReader(int64(i*fi), infile.Size)
			}
			cleaned, err := lfs.PointerClean(inputData, infile.Filename, infile.Size, nil)
			if err != nil {
				repo.t.Errorf("Error creating pointer file: %v", err)
				continue
			}

			output.Files = append(output.Files, cleaned.Pointer)
			// Write pointer to local filename for adding (not using clean filter)
			os.MkdirAll(filepath.Dir(infile.Filename), 0755)
			f, err := os.Create(infile.Filename)
			if err != nil {
				repo.t.Errorf("Error creating pointer file: %v", err)
				continue
			}
			_, err = cleaned.Pointer.Encode(f)
			if err != nil {
				f.Close()
				repo.t.Errorf("Error encoding pointer file: %v", err)
				continue
			}
			f.Close() // early close in a loop, don't defer
			RunGitCommand(repo.t, true, "add", infile.Filename)

		}
		// Now commit
		commitAtDate(input.CommitDate, input.CommitterName, input.CommitterEmail,
			fmt.Sprintf("Test commit %d", i))
		commit, err := git.GetCommitSummary("HEAD")
		if err != nil {
			repo.t.Fatalf("Error determining commit SHA: %v", err)
		}

		// tags
		for _, tag := range input.Tags {
			// Use annotated tags, assume full release tags (also tag objects have edge cases)
			RunGitCommand(repo.t, true, "tag", "-a", "-m", "Added tag", tag)
		}

		output.Sha = commit.Sha
		output.Parents = commit.Parents
		outputs = append(outputs, output)
	}

	// Restore cwd
	err = os.Chdir(oldwd)
	if err != nil {
		repo.t.Fatalf("Can't restore old cwd %v", err)
	}

	return outputs
}

// Add a new remote (generate a path for it to live in, will be cleaned up)
func (r *Repo) AddRemote(name string) *Repo {
	if _, exists := r.Remotes[name]; exists {
		r.t.Fatalf("Remote %v already exists", name)
	}
	remote := NewCustomRepo(r.t, &RepoCreateSettings{RepoTypeBare})
	r.Remotes[name] = remote
	RunGitCommand(r.t, true, "remote", "add", name, remote.Path)
	return remote
}

// Just a psuedo-random stream of bytes (not cryptographic)
// Calls RNG a bit less often than using rand.Source directly
type PlaceholderDataReader struct {
	source    rand.Source
	bytesLeft int64
}

func NewPlaceholderDataReader(seed, size int64) *PlaceholderDataReader {
	return &PlaceholderDataReader{rand.NewSource(seed), size}
}

func (r *PlaceholderDataReader) Read(p []byte) (int, error) {
	c := len(p)
	i := 0
	for i < c && r.bytesLeft > 0 {
		// Use all 8 bytes of the 64-bit random number
		val64 := r.source.Int63()
		for j := 0; j < 8 && i < c && r.bytesLeft > 0; j++ {
			// Duplicate this byte 16 times (faster)
			for k := 0; k < 16 && r.bytesLeft > 0; k++ {
				p[i] = byte(val64)
				i++
				r.bytesLeft--
			}
			// Next byte from the 8-byte number
			val64 = val64 >> 8
		}
	}
	var err error
	if r.bytesLeft == 0 {
		err = io.EOF
	}
	return i, err
}
