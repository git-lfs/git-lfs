package api

import (
	"errors"
	"fmt"
	"net/http"

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
	Oid     string                   `json:"oid,omitempty"`
	Size    int64                    `json:"size"`
	Actions map[string]*LinkRelation `json:"actions,omitempty"`
	Links   map[string]*LinkRelation `json:"_links,omitempty"`
	Error   *ObjectError             `json:"error,omitempty"`
}

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

type LinkRelation struct {
	Href   string            `json:"href"`
	Header map[string]string `json:"header,omitempty"`
}
