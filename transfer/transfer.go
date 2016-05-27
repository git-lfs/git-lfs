// Package transfer collects together adapters for uploading and downloading LFS content
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package transfer

import (
	"sync"

	"github.com/github/git-lfs/api"
)

type Direction int

const (
	Upload   = Direction(iota)
	Download = Direction(iota)
)

var (
	adapterMutex     sync.Mutex
	downloadAdapters = make(map[string]TransferAdapter)
	uploadAdapters   = make(map[string]TransferAdapter)
)

type TransferProgressCallback func(name string, totalSize, readSoFar int64, readSinceLast int) error

// TransferAdapter is implemented by types which can upload and/or download LFS
// file content to a remote store. Each TransferAdapter accepts one or more requests
// which it may schedule and parallelise in whatever way it chooses, clients of
// this interface will receive notifications of progress and completion asynchronously.
// TransferAdapters support transfers in one direction; if an implementation
// provides support for upload and download, it should be instantiated twice,
// advertising support for each direction separately.
// Note that TransferAdapter only implements the actual upload/download of content
// itself; organising the wider process including calling the API to get URLs,
// handling progress reporting and retries is the job of the core TransferQueue.
// This is so that the orchestration remains core & standard but TransferAdapter
// can be changed to physically transfer to different hosts with less code.
type TransferAdapter interface {
	// Name returns the identifier of this adapter, must be unique within a Direction
	// (separate sets for upload and download so may be an entry in both)
	Name() string
	// Direction returns whether this instance is an upload or download instance
	// TransferAdapter instances can only be one or the other, although the same
	// type may be instantiated once for each direction
	Direction() Direction
	// Begin a new batch of uploads or downloads. Call this first, followed by
	// one or more Add calls. maxConcurrency controls the number of transfers
	// that may be done at once. The passed in callback will receive updates on
	// progress, and the completion channel will receive completion notifications
	// Either argument may be nil if not required by the client
	Begin(maxConcurrency int, cb TransferProgressCallback, completion chan TransferResult) error
	// Add queues a download/upload, which will complete asynchronously and
	// notify the callbacks given to Begin()
	Add(t *Transfer)
	// Indicate that all transfers have been scheduled and resources can be released
	// once the queued items have completed.
	// This call blocks until all items have been processed
	End()
	// ClearTempStorage clears any temporary files, such as unfinished downloads that
	// would otherwise be resumed
	ClearTempStorage() error
}

// General struct for both uploads and downloads
type Transfer struct {
	// Name of the file that triggered this transfer
	Name string
	// Object from API which provides the core data for this transfer
	Object *api.ObjectResource
	// Path for uploads is the source of data to send, for downloads is the
	// location to place the final result
	Path string
}

func NewTransfer(name string, obj *api.ObjectResource, path string) *Transfer {
	return &Transfer{name, obj, path}
}

// Result of a transfer returned through CompletionChannel()
type TransferResult struct {
	Transfer *Transfer
	// This will be non-nil if there was an error transferring this item
	Error error
}

// GetAdapters returns a list of registered adapters for the given direction
func GetAdapters(dir Direction) []TransferAdapter {
	switch dir {
	case Upload:
		return GetUploadAdapters()
	case Download:
		return GetDownloadAdapters()
	}
	return nil
}

// GetDownloadAdapters returns a list of registered adapters able to perform downloads
func GetDownloadAdapters() []TransferAdapter {
	adapterMutex.Lock()
	defer adapterMutex.Unlock()

	ret := make([]TransferAdapter, 0, len(downloadAdapters))
	for _, a := range downloadAdapters {
		ret = append(ret, a)
	}
	return ret
}

// GetUploadAdapters returns a list of registered adapters able to perform uploads
func GetUploadAdapters() []TransferAdapter {
	adapterMutex.Lock()
	defer adapterMutex.Unlock()

	ret := make([]TransferAdapter, 0, len(uploadAdapters))
	for _, a := range uploadAdapters {
		ret = append(ret, a)
	}
	return ret
}

// RegisterAdapter registers an upload or download adapter. If an adapter is
// already registered for that direction with the same name, it is overridden
func RegisterAdapter(adapter TransferAdapter) {
	adapterMutex.Lock()
	defer adapterMutex.Unlock()

	switch adapter.Direction() {
	case Upload:
		uploadAdapters[adapter.Name()] = adapter
	case Download:
		downloadAdapters[adapter.Name()] = adapter
	}
}

// Get a specific adapter by name and direction, or nil if doesn't exist
func GetAdapter(name string, dir Direction) TransferAdapter {
	adapterMutex.Lock()
	defer adapterMutex.Unlock()

	switch dir {
	case Upload:
		if u, ok := uploadAdapters[name]; ok {
			return u
		}
	case Download:
		if d, ok := downloadAdapters[name]; ok {
			return d
		}
	}
	return nil
}

func GetDownloadAdapter(name string) TransferAdapter {
	return GetAdapter(name, Download)
}
func GetUploadAdapter(name string) TransferAdapter {
	return GetAdapter(name, Upload)
}
