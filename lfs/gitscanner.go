package lfs

import (
	"errors"
	"time"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/filepathfilter"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

var missingCallbackErr = errors.New(tr.Tr.Get("no callback given"))

// IsCallbackMissing returns a boolean indicating whether the error is reporting
// that a GitScanner is missing a required GitScannerCallback.
func IsCallbackMissing(err error) bool {
	return err == missingCallbackErr
}

// GitScanner scans objects in a Git repository for LFS pointers.
type GitScanner struct {
	Filter *filepathfilter.Filter

	cfg              *config.Configuration
	mode             ScanningMode
	skipDeletedBlobs bool
	commitsOnly      bool
	foundPointer     GitScannerFoundPointer

	// only set by NewGitScannerForPush()
	remote             string
	skippedRefs        []string
	foundLockable      GitScannerFoundLockable
	potentialLockables GitScannerSet
}

type GitScannerFoundPointer func(*WrappedPointer, error)
type GitScannerFoundLockable func(filename string)

type GitScannerSet interface {
	Contains(string) bool
}

type ScanningMode int

const (
	ScanRefsMode          = ScanningMode(iota) // 0 - or default scan mode
	ScanAllMode           = ScanningMode(iota)
	ScanRangeToRemoteMode = ScanningMode(iota)
)

// NewGitScanner initializes a *GitScanner for a Git repository in the current
// working directory.
func NewGitScanner(cfg *config.Configuration, cb GitScannerFoundPointer) *GitScanner {
	return &GitScanner{cfg: cfg, foundPointer: cb}
}

// NewGitScannerForPush initializes a *GitScanner for a Git repository
// in the current working directory, to scan for objects to push to the
// given remote and for locks on non-LFS objects held by other users.
// Needed for ScanMultiRangeToRemote(), and for ScanRefWithDeleted() when
// used for a "git lfs push --all" command.
func NewGitScannerForPush(cfg *config.Configuration, remote string, cb GitScannerFoundLockable, potentialLockables GitScannerSet) *GitScanner {
	return &GitScanner{
		cfg:                cfg,
		remote:             remote,
		skippedRefs:        calcSkippedRefs(remote),
		foundLockable:      cb,
		potentialLockables: potentialLockables,
	}
}

// ScanMultiRangeToRemote scans through all unique objects reachable from the
// "include" ref but not reachable from any "exclude" refs and which the
// given remote does not have. See NewGitScannerForPush().
func (s *GitScanner) ScanMultiRangeToRemote(include string, exclude []string, cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	if len(s.remote) == 0 {
		return errors.New(tr.Tr.Get("unable to scan starting at %q: no remote set", include))
	}

	s.mode = ScanRangeToRemoteMode

	start := time.Now()
	err = scanRefsToChanSingleIncludeMultiExclude(s, callback, include, exclude, s.cfg.GitEnv(), s.cfg.OSEnv())
	tracerx.PerformanceSince("ScanMultiRangeToRemote", start)

	return err
}

// ScanRefs scans through all unique objects reachable from the "include" refs
// but not reachable from any "exclude" refs, including objects that have
// been modified or deleted.
func (s *GitScanner) ScanRefs(include, exclude []string, cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	start := time.Now()
	err = scanRefsToChan(s, callback, include, exclude, s.cfg.GitEnv(), s.cfg.OSEnv())
	tracerx.PerformanceSince("ScanRefs", start)

	return err
}

// ScanRefRange scans through all unique objects reachable from the "include"
// ref but not reachable from the "exclude" ref, including objects that have
// been modified or deleted.
func (s *GitScanner) ScanRefRange(include, exclude string, cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	start := time.Now()
	err = scanRefsToChanSingleIncludeExclude(s, callback, include, exclude, s.cfg.GitEnv(), s.cfg.OSEnv())
	tracerx.PerformanceSince("ScanRefRange", start)

	return err
}

// ScanRefRangeByTree scans through all objects reachable from the "include"
// ref but not reachable from the "exclude" ref, including objects that have
// been modified or deleted.  Objects which appear in multiple trees will
// be visited once per tree.
func (s *GitScanner) ScanRefRangeByTree(include, exclude string, cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	s.commitsOnly = true

	start := time.Now()
	err = scanRefsByTree(s, callback, []string{include}, []string{exclude}, s.cfg.GitEnv(), s.cfg.OSEnv())
	tracerx.PerformanceSince("ScanRefRangeByTree", start)

	return err
}

// ScanRefWithDeleted scans through all unique objects in the given ref,
// including objects that have been modified or deleted.
func (s *GitScanner) ScanRefWithDeleted(ref string, cb GitScannerFoundPointer) error {
	return s.ScanRefRange(ref, "", cb)
}

// ScanRef scans through all unique objects in the current ref, excluding
// objects that have been modified or deleted before the ref.
func (s *GitScanner) ScanRef(ref string, cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	s.skipDeletedBlobs = true

	start := time.Now()
	err = scanRefsToChanSingleIncludeExclude(s, callback, ref, "", s.cfg.GitEnv(), s.cfg.OSEnv())
	tracerx.PerformanceSince("ScanRef", start)

	return err
}

// ScanRefByTree scans through all objects in the current ref, excluding
// objects that have been modified or deleted before the ref.  Objects which
// appear in multiple trees will be visited once per tree.
func (s *GitScanner) ScanRefByTree(ref string, cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	s.skipDeletedBlobs = true
	s.commitsOnly = true

	start := time.Now()
	err = scanRefsByTree(s, callback, []string{ref}, []string{}, s.cfg.GitEnv(), s.cfg.OSEnv())
	tracerx.PerformanceSince("ScanRefByTree", start)

	return err
}

// ScanAll scans through all unique objects in the repository, including
// objects that have been modified or deleted.
func (s *GitScanner) ScanAll(cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	s.mode = ScanAllMode

	start := time.Now()
	err = scanRefsToChanSingleIncludeExclude(s, callback, "", "", s.cfg.GitEnv(), s.cfg.OSEnv())
	tracerx.PerformanceSince("ScanAll", start)

	return err
}

// ScanTree takes a ref and returns WrappedPointer objects in the tree at that
// ref. Differs from ScanRefs in that multiple files in the tree with the same
// content are all reported.
func (s *GitScanner) ScanTree(ref string, cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	start := time.Now()
	err = runScanTree(callback, ref, s.Filter, s.cfg.GitEnv(), s.cfg.OSEnv())
	tracerx.PerformanceSince("ScanTree", start)

	return err
}

// ScanUnpushed scans history for all LFS pointers which have been added but not
// pushed to the named remote. remote can be left blank to mean 'any remote'.
func (s *GitScanner) ScanUnpushed(remote string, cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	start := time.Now()
	err = scanUnpushed(callback, remote)
	tracerx.PerformanceSince("ScanUnpushed", start)

	return err
}

// ScanStashed scans for all LFS pointers referenced solely by a stash
func (s *GitScanner) ScanStashed(cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	start := time.Now()
	err = scanStashed(callback)
	tracerx.PerformanceSince("ScanStashed", start)

	return err
}

// ScanPreviousVersions scans changes reachable from ref (commit) back to since.
// Returns channel of pointers for *previous* versions that overlap that time.
// Does not include pointers which were still in use at ref (use ScanRefsToChan
// for that)
func (s *GitScanner) ScanPreviousVersions(ref string, since time.Time, cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	start := time.Now()
	err = logPreviousSHAs(callback, ref, s.Filter, since)
	tracerx.PerformanceSince("ScanPreviousVersions", start)

	return err
}

// ScanIndex scans the git index for modified LFS objects.
func (s *GitScanner) ScanIndex(ref string, workingDir string, cb GitScannerFoundPointer) error {
	callback, err := firstGitScannerCallback(cb, s.foundPointer)
	if err != nil {
		return err
	}

	start := time.Now()
	err = scanIndex(callback, ref, workingDir, s.Filter, s.cfg.GitEnv(), s.cfg.OSEnv())
	tracerx.PerformanceSince("ScanIndex", start)

	return err
}

func firstGitScannerCallback(callbacks ...GitScannerFoundPointer) (GitScannerFoundPointer, error) {
	for _, cb := range callbacks {
		if cb == nil {
			continue
		}
		return cb, nil
	}

	return nil, missingCallbackErr
}
