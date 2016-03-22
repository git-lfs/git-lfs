package lfs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/localstorage"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

const (
	Version            = "1.1.1"
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
	objects            *localstorage.LocalStorage
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

func LocalMediaPath(oid string) (string, error) {
	return objects.BuildObjectPath(oid)
}

func ObjectExistsOfSize(oid string, size int64) bool {
	path := objects.ObjectPath(oid)
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
		// Make sure we've fully evaluated symlinks, failure to do consistently
		// can cause discrepancies
		LocalGitDir = ResolveSymlinks(LocalGitDir)
		LocalWorkingDir = ResolveSymlinks(LocalWorkingDir)

		LocalGitStorageDir = resolveGitStorageDir(LocalGitDir)
		TempDir = filepath.Join(LocalGitDir, "lfs", "tmp") // temp files per worktree

		objs, err := localstorage.New(
			filepath.Join(LocalGitStorageDir, "lfs", "objects"),
			filepath.Join(TempDir, "objects"),
		)

		if err != nil {
			panic(fmt.Sprintf("Error trying to init LocalStorage: %s", err))
		}

		objects = objs
		LocalMediaDir = objs.RootDir
		LocalObjectTempDir = objs.TempDir
		LocalLogDir = filepath.Join(objs.RootDir, "logs")
		if err := os.MkdirAll(LocalLogDir, localLogDirPerms); err != nil {
			panic(fmt.Errorf("Error trying to create log directory in '%s': %s", LocalLogDir, err))
		}
	} else {
		errMsg := err.Error()
		tracerx.Printf("Error running 'git rev-parse': %s", errMsg)
		if !strings.Contains(errMsg, "Not a git repository") {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
		}
	}
}

func ClearTempObjects() error {
	if objects == nil {
		return nil
	}
	return objects.ClearTempObjects()
}

func ScanObjectsChan() <-chan localstorage.Object {
	return objects.ScanObjectsChan()
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

func traceHttpReq(req *http.Request) string {
	return fmt.Sprintf("%s %s", req.Method, strings.SplitN(req.URL.String(), "?", 2)[0])
}

// only used in tests
func AllObjects() []localstorage.Object {
	return objects.AllObjects()
}
