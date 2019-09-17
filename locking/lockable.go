package locking

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/git/gitattr"
	"github.com/git-lfs/git-lfs/tools"
)

// GetLockablePatterns returns a list of patterns in .gitattributes which are
// marked as lockable
func (c *Client) GetLockablePatterns() []string {
	c.ensureLockablesLoaded()
	return c.lockablePatterns
}

// getLockableFilter returns the internal filter used to check if a file is lockable
func (c *Client) getLockableFilter() *filepathfilter.Filter {
	c.ensureLockablesLoaded()
	return c.lockableFilter
}

func (c *Client) ensureLockablesLoaded() {
	c.lockableMutex.Lock()
	defer c.lockableMutex.Unlock()

	// Only load once
	if c.lockablePatterns == nil {
		c.refreshLockablePatterns()
	}
}

// Internal function to repopulate lockable patterns
// You must have locked the c.lockableMutex in the caller
func (c *Client) refreshLockablePatterns() {
	paths := git.GetAttributePaths(gitattr.NewMacroProcessor(), c.LocalWorkingDir, c.LocalGitDir)
	// Always make non-nil even if empty
	c.lockablePatterns = make([]string, 0, len(paths))
	for _, p := range paths {
		if p.Lockable {
			c.lockablePatterns = append(c.lockablePatterns, p.Path)
		}
	}
	c.lockableFilter = filepathfilter.New(c.lockablePatterns, nil)
}

// IsFileLockable returns whether a specific file path is marked as Lockable,
// ie has the 'lockable' attribute in .gitattributes
// Lockable patterns are cached once for performance, unless you call RefreshLockablePatterns
// path should be relative to repository root
func (c *Client) IsFileLockable(path string) bool {
	return c.getLockableFilter().Allows(path)
}

// FixAllLockableFileWriteFlags recursively scans the repo looking for files which
// are lockable, and makes sure their write flags are set correctly based on
// whether they are currently locked or unlocked.
// Files which are unlocked are made read-only, files which are locked are made
// writeable.
// This function can be used after a clone or checkout to ensure that file
// state correctly reflects the locking state
func (c *Client) FixAllLockableFileWriteFlags() error {
	return c.fixFileWriteFlags(c.LocalWorkingDir, c.LocalWorkingDir, c.getLockableFilter(), nil)
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
func (c *Client) FixFileWriteFlagsInDir(dir string, lockablePatterns, unlockablePatterns []string) error {

	// early-out if no patterns
	if len(lockablePatterns) == 0 && len(unlockablePatterns) == 0 {
		return nil
	}

	absPath := dir
	if !filepath.IsAbs(dir) {
		absPath = filepath.Join(c.LocalWorkingDir, dir)
	}
	stat, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("%q is not a valid directory", dir)
	}

	var lockableFilter *filepathfilter.Filter
	var unlockableFilter *filepathfilter.Filter
	if lockablePatterns != nil {
		lockableFilter = filepathfilter.New(lockablePatterns, nil)
	}
	if unlockablePatterns != nil {
		unlockableFilter = filepathfilter.New(unlockablePatterns, nil)
	}

	return c.fixFileWriteFlags(absPath, c.LocalWorkingDir, lockableFilter, unlockableFilter)
}

// Internal implementation of fixing file write flags with precompiled filters
func (c *Client) fixFileWriteFlags(absPath, workingDir string, lockable, unlockable *filepathfilter.Filter) error {

	// Build a list of files
	lsFiles, err := git.NewLsFiles(workingDir, !c.ModifyIgnoredFiles)
	if err != nil {
		return err
	}

	for f := range lsFiles.Files {
		err = c.fixSingleFileWriteFlags(f, lockable, unlockable)
		if err != nil {
			return err
		}
	}

	return nil
}

// FixLockableFileWriteFlags checks each file in the provided list, and for
// those which are lockable, makes sure their write flags are set correctly
// based on whether they are currently locked or unlocked. Files which are
// unlocked are made read-only, files which are locked are made writeable.
// Files which are not lockable are ignored.
// This function can be used after a clone or checkout to ensure that file
// state correctly reflects the locking state, and is more efficient than
// FixAllLockableFileWriteFlags when you know which files changed
func (c *Client) FixLockableFileWriteFlags(files []string) error {
	// early-out if no lockable patterns
	if len(c.GetLockablePatterns()) == 0 {
		return nil
	}

	var errs []error
	for _, f := range files {
		err := c.fixSingleFileWriteFlags(f, c.getLockableFilter(), nil)
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
func (c *Client) fixSingleFileWriteFlags(file string, lockable, unlockable *filepathfilter.Filter) error {
	// Convert to git-style forward slash separators if necessary
	// Necessary to match attributes
	if filepath.Separator == '\\' {
		file = strings.Replace(file, "\\", "/", -1)
	}
	if lockable != nil && lockable.Allows(file) {
		// Lockable files are writeable only if they're currently locked
		err := tools.SetFileWriteFlag(file, c.IsFileLockedByCurrentCommitter(file))
		// Ignore not exist errors
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	} else if unlockable != nil && unlockable.Allows(file) {
		// Unlockable files are always writeable
		// We only check files which match the incoming patterns to avoid
		// checking every file in the system all the time, and only do it
		// when a file has had its lockable attribute removed
		err := tools.SetFileWriteFlag(file, true)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
