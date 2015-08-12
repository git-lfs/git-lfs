package lfs

// This is like DownloadQueue except it doesn't do the transfer, it just uses
// the "download" API point to verify that the server has objects

type Checkable struct {
	Pointer *WrappedPointer
	object  *objectResource
}

func NewCheckable(p *WrappedPointer) *Checkable {
	return &Checkable{Pointer: p}
}

func (d *Checkable) Check() (*objectResource, *WrappedError) {
	return DownloadCheck(d.Pointer.Oid)
}

func (d *Checkable) Transfer(cb CopyCallback) *WrappedError {
	// just report completion of check but don't do anything
	cb(d.Size(), d.Size(), int(d.Size()))
	return nil
}

func (d *Checkable) Object() *objectResource {
	return d.object
}

func (d *Checkable) Oid() string {
	return d.Pointer.Oid
}

func (d *Checkable) Size() int64 {
	return d.Pointer.Size
}

func (d *Checkable) Name() string {
	return d.Pointer.Name
}

func (d *Checkable) SetObject(o *objectResource) {
	d.object = o
}

// NewCheckQueue builds a checking queue, allowing `workers` concurrent check operations.
func NewCheckQueue(files int, size int64, dryRun bool) *TransferQueue {
	q := newTransferQueue(files, size, dryRun)
	// API operation is still download, but it will only perform the API call (check)
	q.transferKind = "download"
	return q
}
