package commands

import (
	"os"

	"github.com/github/git-lfs/lfs"
)

var uploadMissingErr = "%s does not exist in .git/lfs/objects. Tried %s, which matches %s."

type uploadContext struct {
	RemoteName   string
	DryRun       bool
	uploadedOids lfs.StringSet
}

func newUploadContext() *uploadContext {
	return &uploadContext{
		uploadedOids: lfs.NewStringSet(),
		RemoteName:   lfs.Config.CurrentRemote,
	}
}

func (c *uploadContext) Upload(unfilteredPointers []*lfs.WrappedPointer) {
	filtered := c.filterUploadedObjects(noSkip, unfilteredPointers)

	totalSize := int64(0)
	for _, p := range filtered {
		totalSize += p.Size
	}

	uploadQueue := lfs.NewUploadQueue(len(filtered), totalSize, c.DryRun)

	if c.DryRun {
		for _, pointer := range filtered {
			Print("push %s => %s", pointer.Oid, pointer.Name)
			c.uploadedOids.Add(pointer.Oid)
		}
		return
	}

	c.filterServerObjects(filtered)
	pointers := c.filterUploadedObjects(uploadQueue, filtered)

	for _, pointer := range pointers {
		u, err := lfs.NewUploadable(pointer.Oid, pointer.Name)
		if err != nil {
			if lfs.IsCleanPointerError(err) {
				Exit(uploadMissingErr, pointer.Oid, pointer.Name, lfs.ErrorGetContext(err, "pointer").(*lfs.Pointer).Oid)
			} else {
				ExitWithError(err)
			}
		}

		uploadQueue.Add(u)
		c.uploadedOids.Add(pointer.Oid)
	}

	uploadQueue.Wait()
	for _, err := range uploadQueue.Errors() {
		if Debugging || lfs.IsFatalError(err) {
			LoggedError(err, err.Error())
		} else {
			if inner := lfs.GetInnerError(err); inner != nil {
				Error(inner.Error())
			}
			Error(err.Error())
		}
	}

	if len(uploadQueue.Errors()) > 0 {
		os.Exit(2)
	}
}

func (c *uploadContext) filterUploadedObjects(q transferQueueSkip, pointers []*lfs.WrappedPointer) []*lfs.WrappedPointer {
	filtered := make([]*lfs.WrappedPointer, 0, len(pointers))
	for _, pointer := range pointers {
		if c.uploadedOids.Contains(pointer.Oid) {
			q.Skip(pointer.Size)
		} else {
			filtered = append(filtered, pointer)
		}
	}

	return filtered
}

func (c *uploadContext) filterServerObjects(pointers []*lfs.WrappedPointer) {
	missingLocalObjects := make([]*lfs.WrappedPointer, 0, len(pointers))
	missingSize := int64(0)
	for _, pointer := range pointers {
		if !lfs.ObjectExistsOfSize(pointer.Oid, pointer.Size) {
			// We think we need to push this but we don't have it
			// Store for server checking later
			missingLocalObjects = append(missingLocalObjects, pointer)
			missingSize += pointer.Size
		}
	}
	if len(missingLocalObjects) == 0 {
		return
	}

	checkQueue := lfs.NewDownloadCheckQueue(len(missingLocalObjects), missingSize, true)
	for _, p := range missingLocalObjects {
		checkQueue.Add(lfs.NewDownloadCheckable(p))
	}
	// this channel is filled with oids for which Check() succeeded & Transfer() was called
	transferc := checkQueue.Watch()
	done := make(chan int)
	go func() {
		for oid := range transferc {
			c.uploadedOids.Add(oid)
		}
		done <- 1
	}()
	// Currently this is needed to flush the batch but is not enough to sync transferc completely
	checkQueue.Wait()
	<-done
}

type transferQueueSkip interface {
	Skip(int64)
}

type skipNoOp struct{}

func (s *skipNoOp) Skip(n int64) {}

var noSkip = &skipNoOp{}
