package commands

import (
	"os"

	"github.com/github/git-lfs/lfs"
)

type clientContext struct {
	RemoteName   string
	DryRun       bool
	uploadedOids lfs.StringSet
}

func newClient() *clientContext {
	return &clientContext{
		uploadedOids: lfs.NewStringSet(),
	}
}

func (c *clientContext) Upload(unfilteredPointers []*lfs.WrappedPointer) {
	pointers := c.filter(unfilteredPointers)

	if c.DryRun {
		for _, pointer := range pointers {
			Print("push %s => %s", pointer.Oid, pointer.Name)
			c.uploadedOids.Add(pointer.Oid)
		}
		return
	}

	totalSize := int64(0)
	for _, p := range pointers {
		totalSize += p.Size
	}

	uploadQueue := lfs.NewUploadQueue(len(pointers), totalSize, false)
	for _, pointer := range pointers {
		u, err := lfs.NewUploadable(pointer.Oid, pointer.Name)
		if err != nil {
			if lfs.IsCleanPointerError(err) {
				Exit(prePushMissingErrMsg, pointer.Name, lfs.ErrorGetContext(err, "pointer").(*lfs.Pointer).Oid)
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

func (c *clientContext) filter(pointers []*lfs.WrappedPointer) []*lfs.WrappedPointer {
	uploadable := c.filterUploadedObjects(pointers)
	c.filterServerObjects(uploadable)
	return c.filterUploadedObjects(uploadable)
}

func (c *clientContext) filterUploadedObjects(pointers []*lfs.WrappedPointer) []*lfs.WrappedPointer {
	filtered := make([]*lfs.WrappedPointer, 0, len(pointers))
	for _, pointer := range pointers {
		if !c.uploadedOids.Contains(pointer.Oid) {
			filtered = append(filtered, pointer)
		}
	}

	return filtered
}

func (c *clientContext) filterServerObjects(pointers []*lfs.WrappedPointer) {
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
