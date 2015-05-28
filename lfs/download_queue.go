package lfs

type Downloadable struct {
	Pointer *WrappedPointer
	object  *ObjectResource
}

func NewDownloadable(p *WrappedPointer) *Downloadable {
	return &Downloadable{Pointer: p}
}

func (d *Downloadable) Check() (*ObjectResource, *WrappedError) {
	return DownloadCheck(d.Pointer.Oid)
}

func (d *Downloadable) Transfer(cb CopyCallback) *WrappedError {
	err := PointerSmudgeObject(d.Pointer.Pointer, d.object, cb)
	if err != nil {
		return Error(err)
	}
	return nil
}

func (d *Downloadable) Object() *ObjectResource {
	return d.object
}

func (d *Downloadable) Oid() string {
	return d.Pointer.Oid
}

func (d *Downloadable) Size() int64 {
	return d.Pointer.Size
}

func (d *Downloadable) SetObject(o *ObjectResource) {
	d.object = o
}

// NewDownloadQueue builds a DownloadQueue, allowing `workers` concurrent downloads.
func NewDownloadQueue(workers, files int) *TransferQueue {
	q := newTransferQueue(workers, files)
	q.transferKind = "download"
	return q
}
