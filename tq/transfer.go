// Package transfer collects together adapters for uploading and downloading LFS content
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package tq

import (
	"fmt"
	"time"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/errors"
)

type Direction int

const (
	Upload   = Direction(iota)
	Download = Direction(iota)
)

type Transfer struct {
	Name          string       `json:"name"`
	Oid           string       `json:"oid,omitempty"`
	Size          int64        `json:"size"`
	Authenticated bool         `json:"authenticated,omitempty"`
	Actions       ActionSet    `json:"actions,omitempty"`
	Error         *ObjectError `json:"error,omitempty"`
	Path          string       `json:"path"`
}

type ObjectError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// newTransfer creates a new Transfer instance
func newTransfer(name string, obj *api.ObjectResource, path string) *Transfer {
	t := &Transfer{
		Name:          name,
		Oid:           obj.Oid,
		Size:          obj.Size,
		Authenticated: obj.Authenticated,
		Actions:       make(ActionSet),
		Path:          path,
	}

	if obj.Error != nil {
		t.Error = &ObjectError{
			Code:    obj.Error.Code,
			Message: obj.Error.Message,
		}
	}

	for rel, action := range obj.Actions {
		t.Actions[rel] = &Action{
			Href:      action.Href,
			Header:    action.Header,
			ExpiresAt: action.ExpiresAt,
		}
	}

	return t

}

type Action struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header,omitempty"`
	ExpiresAt time.Time         `json:"expires_at,omitempty"`
}

type ActionSet map[string]*Action

const (
	// objectExpirationToTransfer is the duration we expect to have passed
	// from the time that the object's expires_at property is checked to
	// when the transfer is executed.
	objectExpirationToTransfer = 5 * time.Second
)

func (as ActionSet) Get(rel string) (*Action, error) {
	a, ok := as[rel]
	if !ok {
		return nil, &ActionMissingError{Rel: rel}
	}

	if !a.ExpiresAt.IsZero() && a.ExpiresAt.Before(time.Now().Add(objectExpirationToTransfer)) {
		return nil, errors.NewRetriableError(&ActionExpiredErr{Rel: rel, At: a.ExpiresAt})
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

type ActionMissingError struct {
	Rel string
}

func (e ActionMissingError) Error() string {
	return fmt.Sprintf("tq: unable to find action %q", e.Rel)
}

func IsActionExpiredError(err error) bool {
	if _, ok := err.(*ActionExpiredErr); ok {
		return true
	}
	return false
}

func IsActionMissingError(err error) bool {
	if _, ok := err.(*ActionMissingError); ok {
		return true
	}
	return false
}

func toApiObject(t *Transfer) *api.ObjectResource {
	o := &api.ObjectResource{
		Oid:           t.Oid,
		Size:          t.Size,
		Authenticated: t.Authenticated,
		Actions:       make(map[string]*api.LinkRelation),
	}

	for rel, a := range t.Actions {
		o.Actions[rel] = &api.LinkRelation{
			Href:      a.Href,
			Header:    a.Header,
			ExpiresAt: a.ExpiresAt,
		}
	}

	if t.Error != nil {
		o.Error = &api.ObjectError{
			Code:    t.Error.Code,
			Message: t.Error.Message,
		}
	}

	return o
}

// NewAdapterFunc creates new instances of Adapter. Code that wishes
// to provide new Adapter instances should pass an implementation of this
// function to RegisterNewTransferAdapterFunc() on a *Manifest.
// name and dir are to provide context if one func implements many instances
type NewAdapterFunc func(name string, dir Direction) Adapter

type ProgressCallback func(name string, totalSize, readSoFar int64, readSinceLast int) error

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
	Begin(maxConcurrency int, cb ProgressCallback) error
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
