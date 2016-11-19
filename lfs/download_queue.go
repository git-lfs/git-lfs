package lfs

import (
	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/transfer"
)

type Downloadable struct {
	pointer *WrappedPointer
	object  *api.ObjectResource
}

func (d *Downloadable) Object() *api.ObjectResource {
	return d.object
}

func (d *Downloadable) Oid() string {
	return d.pointer.Oid
}

func (d *Downloadable) Size() int64 {
	return d.pointer.Size
}

func (d *Downloadable) Name() string {
	return d.pointer.Name
}

func (d *Downloadable) Path() string {
	p, _ := LocalMediaPath(d.pointer.Oid)
	return p
}

func (d *Downloadable) SetObject(o *api.ObjectResource) {
	d.object = o
}

func NewDownloadable(p *WrappedPointer) *Downloadable {
	return &Downloadable{pointer: p}
}

// NewDownloadCheckQueue builds a checking queue, checks that objects are there but doesn't download
func NewDownloadCheckQueue(files int, size int64) *TransferQueue {
	// Always dry run
	return newTransferQueue(files, size, true, transfer.Download)
}

// NewDownloadQueue builds a DownloadQueue, allowing concurrent downloads.
func NewDownloadQueue(files int, size int64, dryRun bool) *TransferQueue {
	return newTransferQueue(files, size, dryRun, transfer.Download)
}
