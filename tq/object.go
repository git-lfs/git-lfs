package tq

import (
	"fmt"
	"time"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/errors"
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

// NewTransfer creates a new Transfer instance
//
// XXX(taylor): note, this function will be removed from the `tq` package's API
// before landing in master. It is currently used by the smudge operation to
// download a single file and pass _directly_ to the Adapter, whereas it should
// use a transferqueue.
func NewTransfer(name string, obj *api.ObjectResource, path string) *Transfer {
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
