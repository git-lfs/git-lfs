package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/github/git-lfs/httputil"
)

type ObjectError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ObjectError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

type ObjectResource struct {
	Oid           string                   `json:"oid,omitempty"`
	Size          int64                    `json:"size"`
	Authenticated bool                     `json:"authenticated,omitempty"`
	Actions       map[string]*LinkRelation `json:"actions,omitempty"`
	Links         map[string]*LinkRelation `json:"_links,omitempty"`
	Error         *ObjectError             `json:"error,omitempty"`
}

// TODO LEGACY API: remove when legacy API removed
func (o *ObjectResource) NewRequest(relation, method string) (*http.Request, error) {
	rel, ok := o.Rel(relation)
	if !ok {
		if relation == "download" {
			return nil, errors.New("Object not found on the server.")
		}
		return nil, fmt.Errorf("No %q action for this object.", relation)

	}

	req, err := httputil.NewHttpRequest(method, rel.Href, rel.Header)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (o *ObjectResource) Rel(name string) (*LinkRelation, bool) {
	var rel *LinkRelation
	var ok bool

	if o.Actions != nil {
		rel, ok = o.Actions[name]
	} else {
		rel, ok = o.Links[name]
	}

	return rel, ok
}

// IsExpired returns true if any of the actions in this object resource have an
// ExpiresAt field that is after the given instant "now".
//
// If the object contains no actions, or none of the actions it does contain
// have non-zero ExpiresAt fields, the object is not expired.
func (o *ObjectResource) IsExpired(now time.Time) bool {
	for _, a := range o.Actions {
		if !a.ExpiresAt.IsZero() && a.ExpiresAt.Before(now) {
			return true
		}
	}

	return false
}

func (o *ObjectResource) NeedsAuth() bool {
	return !o.Authenticated
}

type LinkRelation struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header,omitempty"`
	ExpiresAt time.Time         `json:"expires_at,omitempty"`
}
