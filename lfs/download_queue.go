package lfs

type Downloadable struct {
	Pointer *WrappedPointer
	object  *objectResource
}

func NewDownloadable(p *WrappedPointer) *Downloadable {
	return &Downloadable{Pointer: p}
}

func (d *Downloadable) Check() (*objectResource, *WrappedError) {
	return DownloadCheck(d.Pointer.Oid)
}

func (d *Downloadable) Transfer(cb CopyCallback) *WrappedError {
	err := PointerSmudgeObject(d.Pointer.Pointer, d.object, cb)
	if err != nil {
		return Error(err)
	}
	return nil
}

func (d *Downloadable) Object() *objectResource {
	return d.object
}

func (d *Downloadable) Oid() string {
	return d.Pointer.Oid
}

func (d *Downloadable) Size() int64 {
	return d.Pointer.Size
}

func (d *Downloadable) Name() string {
	return d.Pointer.Name
}

func (d *Downloadable) SetObject(o *objectResource) {
	d.object = o
}

// NewDownloadQueue builds a DownloadQueue, allowing `workers` concurrent downloads.
func NewDownloadQueue(files int, size int64, dryRun bool) *TransferQueue {
	q := newTransferQueue(files, size, dryRun)
	q.transferKind = "download"
	return q
}
