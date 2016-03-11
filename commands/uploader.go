package commands

import (
	"os"

	"github.com/github/git-lfs/lfs"
)

func scanObjectsLeftToRight(config *lfs.Configuration, left, right string, scanmode lfs.ScanningMode) ([]*lfs.WrappedPointer, error) {
	scanOpt := lfs.NewScanRefsOptions()
	scanOpt.ScanMode = scanmode
	scanOpt.RemoteName = config.CurrentRemote

	pointers, err := lfs.ScanRefs(left, right, scanOpt)
	if err != nil {
		return pointers, err
	}

	return filterUploadableObjects(config, pointers), nil
}

func uploadObjects(config *lfs.Configuration, pointers []*lfs.WrappedPointer, dryRun bool) {
	if dryRun {
		for _, pointer := range pointers {
			Print("push %s => %s", pointer.Oid, pointer.Name)
			config.Uploaded(pointer.Oid)
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
		config.Uploaded(pointer.Oid)
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

func filterUploadableObjects(config *lfs.Configuration, pointers []*lfs.WrappedPointer) []*lfs.WrappedPointer {
	uploadable := filterUploadedObjects(config, pointers)
	filterServerObjects(config, uploadable)
	return filterUploadedObjects(config, uploadable)
}

func filterUploadedObjects(config *lfs.Configuration, pointers []*lfs.WrappedPointer) []*lfs.WrappedPointer {
	filtered := make([]*lfs.WrappedPointer, 0, len(pointers))
	for _, pointer := range pointers {
		if !config.HasUploaded(pointer.Oid) {
			filtered = append(filtered, pointer)
		}
	}

	return filtered
}

func filterServerObjects(config *lfs.Configuration, pointers []*lfs.WrappedPointer) {
	var missingLocalObjects []*lfs.WrappedPointer
	var missingSize int64
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
			config.Uploaded(oid)
		}
		done <- 1
	}()
	// Currently this is needed to flush the batch but is not enough to sync transferc completely
	checkQueue.Wait()
	<-done
}
