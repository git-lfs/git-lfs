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
	"sync"
	"time"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/localstorage"
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

// Callback interface (testing.T compatible)
type RepoCallback interface {
	// Fatalf reports error and fails
	Fatalf(format string, args ...interface{})
	// Errorf reports error and continues
	Errorf(format string, args ...interface{})
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
	// Test callback
	callback RepoCallback
}

// Change to repo dir but save current dir
func (r *Repo) Pushd() {
	if r.popDir != "" {
		r.callback.Fatalf("Cannot Pushd twice")
	}
	oldwd, err := os.Getwd()
	if err != nil {
		r.callback.Fatalf("Can't get cwd %v", err)
	}
	err = os.Chdir(r.Path)
	if err != nil {
		r.callback.Fatalf("Can't chdir %v", err)
	}
	r.popDir = oldwd
	if err := localstorage.ResolveDirs(); err != nil {
		r.callback.Fatalf("init error: %v", err)
	}
}

func (r *Repo) Popd() {
	if r.popDir != "" {
		err := os.Chdir(r.popDir)
		if err != nil {
			r.callback.Fatalf("Can't chdir %v", err)
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

// NewRepo creates a new git repo in a new temp dir
func NewRepo(callback RepoCallback) *Repo {
	return NewCustomRepo(callback, &RepoCreateSettings{RepoType: RepoTypeNormal})
}

// NewCustomRepo creates a new git repo in a new temp dir with more control over settings
func NewCustomRepo(callback RepoCallback, settings *RepoCreateSettings) *Repo {
	ret := &Repo{
		Settings: settings,
		Remotes:  make(map[string]*Repo),
		callback: callback}

	path, err := ioutil.TempDir("", "lfsRepo")
	if err != nil {
		callback.Fatalf("Can't create temp dir for git repo: %v", err)
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
			callback.Fatalf("Can't create temp dir for git repo: %v", err)
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
		callback.Fatalf("Unable to create git repo at %v: %v", path, err)
	}

	// Configure default user/email so not reliant on env
	ret.Pushd()
	RunGitCommand(callback, true, "config", "user.name", "Git LFS Tests")
	RunGitCommand(callback, true, "config", "user.email", "git-lfs@example.com")
	ret.Popd()

	return ret
}

// WrapRepo creates a new Repo instance for an existing git repo
func WrapRepo(c RepoCallback, path string) *Repo {
	return &Repo{Path: path, callback: c, Settings: &RepoCreateSettings{RepoType: RepoTypeNormal}}
}

// Simplistic fire & forget running of git command - returns combined output
func RunGitCommand(callback RepoCallback, failureCheck bool, args ...string) string {
	outp, err := exec.Command("git", args...).CombinedOutput()
	if failureCheck && err != nil {
		callback.Fatalf("Error running git command 'git %v': %v %v", strings.Join(args, " "), err, string(outp))
	}
	return string(outp)

}

// Input data for a single file in a commit
type FileInput struct {
	// Name of file (required)
	Filename string
	// Size of file (required)
	Size int64
	// Input data (optional, if provided will be source of data)
	DataReader io.Reader
	// Input data (optional, if provided will be source of data)
	Data string
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
		env = append(env, fmt.Sprintf("GIT_COMMITTER_DATE=%v", git.FormatGitDate(atDate)))
	}
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v %v", err, string(out))
	}
	return nil
}

func (repo *Repo) AddCommits(inputs []*CommitInput) []*CommitOutput {
	var storageOnce sync.Once

	if repo.Settings.RepoType == RepoTypeBare {
		repo.callback.Fatalf("Cannot use AddCommits on a bare repo; clone it & push changes instead")
	}

	// Change to repo working dir
	oldwd, err := os.Getwd()
	if err != nil {
		repo.callback.Fatalf("Can't get cwd %v", err)
	}
	err = os.Chdir(repo.Path)
	if err != nil {
		repo.callback.Fatalf("Can't chdir to repo %v", err)
	}
	// Used to check whether we need to checkout another commit before
	lastBranch := "master"
	outputs := make([]*CommitOutput, 0, len(inputs))
	// Deterministic sequence of seeds for file data
	seedSequence := rand.NewSource(0)
	for i, input := range inputs {
		output := &CommitOutput{}
		// first, are we on the correct branch
		if len(input.ParentBranches) > 0 {
			if input.ParentBranches[0] != lastBranch {
				RunGitCommand(repo.callback, true, "checkout", input.ParentBranches[0])
				lastBranch = input.ParentBranches[0]
			}
		}
		// Is this a merge?
		if len(input.ParentBranches) > 1 {
			// Always take the *other* side in a merge so we adopt changes
			// also don't automatically commit, we'll do that below
			args := []string{"merge", "--no-ff", "--no-commit", "--strategy-option=theirs"}
			args = append(args, input.ParentBranches[1:]...)
			RunGitCommand(repo.callback, false, args...)
		} else if input.NewBranch != "" {
			RunGitCommand(repo.callback, true, "checkout", "-b", input.NewBranch)
			lastBranch = input.NewBranch
		}
		// Any files to write?
		for _, infile := range input.Files {
			inputData := infile.DataReader
			if inputData == nil && infile.Data != "" {
				inputData = strings.NewReader(infile.Data)
			}
			if inputData == nil {
				// Different data for each file but deterministic
				inputData = NewPlaceholderDataReader(seedSequence.Int63(), infile.Size)
			}
			cleaned, err := lfs.PointerClean(inputData, infile.Filename, infile.Size, nil)
			if err != nil {
				repo.callback.Errorf("Error creating pointer file: %v", err)
				continue
			}
			// this only created the temp file, move to final location
			tmpfile := cleaned.Filename
			storageOnce.Do(func() { localstorage.ResolveDirs() })
			mediafile, err := lfs.LocalMediaPath(cleaned.Oid)
			if err != nil {
				repo.callback.Errorf("Unable to get local media path: %v", err)
				continue
			}
			if _, err := os.Stat(mediafile); err != nil {
				if err := os.Rename(tmpfile, mediafile); err != nil {
					repo.callback.Errorf("Unable to move %s to %s: %v", tmpfile, mediafile, err)
					continue
				}
			}

			output.Files = append(output.Files, cleaned.Pointer)
			// Write pointer to local filename for adding (not using clean filter)
			os.MkdirAll(filepath.Dir(infile.Filename), 0755)
			f, err := os.Create(infile.Filename)
			if err != nil {
				repo.callback.Errorf("Error creating pointer file: %v", err)
				continue
			}
			_, err = cleaned.Pointer.Encode(f)
			if err != nil {
				f.Close()
				repo.callback.Errorf("Error encoding pointer file: %v", err)
				continue
			}
			f.Close() // early close in a loop, don't defer
			RunGitCommand(repo.callback, true, "add", infile.Filename)

		}
		// Now commit
		err = commitAtDate(input.CommitDate, input.CommitterName, input.CommitterEmail,
			fmt.Sprintf("Test commit %d", i))
		if err != nil {
			repo.callback.Fatalf("Error committing: %v", err)
		}

		commit, err := git.GetCommitSummary("HEAD")
		if err != nil {
			repo.callback.Fatalf("Error determining commit SHA: %v", err)
		}

		// tags
		for _, tag := range input.Tags {
			// Use annotated tags, assume full release tags (also tag objects have edge cases)
			RunGitCommand(repo.callback, true, "tag", "-a", "-m", "Added tag", tag)
		}

		output.Sha = commit.Sha
		output.Parents = commit.Parents
		outputs = append(outputs, output)
	}

	// Restore cwd
	err = os.Chdir(oldwd)
	if err != nil {
		repo.callback.Fatalf("Can't restore old cwd %v", err)
	}

	return outputs
}

// Add a new remote (generate a path for it to live in, will be cleaned up)
func (r *Repo) AddRemote(name string) *Repo {
	if _, exists := r.Remotes[name]; exists {
		r.callback.Fatalf("Remote %v already exists", name)
	}
	remote := NewCustomRepo(r.callback, &RepoCreateSettings{RepoTypeBare})
	r.Remotes[name] = remote
	RunGitCommand(r.callback, true, "remote", "add", name, remote.Path)
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

// RefsByName implements sort.Interface for []*git.Ref based on name
type RefsByName []*git.Ref

func (a RefsByName) Len() int           { return len(a) }
func (a RefsByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RefsByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// WrappedPointersByOid implements sort.Interface for []*lfs.WrappedPointer based on oid
type WrappedPointersByOid []*lfs.WrappedPointer

func (a WrappedPointersByOid) Len() int           { return len(a) }
func (a WrappedPointersByOid) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a WrappedPointersByOid) Less(i, j int) bool { return a[i].Pointer.Oid < a[j].Pointer.Oid }

// PointersByOid implements sort.Interface for []*lfs.Pointer based on oid
type PointersByOid []*lfs.Pointer

func (a PointersByOid) Len() int           { return len(a) }
func (a PointersByOid) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PointersByOid) Less(i, j int) bool { return a[i].Oid < a[j].Oid }
