package lfs

import (
	"encoding/hex"
	"sync"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/git"
)

// The nameMap structure provides a goroutine-safe mapping of Git object IDs
// (as SHA hex strings) to their pathspecs, either fully-qualified directory
// paths for trees or file paths for blobs, as returned by "git rev-list".
type nameMap struct {
	names map[string]string
	mutex *sync.Mutex
}

func (m *nameMap) getName(sha string) (string, bool) {
	m.mutex.Lock()
	name, ok := m.names[sha]
	m.mutex.Unlock()
	return name, ok
}

func (m *nameMap) setName(sha, name string) {
	m.mutex.Lock()
	m.names[sha] = name
	m.mutex.Unlock()
}

func newNameMap() *nameMap {
	return &nameMap{
		names: make(map[string]string, 0),
		mutex: &sync.Mutex{},
	}
}

type lockableNameSet struct {
	nameMap *nameMap
	set     GitScannerSet
}

// Determines if the given blob sha matches a locked file.
func (s *lockableNameSet) Check(blobSha string) (string, bool) {
	if s == nil || s.nameMap == nil || s.set == nil {
		return "", false
	}

	name, ok := s.nameMap.getName(blobSha)
	if !ok {
		return name, ok
	}

	if s.set.Contains(name) {
		return name, true
	}
	return name, false
}

func noopFoundLockable(name string) {}

// scanRefsToChan scans through all unique objects reachable from the
// "include" refs and not reachable from any "exclude" refs and invokes the
// provided callback for each pointer file, valid or invalid, that it finds.
// Reports unique OIDs once only, not multiple times if more than one file
// has the same content.
func scanRefsToChan(scanner *GitScanner, pointerCb GitScannerFoundPointer, include, exclude []string, gitEnv, osEnv config.Environment) error {
	revs, nameMap, err := revListShas(scanner, include, exclude)
	if err != nil {
		return err
	}

	lockableSet := &lockableNameSet{nameMap: nameMap, set: scanner.potentialLockables}
	smallShas, batchLockableCh, err := catFileBatchCheck(revs, lockableSet)
	if err != nil {
		return err
	}

	lockableCb := scanner.foundLockable
	if lockableCb == nil {
		lockableCb = noopFoundLockable
	}

	go func(cb GitScannerFoundLockable, ch chan string) {
		for name := range ch {
			cb(name)
		}
	}(lockableCb, batchLockableCh)

	pointers, checkLockableCh, err := catFileBatch(smallShas, lockableSet, gitEnv, osEnv)
	if err != nil {
		return err
	}

	for p := range pointers.Results {
		if name, ok := nameMap.getName(p.Sha1); ok {
			p.Name = name
		}

		if scanner.Filter.Allows(p.Name) {
			pointerCb(p, nil)
		}
	}

	for lockableName := range checkLockableCh {
		if scanner.Filter.Allows(lockableName) {
			lockableCb(lockableName)
		}
	}

	if err := pointers.Wait(); err != nil {
		pointerCb(nil, err)
	}

	return nil
}

// scanRefsToChanSingleIncludeExclude scans through all unique objects
// reachable from the "include" ref and not reachable from the "exclude" ref
// and invokes the provided callback for each pointer file, valid or invalid,
// that it finds.
// Reports unique OIDs once only, not multiple times if more than one file
// has the same content.
func scanRefsToChanSingleIncludeExclude(scanner *GitScanner, pointerCb GitScannerFoundPointer, include, exclude string, gitEnv, osEnv config.Environment) error {
	return scanRefsToChan(scanner, pointerCb, []string{include}, []string{exclude}, gitEnv, osEnv)
}

// scanRefsToChanSingleIncludeMultiExclude scans through all unique objects
// reachable from the "include" ref and not reachable from any "exclude" refs
// and invokes the provided callback for each pointer file, valid or invalid,
// that it finds.
// Reports unique OIDs once only, not multiple times if more than one file
// has the same content.
func scanRefsToChanSingleIncludeMultiExclude(scanner *GitScanner, pointerCb GitScannerFoundPointer, include string, exclude []string, gitEnv, osEnv config.Environment) error {
	return scanRefsToChan(scanner, pointerCb, []string{include}, exclude, gitEnv, osEnv)
}

// scanRefsByTree scans through all objects reachable from the "include" refs
// and not reachable from any "exclude" refs and invokes the provided callback
// for each pointer file, valid or invalid, that it finds.
// Objects which appear in multiple trees will be visited once per tree.
func scanRefsByTree(scanner *GitScanner, pointerCb GitScannerFoundPointer, include, exclude []string, gitEnv, osEnv config.Environment) error {
	revs, _, err := revListShas(scanner, include, exclude)
	if err != nil {
		return err
	}

	errchan := make(chan error, 20) // multiple errors possible
	wg := &sync.WaitGroup{}

	for r := range revs.Results {
		wg.Add(1)
		go func(rev string) {
			defer wg.Done()
			err := runScanTreeForPointers(pointerCb, rev, gitEnv, osEnv)
			if err != nil {
				errchan <- err
			}
		}(r)
	}

	wg.Wait()
	close(errchan)
	for err := range errchan {
		if err != nil {
			return err
		}
	}

	return revs.Wait()
}

// revListShas uses git rev-list to return the list of object sha1s
// for the given ref. If all is true, ref is ignored. It returns a
// channel from which sha1 strings can be read, and a map of the sha1
// object IDs to their pathspecs, which will be populated as the sha1
// strings are written to the channel.
func revListShas(scanner *GitScanner, include, exclude []string) (*StringChannelWrapper, *nameMap, error) {
	nameMap := newNameMap()
	revListScanner, err := git.NewRevListScanner(include, exclude, &git.ScanRefsOptions{
		Mode:             git.ScanningMode(scanner.mode),
		SkipDeletedBlobs: scanner.skipDeletedBlobs,
		CommitsOnly:      scanner.commitsOnly,
		Remote:           scanner.remote,
		SkippedRefs:      scanner.skippedRefs,
		Names:            nameMap.names,
		Mutex:            nameMap.mutex,
	})

	if err != nil {
		return nil, nil, err
	}

	revs := make(chan string, chanBufSize)
	errs := make(chan error, 5) // may be multiple errors

	go func() {
		for revListScanner.Scan() {
			sha := hex.EncodeToString(revListScanner.OID())
			if name := revListScanner.Name(); len(name) > 0 {
				nameMap.setName(sha, name)
			}
			revs <- sha
		}

		if err = revListScanner.Err(); err != nil {
			errs <- err
		}

		if err = revListScanner.Close(); err != nil {
			errs <- err
		}

		close(revs)
		close(errs)
	}()

	return NewStringChannelWrapper(revs, errs), nameMap, nil
}
