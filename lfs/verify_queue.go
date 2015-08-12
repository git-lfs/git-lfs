package lfs

// This is like DownloadQueue except it doesn't do the transfer, it just uses
// the "download" API point to verify that the server has objects

type Verifiable struct {
	Pointer *WrappedPointer
	object  *objectResource
}

func NewVerifiable(p *WrappedPointer) *Verifiable {
	return &Verifiable{Pointer: p}
}

func (d *Verifiable) Check() (*objectResource, *WrappedError) {
	obj, wrerr := DownloadCheck(d.Pointer.Oid)

	if wrerr != nil {
		// Add some extra useful context
		wrerr.Set("oid", d.Pointer.Oid)
		wrerr.Set("name", d.Pointer.Name)
	}
	return obj, wrerr
}

func (d *Verifiable) Transfer(cb CopyCallback) *WrappedError {
	// just report completion of check but don't do anything
	cb(d.Size(), d.Size(), int(d.Size()))
	return nil
}

func (d *Verifiable) Object() *objectResource {
	return d.object
}

func (d *Verifiable) Oid() string {
	return d.Pointer.Oid
}

func (d *Verifiable) Size() int64 {
	return d.Pointer.Size
}

func (d *Verifiable) Name() string {
	return d.Pointer.Name
}

func (d *Verifiable) SetObject(o *objectResource) {
	d.object = o
}

// NewVerifyQueue builds a VerifyQueue, allowing `workers` concurrent verify operations.
func NewVerifyQueue(files int, size int64, dryRun bool) *TransferQueue {
	q := newTransferQueue(files, size, dryRun)
	// operation is still download, but it will only perform the API call (check)
	q.transferKind = "download"
	return q
}
