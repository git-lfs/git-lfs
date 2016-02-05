package lfs

// The ability to check that a file can be downloaded
type DownloadCheckable struct {
	Pointer *WrappedPointer
	object  *ObjectResource
}

func NewDownloadCheckable(p *WrappedPointer) *DownloadCheckable {
	return &DownloadCheckable{Pointer: p}
}

func (d *DownloadCheckable) Check() (*ObjectResource, error) {
	return DownloadCheck(d.Pointer.Oid)
}

func (d *DownloadCheckable) Transfer(cb CopyCallback) error {
	// just report completion of check but don't do anything
	cb(d.Size(), d.Size(), int(d.Size()))
	return nil
}

func (d *DownloadCheckable) Object() *ObjectResource {
	return d.object
}

func (d *DownloadCheckable) Oid() string {
	return d.Pointer.Oid
}

func (d *DownloadCheckable) Size() int64 {
	return d.Pointer.Size
}

func (d *DownloadCheckable) Name() string {
	return d.Pointer.Name
}

func (d *DownloadCheckable) SetObject(o *ObjectResource) {
	d.object = o
}

// NewDownloadCheckQueue builds a checking queue, allowing `workers` concurrent check operations.
func NewDownloadCheckQueue(files int, size int64, dryRun bool) *TransferQueue {
	q := newTransferQueue(files, size, dryRun)
	// API operation is still download, but it will only perform the API call (check)
	q.transferKind = "download"
	return q
}

// The ability to actually download
type Downloadable struct {
	*DownloadCheckable
}

func NewDownloadable(p *WrappedPointer) *Downloadable {
	return &Downloadable{DownloadCheckable: NewDownloadCheckable(p)}
}

func (d *Downloadable) Transfer(cb CopyCallback) error {
	err := PointerSmudgeObject(d.Pointer.Pointer, d.object, cb)
	if err != nil {
		return Error(err)
	}
	return nil
}

// NewDownloadQueue builds a DownloadQueue, allowing `workers` concurrent downloads.
func NewDownloadQueue(files int, size int64, dryRun bool, metadata *TransferMetadata) *TransferQueue {
	q := newTransferQueue(files, size, dryRun)
	q.transferKind = "download"
	q.metadata = metadata
	return q
}
