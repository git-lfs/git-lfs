// Package transfer collects together adapters for uploading and downloading LFS content
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package tq

import (
	"fmt"
	"time"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/tools"
)

type Direction int

const (
	Upload   = Direction(iota)
	Download = Direction(iota)
	Checkout = Direction(iota)
)

// Verb returns a string containing the verb form of the receiving action.
func (d Direction) Verb() string {
	switch d {
	case Checkout:
		return "Checking out"
	case Download:
		return "Downloading"
	case Upload:
		return "Uploading"
	default:
		return "<unknown>"
	}
}

func (d Direction) String() string {
	switch d {
	case Checkout:
		return "checkout"
	case Download:
		return "download"
	case Upload:
		return "upload"
	default:
		return "<unknown>"
	}
}

type Transfer struct {
	Name          string       `json:"name,omitempty"`
	Oid           string       `json:"oid,omitempty"`
	Size          int64        `json:"size"`
	Authenticated bool         `json:"authenticated,omitempty"`
	Actions       ActionSet    `json:"actions,omitempty"`
	Links         ActionSet    `json:"_links,omitempty"`
	Error         *ObjectError `json:"error,omitempty"`
	Path          string       `json:"path,omitempty"`
	Missing       bool         `json:"-"`
}

func (t *Transfer) Rel(name string) (*Action, error) {
	a, err := t.Actions.Get(name)
	if a != nil || err != nil {
		return a, err
	}

	if t.Links != nil {
		a, err := t.Links.Get(name)
		if a != nil || err != nil {
			return a, err
		}
	}

	return nil, nil
}

type ObjectError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ObjectError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// newTransfer returns a copy of the given Transfer, with the name and path
// values set.
func newTransfer(tr *Transfer, name string, path string) *Transfer {
	t := &Transfer{
		Name:          name,
		Path:          path,
		Oid:           tr.Oid,
		Size:          tr.Size,
		Authenticated: tr.Authenticated,
		Actions:       make(ActionSet),
	}

	if tr.Error != nil {
		t.Error = &ObjectError{
			Code:    tr.Error.Code,
			Message: tr.Error.Message,
		}
	}

	for rel, action := range tr.Actions {
		t.Actions[rel] = &Action{
			Href:      action.Href,
			Header:    action.Header,
			ExpiresAt: action.ExpiresAt,
			ExpiresIn: action.ExpiresIn,
			createdAt: action.createdAt,
		}
	}

	if tr.Links != nil {
		t.Links = make(ActionSet)

		for rel, link := range tr.Links {
			t.Links[rel] = &Action{
				Href:      link.Href,
				Header:    link.Header,
				ExpiresAt: link.ExpiresAt,
				ExpiresIn: link.ExpiresIn,
				createdAt: link.createdAt,
			}
		}
	}

	return t
}

type Action struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header,omitempty"`
	ExpiresAt time.Time         `json:"expires_at,omitempty"`
	ExpiresIn int               `json:"expires_in,omitempty"`

	createdAt time.Time
}

func (a *Action) IsExpiredWithin(d time.Duration) (time.Time, bool) {
	return tools.IsExpiredAtOrIn(a.createdAt, d, a.ExpiresAt, time.Duration(a.ExpiresIn)*time.Second)
}

type ActionSet map[string]*Action

const (
	// objectExpirationToTransfer is the duration we expect to have passed
	// from the time that the object's expires_at (or expires_in) property
	// is checked to when the transfer is executed.
	objectExpirationToTransfer = 5 * time.Second
)

func (as ActionSet) Get(rel string) (*Action, error) {
	a, ok := as[rel]
	if !ok {
		return nil, nil
	}

	if at, expired := a.IsExpiredWithin(objectExpirationToTransfer); expired {
		return nil, errors.NewRetriableError(&ActionExpiredErr{Rel: rel, At: at})
	}

	return a, nil
}

type ActionExpiredErr struct {
	Rel string
	At  time.Time
}

func (e ActionExpiredErr) Error() string {
	return fmt.Sprintf("tq: action %q expires at %s",
		e.Rel, e.At.In(time.Local).Format(time.RFC822))
}

func IsActionExpiredError(err error) bool {
	if _, ok := err.(*ActionExpiredErr); ok {
		return true
	}
	return false
}

// NewAdapterFunc creates new instances of Adapter. Code that wishes
// to provide new Adapter instances should pass an implementation of this
// function to RegisterNewTransferAdapterFunc() on a *Manifest.
// name and dir are to provide context if one func implements many instances
type NewAdapterFunc func(name string, dir Direction) Adapter

type ProgressCallback func(name string, totalSize, readSoFar int64, readSinceLast int) error

type AdapterConfig interface {
	APIClient() *lfsapi.Client
	ConcurrentTransfers() int
	Remote() string
}

type adapterConfig struct {
	apiClient           *lfsapi.Client
	concurrentTransfers int
	remote              string
}

func (c *adapterConfig) ConcurrentTransfers() int {
	return c.concurrentTransfers
}

func (c *adapterConfig) APIClient() *lfsapi.Client {
	return c.apiClient
}

func (c *adapterConfig) Remote() string {
	return c.remote
}

// Adapter is implemented by types which can upload and/or download LFS
// file content to a remote store. Each Adapter accepts one or more requests
// which it may schedule and parallelise in whatever way it chooses, clients of
// this interface will receive notifications of progress and completion asynchronously.
// TransferAdapters support transfers in one direction; if an implementation
// provides support for upload and download, it should be instantiated twice,
// advertising support for each direction separately.
// Note that Adapter only implements the actual upload/download of content
// itself; organising the wider process including calling the API to get URLs,
// handling progress reporting and retries is the job of the core TransferQueue.
// This is so that the orchestration remains core & standard but Adapter
// can be changed to physically transfer to different hosts with less code.
type Adapter interface {
	// Name returns the name of this adapter, which is the same for all instances
	// of this type of adapter
	Name() string
	// Direction returns whether this instance is an upload or download instance
	// Adapter instances can only be one or the other, although the same
	// type may be instantiated for each direction
	Direction() Direction
	// Begin a new batch of uploads or downloads. Call this first, followed by
	// one or more Add calls. maxConcurrency controls the number of transfers
	// that may be done at once. The passed in callback will receive updates on
	// progress. Either argument may be nil if not required by the client.
	Begin(cfg AdapterConfig, cb ProgressCallback) error
	// Add queues a download/upload, which will complete asynchronously and
	// notify the callbacks given to Begin()
	Add(transfers ...*Transfer) (results <-chan TransferResult)
	// Indicate that all transfers have been scheduled and resources can be released
	// once the queued items have completed.
	// This call blocks until all items have been processed
	End()
	// ClearTempStorage clears any temporary files, such as unfinished downloads that
	// would otherwise be resumed
	ClearTempStorage() error
}

// Result of a transfer returned through CompletionChannel()
type TransferResult struct {
	Transfer *Transfer
	// This will be non-nil if there was an error transferring this item
	Error error
}
