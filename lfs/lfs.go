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

const Version = "0.5.1"

var (
	LargeSizeThreshold = 5 * 1024 * 1024
	TempDir            = filepath.Join(os.TempDir(), "git-lfs")
	UserAgent          string
	LocalWorkingDir    string
	LocalGitDir        string
	LocalMediaDir      string
	LocalLogDir        string
	checkedTempDir     string
)

func TempFile(prefix string) (*os.File, error) {
	if checkedTempDir != TempDir {
		if err := os.MkdirAll(TempDir, 0755); err != nil {
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

func LocalMediaPath(sha string) (string, error) {
	path := filepath.Join(LocalMediaDir, sha[0:2], sha[2:4])
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("Error trying to create local media directory in '%s': %s", path, err)
	}

	return filepath.Join(path, sha), nil
}

func Environ() []string {
	osEnviron := os.Environ()
	env := make([]string, 6, len(osEnviron)+6)
	env[0] = fmt.Sprintf("LocalWorkingDir=%s", LocalWorkingDir)
	env[1] = fmt.Sprintf("LocalGitDir=%s", LocalGitDir)
	env[2] = fmt.Sprintf("LocalMediaDir=%s", LocalMediaDir)
	env[3] = fmt.Sprintf("TempDir=%s", TempDir)
	env[4] = fmt.Sprintf("ConcurrentTransfers=%d", Config.ConcurrentTransfers())
	env[5] = fmt.Sprintf("BatchTransfer=%v", Config.BatchTransfer())

	for _, e := range osEnviron {
		if !strings.Contains(e, "GIT_") {
			continue
		}
		env = append(env, e)
	}

	return env
}

func InRepo() bool {
	return LocalWorkingDir != ""
}

func init() {
	var err error

	tracerx.DefaultKey = "GIT"
	tracerx.Prefix = "trace git-lfs: "

	LocalWorkingDir, LocalGitDir, err = resolveGitDir()
	if err == nil {
		LocalMediaDir = filepath.Join(LocalGitDir, "lfs", "objects")
		LocalLogDir = filepath.Join(LocalMediaDir, "logs")
		TempDir = filepath.Join(LocalGitDir, "lfs", "tmp")

		if err := os.MkdirAll(LocalMediaDir, 0755); err != nil {
			panic(fmt.Errorf("Error trying to create objects directory in '%s': %s", LocalMediaDir, err))
		}

		if err := os.MkdirAll(LocalLogDir, 0755); err != nil {
			panic(fmt.Errorf("Error trying to create log directory in '%s': %s", LocalLogDir, err))
		}

		if err := os.MkdirAll(TempDir, 0755); err != nil {
			panic(fmt.Errorf("Error trying to create temp directory in '%s': %s", TempDir, err))
		}

	}

	gitVersion, err := git.Config.Version()
	if err != nil {
		gitVersion = "unknown"
	}

	UserAgent = fmt.Sprintf("git-lfs/%s (GitHub; %s %s; git %s; go %s)", Version,
		runtime.GOOS,
		runtime.GOARCH,
		strings.Replace(gitVersion, "git version ", "", 1),
		strings.Replace(runtime.Version(), "go", "", 1))
}

func resolveGitDir() (string, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	return recursiveResolveGitDir(wd)
}

func recursiveResolveGitDir(dir string) (string, string, error) {
	var cleanDir = filepath.Clean(dir)
	if cleanDir[len(cleanDir)-1] == os.PathSeparator {
		return "", "", fmt.Errorf("Git repository not found")
	}

	if filepath.Base(dir) == gitExt {
		// We're in the `.git` directory.  Make no assumptions about the working directory.
		return "", dir, nil
	}

	gitDir := filepath.Join(dir, gitExt)
	info, err := os.Stat(gitDir)
	if err != nil {
		// Found neither a directory nor a file named `.git`.
		// Move one directory up.
		return recursiveResolveGitDir(filepath.Dir(dir))
	}

	if !info.IsDir() {
		// Found a file named `.git` (we're in a submodule).
		return resolveDotGitFile(gitDir)
	}

	// Found the `.git` directory.
	return dir, gitDir, nil
}

func resolveDotGitFile(file string) (string, string, error) {
	// The local working directory is the directory the `.git` file is located in.
	wd := filepath.Dir(file)

	// The `.git` file tells us where the submodules `.git` directory is.
	gitDir, err := processDotGitFile(file)
	if err != nil {
		return "", "", err
	}

	return wd, gitDir, nil
}

func processDotGitFile(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data := make([]byte, 512)
	n, err := f.Read(data)
	if err != nil {
		return "", err
	}

	contents := string(data[0:n])

	if !strings.HasPrefix(contents, gitPtrPrefix) {
		// The `.git` file has no entry telling us about gitdir.
		return "", nil
	}

	dir := strings.TrimSpace(strings.Split(contents, gitPtrPrefix)[1])

	if filepath.IsAbs(dir) {
		// The .git file contains an absolute path.
		return dir, nil
	}

	// The .git file contains a relative path.
	// Create an absolute path based on the directory the .git file is located in.
	absDir := filepath.Join(filepath.Dir(file), dir)

	return absDir, nil
}

const (
	gitExt       = ".git"
	gitPtrPrefix = "gitdir: "
)
