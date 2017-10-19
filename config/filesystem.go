package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

var (
	localWorkingDir    string
	LocalGitDir        string // parent of index / config / hooks etc
	LocalGitStorageDir string // parent of objects/lfs (may be same as LocalGitDir but may not)
	LocalReferenceDir  string // alternative local media dir (relative to clone reference repo)
	LocalLogDir        string
)

// Determins the LocalWorkingDir, LocalGitDir etc
func ResolveGitBasicDirs() {
	var err error
	LocalGitDir, localWorkingDir, err = git.GitAndRootDirs()
	if err == nil {
		// Make sure we've fully evaluated symlinks, failure to do consistently
		// can cause discrepancies
		LocalGitDir = tools.ResolveSymlinks(LocalGitDir)
		localWorkingDir = tools.ResolveSymlinks(localWorkingDir)

		LocalGitStorageDir = resolveGitStorageDir(LocalGitDir)
		LocalReferenceDir = resolveReferenceDir(LocalGitStorageDir)

	} else {
		errMsg := err.Error()
		tracerx.Printf("Error running 'git rev-parse': %s", errMsg)
		if !strings.Contains(errMsg, "Not a git repository") {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
		}
	}
}

func resolveReferenceDir(gitStorageDir string) string {
	cloneReferencePath := filepath.Join(gitStorageDir, "objects", "info", "alternates")
	if tools.FileExists(cloneReferencePath) {
		buffer, err := ioutil.ReadFile(cloneReferencePath)
		if err == nil {
			path := strings.TrimSpace(string(buffer[:]))
			referenceLfsStoragePath := filepath.Join(filepath.Dir(path), "lfs", "objects")
			if tools.DirExists(referenceLfsStoragePath) {
				return referenceLfsStoragePath
			}
		}
	}
	return ""
}

// From a git dir, get the location that objects are to be stored (we will store lfs alongside)
// Sometimes there is an additional level of redirect on the .git folder by way of a commondir file
// before you find object storage, e.g. 'git worktree' uses this. It redirects to gitdir either by GIT_DIR
// (during setup) or .git/git-dir: (during use), but this only contains the index etc, the objects
// are found in another git dir via 'commondir'.
func resolveGitStorageDir(gitDir string) string {
	commondirpath := filepath.Join(gitDir, "commondir")
	if tools.FileExists(commondirpath) && !tools.DirExists(filepath.Join(gitDir, "objects")) {
		// no git-dir: prefix in commondir
		storage, err := processGitRedirectFile(commondirpath, "")
		if err == nil {
			return storage
		}
	}
	return gitDir
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
