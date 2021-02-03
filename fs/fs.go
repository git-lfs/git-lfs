package fs

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

var oidRE = regexp.MustCompile(`\A[[:alnum:]]{64}`)

// Environment is a copy of a subset of the interface
// github.com/git-lfs/git-lfs/config.Environment.
//
// For more information, see config/environment.go.
type Environment interface {
	Get(key string) (val string, ok bool)
}

// Object represents a locally stored LFS object.
type Object struct {
	Oid  string
	Size int64
}

type Filesystem struct {
	GitStorageDir string   // parent of objects/lfs (may be same as GitDir but may not)
	LFSStorageDir string   // parent of lfs objects and tmp dirs. Default: ".git/lfs"
	ReferenceDirs []string // alternative local media dirs (relative to clone reference repo)
	lfsobjdir     string
	tmpdir        string
	logdir        string
	repoPerms     os.FileMode
	mu            sync.Mutex
}

func (f *Filesystem) EachObject(fn func(Object) error) error {
	var eachErr error
	tools.FastWalkDir(f.LFSObjectDir(), func(parentDir string, info os.FileInfo, err error) {
		if err != nil {
			eachErr = err
			return
		}
		if eachErr != nil || info.IsDir() {
			return
		}
		if oidRE.MatchString(info.Name()) {
			fn(Object{Oid: info.Name(), Size: info.Size()})
		}
	})
	return eachErr
}

func (f *Filesystem) ObjectExists(oid string, size int64) bool {
	return tools.FileExistsOfSize(f.ObjectPathname(oid), size)
}

func (f *Filesystem) ObjectPath(oid string) (string, error) {
	if len(oid) < 4 {
		return "", fmt.Errorf("too short object ID: %q", oid)
	}
	dir := f.localObjectDir(oid)
	if err := tools.MkdirAll(dir, f); err != nil {
		return "", fmt.Errorf("error trying to create local storage directory in %q: %s", dir, err)
	}
	return filepath.Join(dir, oid), nil
}

func (f *Filesystem) ObjectPathname(oid string) string {
	return filepath.Join(f.localObjectDir(oid), oid)
}

func (f *Filesystem) DecodePathname(path string) string {
	return string(DecodePathBytes([]byte(path)))
}

func (f *Filesystem) RepositoryPermissions(executable bool) os.FileMode {
	if executable {
		return tools.ExecutablePermissions(f.repoPerms)
	}
	return f.repoPerms
}

/**
 * Revert non ascii chracters escaped by git or windows (as octal sequences \000) back to bytes.
 */
func DecodePathBytes(path []byte) []byte {
	var expression = regexp.MustCompile(`\\[0-9]{3}`)
	var buffer bytes.Buffer

	// strip quotes if any
	if len(path) > 2 && path[0] == '"' && path[len(path)-1] == '"' {
		path = path[1 : len(path)-1]
	}

	base := 0
	for _, submatches := range expression.FindAllSubmatchIndex(path, -1) {
		buffer.Write(path[base:submatches[0]])

		match := string(path[submatches[0]+1 : submatches[0]+4])

		k, err := strconv.ParseUint(match, 8, 64)
		if err != nil {
			return path
		} // abort on error

		buffer.Write([]byte{byte(k)})
		base = submatches[1]
	}

	buffer.Write(path[base:len(path)])

	return buffer.Bytes()
}

func (f *Filesystem) localObjectDir(oid string) string {
	return filepath.Join(f.LFSObjectDir(), oid[0:2], oid[2:4])
}

func (f *Filesystem) ObjectReferencePaths(oid string) []string {
	if len(f.ReferenceDirs) == 0 {
		return nil
	}

	var paths []string
	for _, ref := range f.ReferenceDirs {
		paths = append(paths, filepath.Join(ref, oid[0:2], oid[2:4], oid))
	}
	return paths
}

func (f *Filesystem) LFSObjectDir() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.lfsobjdir) == 0 {
		f.lfsobjdir = filepath.Join(f.LFSStorageDir, "objects")
		tools.MkdirAll(f.lfsobjdir, f)
	}

	return f.lfsobjdir
}

func (f *Filesystem) LogDir() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.logdir) == 0 {
		f.logdir = filepath.Join(f.LFSStorageDir, "logs")
		tools.MkdirAll(f.logdir, f)
	}

	return f.logdir
}

func (f *Filesystem) TempDir() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.tmpdir) == 0 {
		f.tmpdir = filepath.Join(f.LFSStorageDir, "tmp")
		tools.MkdirAll(f.tmpdir, f)
	}

	return f.tmpdir
}

func (f *Filesystem) Cleanup() error {
	if f == nil {
		return nil
	}
	return f.cleanupTmp()
}

// New initializes a new *Filesystem with the given directories. gitdir is the
// path to the bare repo, workdir is the path to the repository working
// directory, and lfsdir is the optional path to the `.git/lfs` directory.
// repoPerms is the permissions for directories in the repository.
func New(env Environment, gitdir, workdir, lfsdir string, repoPerms os.FileMode) *Filesystem {
	fs := &Filesystem{
		GitStorageDir: resolveGitStorageDir(gitdir),
	}

	fs.ReferenceDirs = resolveReferenceDirs(env, fs.GitStorageDir)

	if len(lfsdir) == 0 {
		lfsdir = "lfs"
	}

	if filepath.IsAbs(lfsdir) {
		fs.LFSStorageDir = lfsdir
	} else {
		fs.LFSStorageDir = filepath.Join(fs.GitStorageDir, lfsdir)
	}

	fs.repoPerms = repoPerms

	return fs
}

func resolveReferenceDirs(env Environment, gitStorageDir string) []string {
	var references []string

	envAlternates, ok := env.Get("GIT_ALTERNATE_OBJECT_DIRECTORIES")
	if ok {
		splits := strings.Split(envAlternates, string(os.PathListSeparator))
		for _, split := range splits {
			if dir, ok := existsAlternate(split); ok {
				references = append(references, dir)
			}
		}
	}

	cloneReferencePath := filepath.Join(gitStorageDir, "objects", "info", "alternates")
	if tools.FileExists(cloneReferencePath) {
		f, err := os.Open(cloneReferencePath)
		if err != nil {
			tracerx.Printf("could not open %s: %s",
				cloneReferencePath, err)
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if len(text) == 0 || strings.HasPrefix(text, "#") {
				continue
			}

			if dir, ok := existsAlternate(text); ok {
				references = append(references, dir)
			}
		}

		if err := scanner.Err(); err != nil {
			tracerx.Printf("could not scan %s: %s",
				cloneReferencePath, err)
		}
	}
	return references
}

// existsAlternate takes an object directory given in "objs" (read as a single,
// line from .git/objects/info/alternates). If that is a satisfiable alternates
// directory (i.e., it exists), the directory is returned along with "true". If
// not, the empty string and false is returned instead.
func existsAlternate(objs string) (string, bool) {
	objs = strings.TrimSpace(objs)
	if strings.HasPrefix(objs, "\"") {
		var err error

		unquote := strings.LastIndex(objs, "\"")
		if unquote == 0 {
			return "", false
		}

		objs, err = strconv.Unquote(objs[:unquote+1])
		if err != nil {
			return "", false
		}
	}

	storage := filepath.Join(filepath.Dir(objs), "lfs", "objects")

	if tools.DirExists(storage) {
		return storage, true
	}
	return "", false
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
