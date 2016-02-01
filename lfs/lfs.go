package lfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

const (
	Version            = "1.1.0"
	tempDirPerms       = 0755
	localMediaDirPerms = 0755
	localLogDirPerms   = 0755
)

var (
	LargeSizeThreshold = 5 * 1024 * 1024
	TempDir            = filepath.Join(os.TempDir(), "git-lfs")
	GitCommit          string
	UserAgent          string
	LocalWorkingDir    string
	LocalGitDir        string // parent of index / config / hooks etc
	LocalGitStorageDir string // parent of objects/lfs (may be same as LocalGitDir but may not)
	LocalMediaDir      string // root of lfs objects
	LocalObjectTempDir string // where temporarily downloading objects are stored
	LocalLogDir        string
	checkedTempDir     string
)

func TempFile(prefix string) (*os.File, error) {
	if checkedTempDir != TempDir {
		if err := os.MkdirAll(TempDir, tempDirPerms); err != nil {
			return nil, err
		}
		checkedTempDir = TempDir
	}

	return ioutil.TempFile(TempDir, prefix)
}

func ResetTempDir() error {
	checkedTempDir = ""
	return os.RemoveAll(TempDir)
}

func localMediaDirNoCreate(sha string) string {
	return filepath.Join(LocalMediaDir, sha[0:2], sha[2:4])
}
func localMediaPathNoCreate(sha string) string {
	return filepath.Join(localMediaDirNoCreate(sha), sha)
}

func LocalMediaPath(sha string) (string, error) {
	path := localMediaDirNoCreate(sha)
	if err := os.MkdirAll(path, localMediaDirPerms); err != nil {
		return "", fmt.Errorf("Error trying to create local media directory in '%s': %s", path, err)
	}

	return filepath.Join(path, sha), nil
}

func ObjectExistsOfSize(sha string, size int64) bool {
	path := localMediaPathNoCreate(sha)
	return FileExistsOfSize(path, size)
}

func Environ() []string {
	osEnviron := os.Environ()
	env := make([]string, 0, len(osEnviron)+7)
	env = append(env,
		fmt.Sprintf("LocalWorkingDir=%s", LocalWorkingDir),
		fmt.Sprintf("LocalGitDir=%s", LocalGitDir),
		fmt.Sprintf("LocalGitStorageDir=%s", LocalGitStorageDir),
		fmt.Sprintf("LocalMediaDir=%s", LocalMediaDir),
		fmt.Sprintf("TempDir=%s", TempDir),
		fmt.Sprintf("ConcurrentTransfers=%d", Config.ConcurrentTransfers()),
		fmt.Sprintf("BatchTransfer=%v", Config.BatchTransfer()),
	)

	for _, e := range osEnviron {
		if !strings.Contains(e, "GIT_") {
			continue
		}
		env = append(env, e)
	}

	return env
}

func InRepo() bool {
	return LocalGitDir != ""
}

func ResolveDirs() {
	var err error
	LocalGitDir, LocalWorkingDir, err = git.GitAndRootDirs()
	if err == nil {
		LocalGitStorageDir = resolveGitStorageDir(LocalGitDir)
		LocalMediaDir = filepath.Join(LocalGitStorageDir, "lfs", "objects")
		LocalLogDir = filepath.Join(LocalMediaDir, "logs")
		TempDir = filepath.Join(LocalGitDir, "lfs", "tmp") // temp files per worktree
		if err := os.MkdirAll(LocalMediaDir, localMediaDirPerms); err != nil {
			panic(fmt.Errorf("Error trying to create objects directory in '%s': %s", LocalMediaDir, err))
		}

		if err := os.MkdirAll(LocalLogDir, localLogDirPerms); err != nil {
			panic(fmt.Errorf("Error trying to create log directory in '%s': %s", LocalLogDir, err))
		}

		LocalObjectTempDir = filepath.Join(TempDir, "objects")
		if err := os.MkdirAll(LocalObjectTempDir, tempDirPerms); err != nil {
			panic(fmt.Errorf("Error trying to create temp directory in '%s': %s", TempDir, err))
		}
	} else {
		errMsg := err.Error()
		tracerx.Printf("Error running 'git rev-parse': %s", errMsg)
		if !strings.Contains(errMsg, "Not a git repository") {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
		}
	}
}

func init() {
	tracerx.DefaultKey = "GIT"
	tracerx.Prefix = "trace git-lfs: "

	ResolveDirs()

	gitCommit := ""
	if len(GitCommit) > 0 {
		gitCommit = "; git " + GitCommit
	}
	UserAgent = fmt.Sprintf("git-lfs/%s (GitHub; %s %s; go %s%s)",
		Version,
		runtime.GOOS,
		runtime.GOARCH,
		strings.Replace(runtime.Version(), "go", "", 1),
		gitCommit,
	)
}

func processGitRedirectFile(file, prefix string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	contents := string(data)
	var dir string
	if len(prefix) > 0 {
		if !strings.HasPrefix(contents, prefix) {
			// Prefix required & not found
			return "", nil
		}
		dir = strings.TrimSpace(contents[len(prefix):])
	} else {
		dir = strings.TrimSpace(contents)
	}

	if !filepath.IsAbs(dir) {
		// The .git file contains a relative path.
		// Create an absolute path based on the directory the .git file is located in.
		dir = filepath.Join(filepath.Dir(file), dir)
	}

	return dir, nil
}

// From a git dir, get the location that objects are to be stored (we will store lfs alongside)
// Sometimes there is an additional level of redirect on the .git folder by way of a commondir file
// before you find object storage, e.g. 'git worktree' uses this. It redirects to gitdir either by GIT_DIR
// (during setup) or .git/git-dir: (during use), but this only contains the index etc, the objects
// are found in another git dir via 'commondir'.
func resolveGitStorageDir(gitDir string) string {
	commondirpath := filepath.Join(gitDir, "commondir")
	if FileExists(commondirpath) && !DirExists(filepath.Join(gitDir, "objects")) {
		// no git-dir: prefix in commondir
		storage, err := processGitRedirectFile(commondirpath, "")
		if err == nil {
			return storage
		}
	}
	return gitDir
}

const (
	gitExt       = ".git"
	gitPtrPrefix = "gitdir: "
)

// AllLocalObjects returns a slice of the the objects stored in the local LFS store
// This does not necessarily mean referenced by commits, just stored
// Note: reports final SHA only, extensions are ignored
func AllLocalObjects() []*Pointer {
	c := AllLocalObjectsChan()
	ret := make([]*Pointer, 0, 100)
	for p := range c {
		ret = append(ret, p)
	}
	return ret
}

// AllLocalObjectsChan returns a channel of all the objects stored in the local LFS store
// This does not necessarily mean referenced by commits, just stored
// You should not alter the store until this channel is closed
// Note: reports final SHA only, extensions are ignored
func AllLocalObjectsChan() <-chan *Pointer {
	ret := make(chan *Pointer, chanBufSize)

	go func() {
		defer close(ret)

		scanStorageDir(LocalMediaDir, ret)
	}()
	return ret
}

func scanStorageDir(dir string, c chan *Pointer) {
	// ioutil.ReadDir and filepath.Walk do sorting which is unnecessary & inefficient
	dirf, err := os.Open(dir)
	if err != nil {
		return
	}
	defer dirf.Close()

	direntries, err := dirf.Readdir(0)
	if err != nil {
		tracerx.Printf("Problem with Readdir in %v: %v", dir, err)
		return
	}
	for _, dirfi := range direntries {
		if dirfi.IsDir() {
			subpath := filepath.Join(dir, dirfi.Name())
			scanStorageDir(subpath, c)
		} else {
			// Make sure it's really an object file & not .DS_Store etc
			if oidRE.MatchString(dirfi.Name()) {
				c <- NewPointer(dirfi.Name(), dirfi.Size(), nil)
			}
		}
	}

}
