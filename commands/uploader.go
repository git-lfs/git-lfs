package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/locking"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tq"
	"github.com/rubyist/tracerx"
)

func uploadLeftOrAll(g *lfs.GitScanner, ctx *uploadContext, ref string) error {
	if pushAll {
		if err := g.ScanRefWithDeleted(ref, nil); err != nil {
			return err
		}
	} else {
		if err := g.ScanLeftToRemote(ref, nil); err != nil {
			return err
		}
	}
	return ctx.scannerError()
}

type uploadContext struct {
	Remote       string
	DryRun       bool
	Manifest     *tq.Manifest
	uploadedOids tools.StringSet

	meter progress.Meter
	tq    *tq.TransferQueue

	committerName  string
	committerEmail string

	trackedLocksMu *sync.Mutex

	// ALL verifiable locks
	ourLocks   map[string]locking.Lock
	theirLocks map[string]locking.Lock

	// locks from ourLocks that were modified in this push
	ownedLocks []locking.Lock

	// locks from theirLocks that were modified in this push
	unownedLocks []locking.Lock

	// tracks errors from gitscanner callbacks
	scannerErr error
	errMu      sync.Mutex
}

// Determines if a filename is lockable. Serves as a wrapper around theirLocks
// that implements GitScannerSet.
type gitScannerLockables struct {
	m map[string]locking.Lock
}

func (l *gitScannerLockables) Contains(name string) bool {
	if l == nil {
		return false
	}
	_, ok := l.m[name]
	return ok
}

type verifyState byte

const (
	verifyStateUnknown verifyState = iota
	verifyStateEnabled
	verifyStateDisabled
)

func newUploadContext(remote string, dryRun bool) *uploadContext {
	cfg.CurrentRemote = remote

	ctx := &uploadContext{
		Remote:         remote,
		Manifest:       getTransferManifest(),
		DryRun:         dryRun,
		uploadedOids:   tools.NewStringSet(),
		ourLocks:       make(map[string]locking.Lock),
		theirLocks:     make(map[string]locking.Lock),
		trackedLocksMu: new(sync.Mutex),
	}

	ctx.meter = buildProgressMeter(ctx.DryRun)
	ctx.tq = newUploadQueue(ctx.Manifest, ctx.Remote, tq.WithProgress(ctx.meter), tq.DryRun(ctx.DryRun))
	ctx.committerName, ctx.committerEmail = cfg.CurrentCommitter()

	ourLocks, theirLocks := verifyLocks(remote)
	for _, l := range theirLocks {
		ctx.theirLocks[l.Path] = l
	}
	for _, l := range ourLocks {
		ctx.ourLocks[l.Path] = l
	}

	return ctx
}

func verifyLocks(remote string) (ours, theirs []locking.Lock) {
	endpoint := getAPIClient().Endpoints.Endpoint("upload", remote)

	state := getVerifyStateFor(endpoint)
	if state == verifyStateDisabled {
		return
	}

	lockClient := newLockClient(remote)

	ours, theirs, err := lockClient.VerifiableLocks(0)
	if err != nil {
		if errors.IsNotImplementedError(err) {
			disableFor(endpoint)
		} else {
			Print("Remote %q does not support the LFS locking API. Consider disabling it with:", remote)
			Print("  $ git config 'lfs.%s.locksverify' false", endpoint.Url)

			if state == verifyStateEnabled {
				ExitWithError(err)
			}
		}
	} else if state == verifyStateUnknown {
		Print("Locking support detected on remote %q. Consider enabling it with:", remote)
		Print("  $ git config 'lfs.%s.locksverify' true", endpoint.Url)
	}

	return ours, theirs
}

func (c *uploadContext) scannerError() error {
	c.errMu.Lock()
	defer c.errMu.Unlock()

	return c.scannerErr
}

func (c *uploadContext) addScannerError(err error) {
	c.errMu.Lock()
	defer c.errMu.Unlock()

	if c.scannerErr != nil {
		c.scannerErr = fmt.Errorf("%v\n%v", c.scannerErr, err)
	} else {
		c.scannerErr = err
	}
}

func (c *uploadContext) buildGitScanner() (*lfs.GitScanner, error) {
	gitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			c.addScannerError(err)
		} else {
			uploadPointers(c, p)
		}
	})

	gitscanner.FoundLockable = func(name string) {
		if lock, ok := c.theirLocks[name]; ok {
			c.trackedLocksMu.Lock()
			c.unownedLocks = append(c.unownedLocks, lock)
			c.trackedLocksMu.Unlock()
		}
	}

	gitscanner.PotentialLockables = &gitScannerLockables{m: c.theirLocks}
	return gitscanner, gitscanner.RemoteForPush(c.Remote)
}

// AddUpload adds the given oid to the set of oids that have been uploaded in
// the current process.
func (c *uploadContext) SetUploaded(oid string) {
	c.uploadedOids.Add(oid)
}

// HasUploaded determines if the given oid has already been uploaded in the
// current process.
func (c *uploadContext) HasUploaded(oid string) bool {
	return c.uploadedOids.Contains(oid)
}

func (c *uploadContext) prepareUpload(unfiltered ...*lfs.WrappedPointer) (*tq.TransferQueue, []*lfs.WrappedPointer) {
	numUnfiltered := len(unfiltered)
	uploadables := make([]*lfs.WrappedPointer, 0, numUnfiltered)

	// XXX(taylor): temporary measure to fix duplicate (broken) results from
	// scanner
	uniqOids := tools.NewStringSet()

	// separate out objects that _should_ be uploaded, but don't exist in
	// .git/lfs/objects. Those will skipped if the server already has them.
	for _, p := range unfiltered {
		// object already uploaded in this process, or we've already
		// seen this OID (see above), skip!
		if uniqOids.Contains(p.Oid) || c.HasUploaded(p.Oid) {
			continue
		}
		uniqOids.Add(p.Oid)

		// canUpload determines whether the current pointer "p" can be
		// uploaded through the TransferQueue below. It is set to false
		// only when the file is locked by someone other than the
		// current committer.
		var canUpload bool = true

		if lock, ok := c.theirLocks[p.Name]; ok {
			c.trackedLocksMu.Lock()
			c.unownedLocks = append(c.unownedLocks, lock)
			c.trackedLocksMu.Unlock()
			canUpload = false
		}

		if lock, ok := c.ourLocks[p.Name]; ok {
			c.trackedLocksMu.Lock()
			c.ownedLocks = append(c.ownedLocks, lock)
			c.trackedLocksMu.Unlock()
		}

		if canUpload {
			// estimate in meter early (even if it's not going into
			// uploadables), since we will call Skip() based on the
			// results of the download check queue.
			c.meter.Add(p.Size)

			uploadables = append(uploadables, p)
		}
	}

	return c.tq, uploadables
}

func uploadPointers(c *uploadContext, unfiltered ...*lfs.WrappedPointer) {
	if c.DryRun {
		for _, p := range unfiltered {
			if c.HasUploaded(p.Oid) {
				continue
			}

			Print("push %s => %s", p.Oid, p.Name)
			c.SetUploaded(p.Oid)
		}

		return
	}

	q, pointers := c.prepareUpload(unfiltered...)
	for _, p := range pointers {
		t, err := uploadTransfer(p)
		if err != nil && !errors.IsCleanPointerError(err) {
			ExitWithError(err)
		}

		q.Add(t.Name, t.Path, t.Oid, t.Size)
		c.SetUploaded(p.Oid)
	}
}

func (c *uploadContext) Await() {
	c.tq.Wait()

	for _, err := range c.tq.Errors() {
		FullError(err)
	}

	if len(c.tq.Errors()) > 0 {
		os.Exit(2)
	}

	var avoidPush bool

	c.trackedLocksMu.Lock()
	if ul := len(c.unownedLocks); ul > 0 {
		avoidPush = true

		Print("Unable to push %d locked file(s):", ul)
		for _, unowned := range c.unownedLocks {
			Print("* %s - %s", unowned.Path, unowned.Owner)
		}
	} else if len(c.ownedLocks) > 0 {
		Print("Consider unlocking your own locked file(s): (`git lfs unlock <path>`)")
		for _, owned := range c.ownedLocks {
			Print("* %s", owned.Path)
		}
	}
	c.trackedLocksMu.Unlock()

	if avoidPush {
		Error("WARNING: The above files would have halted this push.")
	}
}

// getVerifyStateFor returns whether or not lock verification is enabled for the
// given "endpoint". If no state has been explicitly set, an "unknown" state
// will be returned instead.
func getVerifyStateFor(endpoint lfsapi.Endpoint) verifyState {
	key := strings.Join([]string{"lfs", endpoint.Url, "locksverify"}, ".")

	v, ok := cfg.Git.Get(key)
	if !ok {
		return verifyStateUnknown
	}

	if enabled, _ := strconv.ParseBool(v); enabled {
		return verifyStateEnabled
	}
	return verifyStateDisabled
}

// disableFor disables lock verification for the given lfsapi.Endpoint,
// "endpoint".
func disableFor(endpoint lfsapi.Endpoint) error {
	tracerx.Printf("commands: disabling lock verification for %q", endpoint.Url)

	key := strings.Join([]string{"lfs", endpoint.Url, "locksverify"}, ".")

	_, err := git.Config.SetLocal("", key, "false")
	return err
}
