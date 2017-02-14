package commands

import (
	"os"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tools"
)

var uploadMissingErr = "%s does not exist in .git/lfs/objects. Tried %s, which matches %s."

type uploadContext struct {
	DryRun       bool
	uploadedOids tools.StringSet
<<<<<<< HEAD
}

func newUploadContext(dryRun bool) *uploadContext {
	return &uploadContext{
		DryRun:       dryRun,
		uploadedOids: tools.NewStringSet(),
	}
=======

	meter progress.Meter
	tq    *tq.TransferQueue

	committerName  string
	committerEmail string

	trackedLocksMu *sync.Mutex

	// ALL verifiable locks
	ourLocks   map[string]*locking.Lock
	theirLocks map[string]*locking.Lock

	// locks from ourLocks that were modified in this push
	ownedLocks []*locking.Lock

	// locks from theirLocks that were modified in this push
	unownedLocks []*locking.Lock
}

func newUploadContext(remote string, dryRun bool) *uploadContext {
	cfg.CurrentRemote = remote

	ctx := &uploadContext{
		Remote:         remote,
		Manifest:       getTransferManifest(),
		DryRun:         dryRun,
		uploadedOids:   tools.NewStringSet(),
		ourLocks:       make(map[string]*locking.Lock),
		theirLocks:     make(map[string]*locking.Lock),
		trackedLocksMu: new(sync.Mutex),
	}

	ctx.meter = buildProgressMeter(ctx.DryRun)
	ctx.tq = newUploadQueue(ctx.Manifest, ctx.Remote, tq.WithProgress(ctx.meter), tq.DryRun(ctx.DryRun))
	ctx.committerName, ctx.committerEmail = cfg.CurrentCommitter()

	lockClient := newLockClient(remote)
	ourLocks, theirLocks, err := lockClient.VerifiableLocks(0)
	if err != nil {
		Error("WARNING: Unable to search for locks contained in this push.")
		Error("         Temporarily skipping check ...")
	} else {
		for _, l := range theirLocks {
			ctx.theirLocks[l.Path] = &l
		}
		for _, l := range ourLocks {
			ctx.ourLocks[l.Path] = &l
		}
	}

	return ctx
>>>>>>> f8a50160... Merge branch 'master' into no-dwarf-tables
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

func (c *uploadContext) prepareUpload(unfiltered []*lfs.WrappedPointer) (*lfs.TransferQueue, []*lfs.WrappedPointer) {
	numUnfiltered := len(unfiltered)
	uploadables := make([]*lfs.WrappedPointer, 0, numUnfiltered)
	missingLocalObjects := make([]*lfs.WrappedPointer, 0, numUnfiltered)
	numObjects := 0
	totalSize := int64(0)
	missingSize := int64(0)

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

		// increment numObjects and totalSize early (even if it's not
		// going into uploadables), since we will call Skip() based on
		// the results of the download check queue
		numObjects += 1
		totalSize += p.Size

		if lfs.ObjectExistsOfSize(p.Oid, p.Size) {
			uploadables = append(uploadables, p)
		} else {
			// We think we need to push this but we don't have it
			// Store for server checking later
			missingLocalObjects = append(missingLocalObjects, p)
			missingSize += p.Size
		}
	}

	// check to see if the server has the missing objects.
	c.checkMissing(missingLocalObjects, missingSize)

	// build the TransferQueue, automatically skipping any missing objects that
	// the server already has.
	uploadQueue := lfs.NewUploadQueue(numObjects, totalSize, c.DryRun)
	for _, p := range missingLocalObjects {
		if c.HasUploaded(p.Oid) {
			// if the server already has this object, call Skip() on
			// the progressmeter to decrement the number of files by
			// 1 and the number of bytes by `p.Size`.
			uploadQueue.Skip(p.Size)
		} else {
			uploadables = append(uploadables, p)
		}
	}

	return uploadQueue, uploadables
}

// This checks the given slice of pointers that don't exist in .git/lfs/objects
// against the server. Anything the server already has does not need to be
// uploaded again.
func (c *uploadContext) checkMissing(missing []*lfs.WrappedPointer, missingSize int64) {
	numMissing := len(missing)
	if numMissing == 0 {
		return
	}

	checkQueue := lfs.NewDownloadCheckQueue(numMissing, missingSize)
	transferCh := checkQueue.Watch()

	done := make(chan int)
	go func() {
		// this channel is filled with oids for which Check() succeeded
		// and Transfer() was called
		for oid := range transferCh {
			c.SetUploaded(oid)
		}
		done <- 1
	}()

	for _, p := range missing {
		checkQueue.Add(lfs.NewDownloadable(p))
	}

	// Currently this is needed to flush the batch but is not enough to sync
	// transferc completely. By the time that checkQueue.Wait() returns, the
	// transferCh will have been closed, allowing the goroutine above to
	// send "1" into the `done` channel.
	checkQueue.Wait()
	<-done
}

func upload(c *uploadContext, unfiltered []*lfs.WrappedPointer) {
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

	q, pointers := c.prepareUpload(unfiltered)
	for _, p := range pointers {
		u, err := lfs.NewUploadable(p.Oid, p.Name)
		if err != nil {
			if errors.IsCleanPointerError(err) {
				Exit(uploadMissingErr, p.Oid, p.Name, errors.GetContext(err, "pointer").(*lfs.Pointer).Oid)
			} else {
				ExitWithError(err)
			}
		}

		q.Add(u)
		c.SetUploaded(p.Oid)
	}

	q.Wait()

	for _, err := range q.Errors() {
		FullError(err)
	}

	if len(q.Errors()) > 0 {
		os.Exit(2)
	}
<<<<<<< HEAD
=======

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
>>>>>>> f8a50160... Merge branch 'master' into no-dwarf-tables
}
