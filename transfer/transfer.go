// Package transfer collects together adapters for uploading and downloading LFS content
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package transfer

import (
	"sync"

	"github.com/github/git-lfs/config"

	"github.com/github/git-lfs/api"
	"github.com/rubyist/tracerx"
)

type Direction int

const (
	Upload   = Direction(iota)
	Download = Direction(iota)
)

// NewTransferAdapterFunc creates new instances of TransferAdapter. Code that wishes
// to provide new TransferAdapter instances should pass an implementation of this
// function to RegisterNewTransferAdapterFunc
// name and dir are to provide context if one func implements many instances
type NewTransferAdapterFunc func(name string, dir Direction) TransferAdapter

var (
	funcMutex            sync.Mutex
	downloadAdapterFuncs = make(map[string]NewTransferAdapterFunc)
	uploadAdapterFuncs   = make(map[string]NewTransferAdapterFunc)
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
	// Name returns the name of this adapter, which is the same for all instances
	// of this type of adapter
	Name() string
	// Direction returns whether this instance is an upload or download instance
	// TransferAdapter instances can only be one or the other, although the same
	// type may be instantiated for each direction
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

// NewTransfer creates a new Transfer instance
func NewTransfer(name string, obj *api.ObjectResource, path string) *Transfer {
	return &Transfer{name, obj, path}
}

// Result of a transfer returned through CompletionChannel()
type TransferResult struct {
	Transfer *Transfer
	// This will be non-nil if there was an error transferring this item
	Error error
}

// GetAdapterNames returns a list of the names of adapters available to be created
func GetAdapterNames(dir Direction) []string {
	switch dir {
	case Upload:
		return GetUploadAdapterNames()
	case Download:
		return GetDownloadAdapterNames()
	}
	return nil
}

// GetDownloadAdapterNames returns a list of the names of download adapters available to be created
func GetDownloadAdapterNames() []string {

	if config.Config.BasicTransfersOnly() {
		return []string{BasicAdapterName}
	}

	initCoreAdaptersIfRequired()

	funcMutex.Lock()
	defer funcMutex.Unlock()

	ret := make([]string, 0, len(downloadAdapterFuncs))
	for n, _ := range downloadAdapterFuncs {
		ret = append(ret, n)
	}
	return ret
}

// GetUploadAdapterNames returns a list of the names of upload adapters available to be created
func GetUploadAdapterNames() []string {

	if config.Config.BasicTransfersOnly() {
		return []string{BasicAdapterName}
	}

	initCoreAdaptersIfRequired()

	funcMutex.Lock()
	defer funcMutex.Unlock()

	ret := make([]string, 0, len(uploadAdapterFuncs))
	for n, _ := range uploadAdapterFuncs {
		ret = append(ret, n)
	}
	return ret
}

// RegisterNewTransferAdapterFunc registers a new function for creating upload
// or download adapters. If a function with that name & direction is already
// registered, it is overridden
func RegisterNewTransferAdapterFunc(name string, dir Direction, f NewTransferAdapterFunc) {
	funcMutex.Lock()
	defer funcMutex.Unlock()

	switch dir {
	case Upload:
		uploadAdapterFuncs[name] = f
	case Download:
		downloadAdapterFuncs[name] = f
	}
}

// Create a new adapter by name and direction; default to BasicAdapterName if doesn't exist
func NewAdapterOrDefault(name string, dir Direction) TransferAdapter {
	if len(name) == 0 {
		name = BasicAdapterName
	}

	a := NewAdapter(name, dir)
	if a == nil {
		tracerx.Printf("Defaulting to basic transfer adapter since %q did not exist", name)
		a = NewAdapter(BasicAdapterName, dir)
	}
	return a
}

// Create a new adapter by name and direction, or nil if doesn't exist
func NewAdapter(name string, dir Direction) TransferAdapter {
	initCoreAdaptersIfRequired()

	funcMutex.Lock()
	defer funcMutex.Unlock()

	switch dir {
	case Upload:
		if u, ok := uploadAdapterFuncs[name]; ok {
			return u(name, dir)
		}
	case Download:
		if d, ok := downloadAdapterFuncs[name]; ok {
			return d(name, dir)
		}
	}
	return nil
}

// Create a new download adapter by name, or BasicAdapterName if doesn't exist
func NewDownloadAdapter(name string) TransferAdapter {
	return NewAdapterOrDefault(name, Download)
}

// Create a new upload adapter by name, or BasicAdapterName if doesn't exist
func NewUploadAdapter(name string) TransferAdapter {
	return NewAdapterOrDefault(name, Upload)
}

var initCoreOnce sync.Once

func initCoreAdaptersIfRequired() {
	// It's important to late-init custom adapters because they rely on Config
	// And if we cause Config to load too early it causes issues
	// That's why this isn't in an init() block
	initCoreOnce.Do(func() {
		ConfigureCustomAdapters()
		if !config.Config.TusTransfersAllowed() {
			delete(uploadAdapterFuncs, TusAdapterName)
		}
	})
}
