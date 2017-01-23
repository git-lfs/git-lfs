package commands

import (
	"os"
	"sync"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/locking"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tq"
)

type uploadContext struct {
	Remote       string
	DryRun       bool
	Manifest     *tq.Manifest
	uploadedOids tools.StringSet

	meter progress.Meter
	tq    *tq.TransferQueue

	committerName  string
	committerEmail string

	locks          map[string]locking.Lock
	trackedLocksMu *sync.Mutex
	ownedLocks     []locking.Lock
	unownedLocks   []locking.Lock
}

func newUploadContext(remote string, dryRun bool) *uploadContext {
	cfg.CurrentRemote = remote

	ctx := &uploadContext{
		Remote:         remote,
		Manifest:       getTransferManifest(),
		DryRun:         dryRun,
		uploadedOids:   tools.NewStringSet(),
		locks:          make(map[string]locking.Lock),
		trackedLocksMu: new(sync.Mutex),
	}

	ctx.meter = buildProgressMeter(ctx.DryRun)
	ctx.tq = newUploadQueue(ctx.Manifest, ctx.Remote, tq.WithProgress(ctx.meter), tq.DryRun(ctx.DryRun))
	ctx.committerName, ctx.committerEmail = cfg.CurrentCommitter()

	lockClient := newLockClient(remote)
	locks, err := lockClient.SearchLocks(nil, 0, false)
	if err != nil {
		Error("WARNING: Unable to search for locks contained in this push.")
		Error("         Temporarily skipping check ...")
	} else {
		for _, l := range locks {
			ctx.locks[l.Path] = l
		}
	}

	return ctx
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

		if lock, ok := c.locks[p.Name]; ok {
			owned := lock.Committer.Name == c.committerName &&
				lock.Committer.Email == c.committerEmail

			c.trackedLocksMu.Lock()
			if owned {
				c.ownedLocks = append(c.ownedLocks, lock)
			} else {
				c.unownedLocks = append(c.unownedLocks, lock)
				canUpload = false
			}
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
			Print("* %s - %s", unowned.Path, unowned.Committer)
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
