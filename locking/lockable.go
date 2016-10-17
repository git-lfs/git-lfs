package locking

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/github/git-lfs/errors"
	"github.com/github/git-lfs/tools"

	"github.com/github/git-lfs/config"
)

var (
	// lockable patterns from .gitattributes
	cachedLockablePatterns []string
	cachedLockableMutex    sync.Mutex
)

// GetLockablePatterns returns a list of patterns in .gitattributes which are
// marked as lockable
func GetLockablePatterns() []string {
	cachedLockableMutex.Lock()
	defer cachedLockableMutex.Unlock()

	// Only load once
	if cachedLockablePatterns == nil {
		// Always make non-nil even if empty
		cachedLockablePatterns = make([]string, 0, 10)

		paths := config.GetAttributePaths()
		for _, p := range paths {
			if p.Lockable {
				cachedLockablePatterns = append(cachedLockablePatterns, p.Path)
			}
		}
	}

	return cachedLockablePatterns

}

// RefreshLockablePatterns causes us to re-read the .gitattributes and caches the result
func RefreshLockablePatterns() {
	cachedLockableMutex.Lock()
	defer cachedLockableMutex.Unlock()
	cachedLockablePatterns = nil
}

// IsFileLockable returns whether a specific file path is marked as Lockable,
// ie has the 'lockable' attribute in .gitattributes
// Lockable patterns are cached once for performance, unless you call RefreshLockablePatterns
// path should be relative to repository root
func IsFileLockable(path string) bool {
	return tools.PathMatchesWildcardPatterns(path, GetLockablePatterns())
}

// FixAllLockableFileWriteFlags recursively scans the repo looking for files which
// are lockable, and makes sure their write flags are set correctly based on
// whether they are currently locked or unlocked.
// Files which are unlocked are made read-only, files which are locked are made
// writeable.
// This function can be used after a clone or checkout to ensure that file
// state correctly reflects the locking state
func FixAllLockableFileWriteFlags() error {
	return FixFileWriteFlagsInDir("", GetLockablePatterns(), nil, true)
}

// FixFileWriteFlagsInDir scans dir (which can either be a relative dir
// from the root of the repo, or an absolute dir within the repo) looking for
// files to change permissions for.
// If lockablePatterns is non-nil, then any file matching those patterns will be
// checked to see if it is currently locked by the current committer, and if so
// it will be writeable, and if not locked it will be read-only.
// If unlockablePatterns is non-nil, then any file matching those patterns will
// be made writeable if it is not already. This can be used to reset files to
// writeable when their 'lockable' attribute is turned off.
func FixFileWriteFlagsInDir(dir string, lockablePatterns, unlockablePatterns []string, recursive bool) error {

	// early-out if no patterns
	if len(lockablePatterns) == 0 && len(unlockablePatterns) == 0 {
		return nil
	}

	absPath := dir
	if !filepath.IsAbs(dir) {
		absPath = filepath.Join(config.LocalWorkingDir, dir)
	}
	stat, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("%q is not a valid directory", dir)
	}

	// For simplicity, don't use goroutines to parallelise recursive scan
	// This routine is almost certainly disk-limited anyway
	// We don't need sorting so don't use ioutil.Readdir or filepath.Walk
	d, err := os.Open(absPath)
	if err != nil {
		return err
	}

	contents, err := d.Readdir(-1)
	if err != nil {
		return err
	}
	var errs []error
	for _, fi := range contents {
		abschild := filepath.Join(absPath, fi.Name())
		if fi.IsDir() {
			if recursive {
				err = FixFileWriteFlagsInDir(abschild, lockablePatterns, unlockablePatterns, recursive)
			}
			continue
		}

		// This is a file, get relative to repo root
		relpath, err := filepath.Rel(config.LocalWorkingDir, abschild)
		if err != nil {
			return err
		}

		err = fixSingleFileWriteFlags(relpath, lockablePatterns, unlockablePatterns)
		if err != nil {
			errs = append(errs, err)
		}

	}
	return errors.Combine(errs)
}

// FixLockableFileWriteFlags checks each file in the provided list, and for
// those which are lockable, makes sure their write flags are set correctly
// based on whether they are currently locked or unlocked. Files which are
// unlocked are made read-only, files which are locked are made writeable.
// Files which are not lockable are ignored.
// This function can be used after a clone or checkout to ensure that file
// state correctly reflects the locking state, and is more efficient than
// FixAllLockableFileWriteFlags when you know which files changed
func FixLockableFileWriteFlags(files []string) error {
	lockablePatterns := GetLockablePatterns()

	// early-out if no lockable patterns
	if len(lockablePatterns) == 0 {
		return nil
	}

	var errs []error
	for _, f := range files {
		err := fixSingleFileWriteFlags(f, lockablePatterns, nil)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Combine(errs)
}

// fixSingleFileWriteFlags fixes write flags on a single file
// If lockablePatterns is non-nil, then any file matching those patterns will be
// checked to see if it is currently locked by the current committer, and if so
// it will be writeable, and if not locked it will be read-only.
// If unlockablePatterns is non-nil, then any file matching those patterns will
// be made writeable if it is not already. This can be used to reset files to
// writeable when their 'lockable' attribute is turned off.
func fixSingleFileWriteFlags(file string, lockablePatterns, unlockablePatterns []string) error {
	// Convert to git-style forward slash separators if necessary
	// Necessary to match attributes
	if filepath.Separator == '\\' {
		file = strings.Replace(file, "\\", "/", -1)
	}
	if tools.PathMatchesWildcardPatterns(file, lockablePatterns) {
		// Lockable files are writeable only if they're currently locked
		err := tools.SetFileWriteFlag(file, IsFileLockedByCurrentCommitter(file))
		if err != nil {
			return err
		}
	} else if tools.PathMatchesWildcardPatterns(file, unlockablePatterns) {
		// Unlockable files are always writeable
		// We only check files which match the incoming patterns to avoid
		// checking every file in the system all the time, and only do it
		// when a file has had its lockable attribute removed
		err := tools.SetFileWriteFlag(file, true)
		if err != nil {
			return err
		}
	}
	return nil
}
