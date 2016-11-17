package lfs

import (
	"fmt"
	"time"
)

// GitScanner scans objects in a Git repository for LFS pointers.
type GitScanner struct {
	remote      string
	skippedRefs []string
}

// NewGitScanner initializes a *GitScanner for a Git repository in the current
// working directory.
func NewGitScanner() *GitScanner {
	return &GitScanner{}
}

// RemoteForPush sets up this *GitScanner to scan for objects to push to the
// given remote. Needed for ScanLeftToRemote().
func (s *GitScanner) RemoteForPush(r string) {
	s.remote = r
	s.skippedRefs = calcSkippedRefs(r)
}

// ScanLeftToRemote scans through all commits starting at the given ref that the
// given remote does not have. See RemoteForPush().
func (s *GitScanner) ScanLeftToRemote(left string) (*PointerChannelWrapper, error) {
	if len(s.remote) == 0 {
		return nil, fmt.Errorf("Unable to scan starting at %q: no remote set.", left)
	}
	return scanRefsToChan(left, "", s.opts(ScanLeftToRemoteMode))
}

// ScanRefRange scans through all commits from the given left and right refs,
// including git objects that have been modified or deleted.
func (s *GitScanner) ScanRefRange(left, right string) (*PointerChannelWrapper, error) {
	opts := s.opts(ScanRefsMode)
	opts.SkipDeletedBlobs = false
	return scanRefsToChan(left, right, opts)
}

// ScanRefWithDeleted scans through all objects in the given ref, including
// git objects that have been modified or deleted.
func (s *GitScanner) ScanRefWithDeleted(ref string) (*PointerChannelWrapper, error) {
	return s.ScanRefRange(ref, "")
}

// ScanRef scans through all objects in the current ref, excluding git objects
// that have been modified or deleted before the ref.
func (s *GitScanner) ScanRef(ref string) (*PointerChannelWrapper, error) {
	opts := s.opts(ScanRefsMode)
	opts.SkipDeletedBlobs = true
	return scanRefsToChan(ref, "", opts)
}

// ScanAll scans through all objects in the git repository.
func (s *GitScanner) ScanAll() (*PointerChannelWrapper, error) {
	opts := s.opts(ScanAllMode)
	opts.SkipDeletedBlobs = false
	return scanRefsToChan("", "", opts)
}

// ScanUnpushed scans history for all LFS pointers which have been added but not
// pushed to the named remote. remote can be left blank to mean 'any remote'.
func (s *GitScanner) ScanUnpushed(remote string) (*PointerChannelWrapper, error) {
	return scanUnpushed(remote)
}

// ScanPreviousVersions scans changes reachable from ref (commit) back to since.
// Returns channel of pointers for *previous* versions that overlap that time.
// Does not include pointers which were still in use at ref (use ScanRefsToChan
// for that)
func (s *GitScanner) ScanPreviousVersions(ref string, since time.Time) (*PointerChannelWrapper, error) {
	return logPreviousSHAs(ref, since)
}

func (s *GitScanner) opts(mode ScanningMode) *ScanRefsOptions {
	opts := newScanRefsOptions()
	opts.ScanMode = mode
	opts.RemoteName = s.remote
	opts.skippedRefs = s.skippedRefs
	return opts
}
