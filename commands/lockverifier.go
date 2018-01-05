package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/locking"
	"github.com/git-lfs/git-lfs/tq"
)

type verifyState byte

const (
	verifyStateUnknown verifyState = iota
	verifyStateEnabled
	verifyStateDisabled
)

func verifyLocksForUpdates(lv *lockVerifier, updates []*git.RefUpdate) {
	for _, update := range updates {
		lv.Verify(update.Right())
	}
}

// lockVerifier verifies locked files before updating one or more refs.
type lockVerifier struct {
	endpoint     lfsapi.Endpoint
	verifyState  verifyState
	verifiedRefs map[string]bool

	// all existing locks
	ourLocks   map[string]*refLock
	theirLocks map[string]*refLock

	// locks from ourLocks that have been modified
	ownedLocks []*refLock

	// locks from theirLocks that have been modified
	unownedLocks []*refLock
}

func (lv *lockVerifier) Verify(ref *git.Ref) {
	if lv.verifyState == verifyStateDisabled || lv.verifiedRefs[ref.Refspec()] {
		return
	}

	lockClient := newLockClient()
	ours, theirs, err := lockClient.VerifiableLocks(ref, 0)
	if err != nil {
		if errors.IsNotImplementedError(err) {
			disableFor(lv.endpoint.Url)
		} else if lv.verifyState == verifyStateUnknown || lv.verifyState == verifyStateEnabled {
			if errors.IsAuthError(err) {
				if lv.verifyState == verifyStateUnknown {
					Error("WARNING: Authentication error: %s", err)
				} else if lv.verifyState == verifyStateEnabled {
					Exit("ERROR: Authentication error: %s", err)
				}
			} else {
				Print("Remote %q does not support the LFS locking API. Consider disabling it with:", cfg.PushRemote())
				Print("  $ git config lfs.%s.locksverify false", lv.endpoint.Url)
				if lv.verifyState == verifyStateEnabled {
					ExitWithError(err)
				}
			}
		}
	} else if lv.verifyState == verifyStateUnknown {
		Print("Locking support detected on remote %q. Consider enabling it with:", cfg.PushRemote())
		Print("  $ git config lfs.%s.locksverify true", lv.endpoint.Url)
	}

	lv.addLocks(ref, ours, lv.ourLocks)
	lv.addLocks(ref, theirs, lv.theirLocks)
	lv.verifiedRefs[ref.Refspec()] = true
}

func (lv *lockVerifier) addLocks(ref *git.Ref, locks []locking.Lock, set map[string]*refLock) {
	for _, l := range locks {
		if rl, ok := set[l.Path]; ok {
			if err := rl.Add(ref, l); err != nil {
				Error("WARNING: error adding %q lock for ref %q: %+v", l.Path, ref, err)
			}
		} else {
			set[l.Path] = lv.newRefLocks(ref, l)
		}
	}
}

// Determines if a filename is lockable. Implements lfs.GitScannerSet
func (lv *lockVerifier) Contains(name string) bool {
	if lv == nil {
		return false
	}
	_, ok := lv.theirLocks[name]
	return ok
}

func (lv *lockVerifier) LockedByThem(name string) bool {
	if lock, ok := lv.theirLocks[name]; ok {
		lv.unownedLocks = append(lv.unownedLocks, lock)
		return true
	}
	return false
}

func (lv *lockVerifier) LockedByUs(name string) bool {
	if lock, ok := lv.ourLocks[name]; ok {
		lv.ownedLocks = append(lv.ownedLocks, lock)
		return true
	}
	return false
}

func (lv *lockVerifier) UnownedLocks() []*refLock {
	return lv.unownedLocks
}

func (lv *lockVerifier) HasUnownedLocks() bool {
	return len(lv.unownedLocks) > 0
}

func (lv *lockVerifier) OwnedLocks() []*refLock {
	return lv.ownedLocks
}

func (lv *lockVerifier) HasOwnedLocks() bool {
	return len(lv.ownedLocks) > 0
}

func (lv *lockVerifier) Enabled() bool {
	return lv.verifyState == verifyStateEnabled
}

func (lv *lockVerifier) newRefLocks(ref *git.Ref, l locking.Lock) *refLock {
	return &refLock{
		allRefs: lv.verifiedRefs,
		path:    l.Path,
		refs:    map[*git.Ref]locking.Lock{ref: l},
	}
}

func newLockVerifier(m *tq.Manifest) *lockVerifier {
	lv := &lockVerifier{
		endpoint:     getAPIClient().Endpoints.Endpoint("upload", cfg.PushRemote()),
		verifiedRefs: make(map[string]bool),
		ourLocks:     make(map[string]*refLock),
		theirLocks:   make(map[string]*refLock),
	}

	// Do not check locks for standalone transfer, because there is no LFS
	// server to ask.
	if m.IsStandaloneTransfer() {
		lv.verifyState = verifyStateDisabled
	} else {
		lv.verifyState = getVerifyStateFor(lv.endpoint.Url)
	}

	return lv
}

// refLock represents a unique locked file path, potentially across multiple
// refs. It tracks each individual lock in case different users locked the
// same path across multiple refs.
type refLock struct {
	path    string
	allRefs map[string]bool
	refs    map[*git.Ref]locking.Lock
}

// Path returns the locked path.
func (r *refLock) Path() string {
	return r.path
}

// Owners returns the list of owners that locked this file, including what
// specific refs the files were locked in. If a user locked a file on all refs,
// don't bother listing them.
//
// Example: technoweenie, bob (refs: foo)
func (r *refLock) Owners() string {
	users := make(map[string][]string, len(r.refs))
	for ref, lock := range r.refs {
		u := lock.Owner.Name
		if _, ok := users[u]; !ok {
			users[u] = make([]string, 0, len(r.refs))
		}
		users[u] = append(users[u], ref.Name)
	}

	owners := make([]string, 0, len(users))
	for name, refs := range users {
		seenRefCount := 0
		for _, ref := range refs {
			if r.allRefs[ref] {
				seenRefCount++
			}
		}
		if seenRefCount == len(r.allRefs) { // lock is included in all refs, so don't list them
			owners = append(owners, name)
			continue
		}

		sort.Strings(refs)
		owners = append(owners, fmt.Sprintf("%s (refs: %s)", name, strings.Join(refs, ", ")))
	}
	sort.Strings(owners)
	return strings.Join(owners, ", ")
}

func (r *refLock) Add(ref *git.Ref, l locking.Lock) error {
	r.refs[ref] = l
	return nil
}

// getVerifyStateFor returns whether or not lock verification is enabled for the
// given url. If no state has been explicitly set, an "unknown" state will be
// returned instead.
func getVerifyStateFor(rawurl string) verifyState {
	uc := config.NewURLConfig(cfg.Git)

	v, ok := uc.Get("lfs", rawurl, "locksverify")
	if !ok {
		if supportsLockingAPI(rawurl) {
			return verifyStateEnabled
		}
		return verifyStateUnknown
	}

	if enabled, _ := strconv.ParseBool(v); enabled {
		return verifyStateEnabled
	}
	return verifyStateDisabled
}
