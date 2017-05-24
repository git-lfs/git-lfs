package lfs

import (
	"encoding/hex"
	"regexp"

	"github.com/git-lfs/git-lfs/git"
)

var z40 = regexp.MustCompile(`\^?0{40}`)

type lockableNameSet struct {
	opt *ScanRefsOptions
	set GitScannerSet
}

// Determines if the given blob sha matches a locked file.
func (s *lockableNameSet) Check(blobSha string) (string, bool) {
	if s == nil || s.opt == nil || s.set == nil {
		return "", false
	}

	name, ok := s.opt.GetName(blobSha)
	if !ok {
		return name, ok
	}

	if s.set.Contains(name) {
		return name, true
	}
	return name, false
}

func noopFoundLockable(name string) {}

// scanRefsToChan takes a ref and returns a channel of WrappedPointer objects
// for all Git LFS pointers it finds for that ref.
// Reports unique oids once only, not multiple times if >1 file uses the same content
func scanRefsToChan(scanner *GitScanner, pointerCb GitScannerFoundPointer, refLeft, refRight string, opt *ScanRefsOptions) error {
	if opt == nil {
		panic("no scan ref options")
	}

	revs, err := revListShas(refLeft, refRight, opt)
	if err != nil {
		return err
	}

	lockableSet := &lockableNameSet{opt: opt, set: scanner.PotentialLockables}
	smallShas, batchLockableCh, err := catFileBatchCheck(revs, lockableSet)
	if err != nil {
		return err
	}

	lockableCb := scanner.FoundLockable
	if lockableCb == nil {
		lockableCb = noopFoundLockable
	}

	go func(cb GitScannerFoundLockable, ch chan string) {
		for name := range ch {
			cb(name)
		}
	}(lockableCb, batchLockableCh)

	pointers, checkLockableCh, err := catFileBatch(smallShas, lockableSet)
	if err != nil {
		return err
	}

	for p := range pointers.Results {
		if name, ok := opt.GetName(p.Sha1); ok {
			p.Name = name
		}
		pointerCb(p, nil)
	}

	for lockableName := range checkLockableCh {
		lockableCb(lockableName)
	}

	if err := pointers.Wait(); err != nil {
		pointerCb(nil, err)
	}

	return nil
}

// revListShas uses git rev-list to return the list of object sha1s
// for the given ref. If all is true, ref is ignored. It returns a
// channel from which sha1 strings can be read.
func revListShas(refLeft, refRight string, opt *ScanRefsOptions) (*StringChannelWrapper, error) {
	scanner, err := git.NewRevListScanner(refLeft, refRight, &git.ScanRefsOptions{
		Mode:             git.ScanningMode(opt.ScanMode),
		Remote:           opt.RemoteName,
		SkipDeletedBlobs: opt.SkipDeletedBlobs,
		SkippedRefs:      opt.skippedRefs,
		Mutex:            opt.mutex,
		Names:            opt.nameMap,
	})

	if err != nil {
		return nil, err
	}

	revs := make(chan string, chanBufSize)
	errs := make(chan error, 5) // may be multiple errors

	go func() {
		for scanner.Scan() {
			sha := hex.EncodeToString(scanner.OID())
			if name := scanner.Name(); len(name) > 0 {
				opt.SetName(sha, name)
			}
			revs <- sha
		}

		if err = scanner.Err(); err != nil {
			errs <- err
		}

		if err = scanner.Close(); err != nil {
			errs <- err
		}

		close(revs)
		close(errs)
	}()

	return NewStringChannelWrapper(revs, errs), nil
}

// Get additional arguments needed to limit 'git rev-list' to just the changes
// in refTo that are also not on remoteName.
//
// Returns a slice of string command arguments, and a slice of string git
// commits to pass to `git rev-list` via STDIN.
func revListArgsRefVsRemote(refTo, remoteName string, skippedRefs []string) ([]string, []string) {
	if len(skippedRefs) < 1 {
		// Safe to use cached
		return []string{refTo, "--not", "--remotes=" + remoteName}, nil
	}

	// Use only the non-missing refs as 'from' points
	commits := make([]string, 1, len(skippedRefs)+1)
	commits[0] = refTo
	return []string{"--stdin"}, append(commits, skippedRefs...)
}
