package lfs

// The ability to check that a file can be downloaded
type DownloadCheckable struct {
	Pointer  *WrappedPointer
	object   *objectResource
	attempts int
}

func NewDownloadCheckable(p *WrappedPointer) *DownloadCheckable {
	return &DownloadCheckable{Pointer: p}
}

func (d *DownloadCheckable) Check() (*objectResource, error) {
	return DownloadCheck(d.Pointer.Oid)
}

func (d *DownloadCheckable) Transfer(cb CopyCallback) error {
	// just report completion of check but don't do anything
	d.attempts++
	cb(d.Size(), d.Size(), int(d.Size()))
	return nil
}

func (d *DownloadCheckable) Object() *objectResource {
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

func (d *DownloadCheckable) Attempts() int {
	return d.attempts
}

func (d *DownloadCheckable) SetObject(o *objectResource) {
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
	attempts int
	*DownloadCheckable
}

func NewDownloadable(p *WrappedPointer) *Downloadable {
	return &Downloadable{DownloadCheckable: NewDownloadCheckable(p)}
}

func (d *Downloadable) Transfer(cb CopyCallback) error {
	d.attempts++
	err := PointerSmudgeObject(d.Pointer.Pointer, d.object, cb)
	if err != nil {
		return Error(err)
	}
	return nil
}

func (d *Downloadable) Attempts() int {
	return d.attempts
}

// NewDownloadQueue builds a DownloadQueue, allowing `workers` concurrent downloads.
func NewDownloadQueue(files int, size int64, dryRun bool) *TransferQueue {
	q := newTransferQueue(files, size, dryRun)
	q.transferKind = "download"
	return q
}
