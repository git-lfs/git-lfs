package locking

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/lfshttp"
)

type lockClient struct {
	*lfsapi.Client
}

type lockRef struct {
	Name string `json:"name,omitempty"`
}

// LockRequest encapsulates the payload sent across the API when a client would
// like to obtain a lock against a particular path on a given remote.
type lockRequest struct {
	// Path is the path that the client would like to obtain a lock against.
	Path string   `json:"path"`
	Ref  *lockRef `json:"ref,omitempty"`
}

// LockResponse encapsulates the information sent over the API in response to
// a `LockRequest`.
type lockResponse struct {
	// Lock is the Lock that was optionally created in response to the
	// payload that was sent (see above). If the lock already exists, then
	// the existing lock is sent in this field instead, and the author of
	// that lock remains the same, meaning that the client failed to obtain
	// that lock. An HTTP status of "409 - Conflict" is used here.
	//
	// If the lock was unable to be created, this field will hold the
	// zero-value of Lock and the Err field will provide a more detailed set
	// of information.
	//
	// If an error was experienced in creating this lock, then the
	// zero-value of Lock should be sent here instead.
	Lock *Lock `json:"lock"`

	// Message is the optional error that was encountered while trying to create
	// the above lock.
	Message          string `json:"message,omitempty"`
	DocumentationURL string `json:"documentation_url,omitempty"`
	RequestID        string `json:"request_id,omitempty"`
}

func (c *lockClient) Lock(remote string, lockReq *lockRequest) (*lockResponse, *http.Response, error) {
	e := c.Endpoints.Endpoint("upload", remote)
	req, err := c.NewRequest("POST", e, "locks", lockReq)
	if err != nil {
		return nil, nil, err
	}

	req = c.Client.LogRequest(req, "lfs.locks.lock")
	res, err := c.DoAPIRequestWithAuth(remote, req)
	if err != nil {
		return nil, res, err
	}

	lockRes := &lockResponse{}
	err = lfshttp.DecodeJSON(res, lockRes)
	if err != nil {
		return nil, res, err
	}
	if lockRes.Lock == nil && len(lockRes.Message) == 0 {
		return nil, res, fmt.Errorf("invalid server response")
	}
	return lockRes, res, nil
}

// UnlockRequest encapsulates the data sent in an API request to remove a lock.
type unlockRequest struct {
	// Force determines whether or not the lock should be "forcibly"
	// unlocked; that is to say whether or not a given individual should be
	// able to break a different individual's lock.
	Force bool     `json:"force"`
	Ref   *lockRef `json:"ref,omitempty"`
}

// UnlockResponse is the result sent back from the API when asked to remove a
// lock.
type unlockResponse struct {
	// Lock is the lock corresponding to the asked-about lock in the
	// `UnlockPayload` (see above). If no matching lock was found, this
	// field will take the zero-value of Lock, and Err will be non-nil.
	Lock *Lock `json:"lock"`

	// Message is an optional field which holds any error that was experienced
	// while removing the lock.
	Message          string `json:"message,omitempty"`
	DocumentationURL string `json:"documentation_url,omitempty"`
	RequestID        string `json:"request_id,omitempty"`
}

func (c *lockClient) Unlock(ref *git.Ref, remote, id string, force bool) (*unlockResponse, *http.Response, error) {
	e := c.Endpoints.Endpoint("upload", remote)
	suffix := fmt.Sprintf("locks/%s/unlock", id)
	req, err := c.NewRequest("POST", e, suffix, &unlockRequest{
		Force: force,
		Ref:   &lockRef{Name: ref.Refspec()},
	})
	if err != nil {
		return nil, nil, err
	}

	req = c.Client.LogRequest(req, "lfs.locks.unlock")
	res, err := c.DoAPIRequestWithAuth(remote, req)
	if err != nil {
		return nil, res, err
	}

	unlockRes := &unlockResponse{}
	err = lfshttp.DecodeJSON(res, unlockRes)
	if err != nil {
		return nil, res, err
	}
	if unlockRes.Lock == nil && len(unlockRes.Message) == 0 {
		return nil, res, fmt.Errorf("invalid server response")
	}
	return unlockRes, res, nil
}

// Filter represents a single qualifier to apply against a set of locks.
type lockFilter struct {
	// Property is the property to search against.
	// Value is the value that the property must take.
	Property, Value string
}

// LockSearchRequest encapsulates the request sent to the server when the client
// would like a list of locks that match the given criteria.
type lockSearchRequest struct {
	// Filters is the set of filters to query against. If the client wishes
	// to obtain a list of all locks, an empty array should be passed here.
	Filters []lockFilter
	// Cursor is an optional field used to tell the server which lock was
	// seen last, if scanning through multiple pages of results.
	//
	// Servers must return a list of locks sorted in reverse chronological
	// order, so the Cursor provides a consistent method of viewing all
	// locks, even if more were created between two requests.
	Cursor string
	// Limit is the maximum number of locks to return in a single page.
	Limit int

	Refspec string
}

func (r *lockSearchRequest) QueryValues() map[string]string {
	q := make(map[string]string)
	for _, filter := range r.Filters {
		q[filter.Property] = filter.Value
	}

	if len(r.Cursor) > 0 {
		q["cursor"] = r.Cursor
	}

	if r.Limit > 0 {
		q["limit"] = strconv.Itoa(r.Limit)
	}

	if len(r.Refspec) > 0 {
		q["refspec"] = r.Refspec
	}

	return q
}

// LockList encapsulates a set of Locks.
type lockList struct {
	// Locks is the set of locks returned back, typically matching the query
	// parameters sent in the LockListRequest call. If no locks were matched
	// from a given query, then `Locks` will be represented as an empty
	// array.
	Locks []Lock `json:"locks"`
	// NextCursor returns the Id of the Lock the client should update its
	// cursor to, if there are multiple pages of results for a particular
	// `LockListRequest`.
	NextCursor string `json:"next_cursor,omitempty"`
	// Message populates any error that was encountered during the search. If no
	// error was encountered and the operation was successful, then a value
	// of nil will be passed here.
	Message          string `json:"message,omitempty"`
	DocumentationURL string `json:"documentation_url,omitempty"`
	RequestID        string `json:"request_id,omitempty"`
}

func (c *lockClient) Search(remote string, searchReq *lockSearchRequest) (*lockList, *http.Response, error) {
	e := c.Endpoints.Endpoint("download", remote)
	req, err := c.NewRequest("GET", e, "locks", nil)
	if err != nil {
		return nil, nil, err
	}

	q := req.URL.Query()
	for key, value := range searchReq.QueryValues() {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	req = c.Client.LogRequest(req, "lfs.locks.search")
	res, err := c.DoAPIRequestWithAuth(remote, req)
	if err != nil {
		return nil, res, err
	}

	locks := &lockList{}
	if res.StatusCode == http.StatusOK {
		err = lfshttp.DecodeJSON(res, locks)
	}

	return locks, res, err
}

// lockVerifiableRequest encapsulates the request sent to the server when the
// client would like a list of locks to verify a Git push.
type lockVerifiableRequest struct {
	Ref *lockRef `json:"ref,omitempty"`

	// Cursor is an optional field used to tell the server which lock was
	// seen last, if scanning through multiple pages of results.
	//
	// Servers must return a list of locks sorted in reverse chronological
	// order, so the Cursor provides a consistent method of viewing all
	// locks, even if more were created between two requests.
	Cursor string `json:"cursor,omitempty"`
	// Limit is the maximum number of locks to return in a single page.
	Limit int `json:"limit,omitempty"`
}

// lockVerifiableList encapsulates a set of Locks to verify a Git push.
type lockVerifiableList struct {
	// Ours is the set of locks returned back matching filenames that the user
	// is allowed to edit.
	Ours []Lock `json:"ours"`

	// Their is the set of locks returned back matching filenames that the user
	// is NOT allowed to edit. Any edits matching these files should reject
	// the Git push.
	Theirs []Lock `json:"theirs"`

	// NextCursor returns the Id of the Lock the client should update its
	// cursor to, if there are multiple pages of results for a particular
	// `LockListRequest`.
	NextCursor string `json:"next_cursor,omitempty"`
	// Message populates any error that was encountered during the search. If no
	// error was encountered and the operation was successful, then a value
	// of nil will be passed here.
	Message          string `json:"message,omitempty"`
	DocumentationURL string `json:"documentation_url,omitempty"`
	RequestID        string `json:"request_id,omitempty"`
}

func (c *lockClient) SearchVerifiable(remote string, vreq *lockVerifiableRequest) (*lockVerifiableList, *http.Response, error) {
	e := c.Endpoints.Endpoint("upload", remote)
	req, err := c.NewRequest("POST", e, "locks/verify", vreq)
	if err != nil {
		return nil, nil, err
	}

	req = c.Client.LogRequest(req, "lfs.locks.verify")
	res, err := c.DoAPIRequestWithAuth(remote, req)
	if err != nil {
		return nil, res, err
	}

	locks := &lockVerifiableList{}
	if res.StatusCode == http.StatusOK {
		err = lfshttp.DecodeJSON(res, locks)
	}

	return locks, res, err
}

// User represents the owner of a lock.
type User struct {
	// Name is the name of the individual who would like to obtain the
	// lock, for instance: "Rick Sanchez".
	Name string `json:"name"`
}

func NewUser(name string) *User {
	return &User{Name: name}
}

// String implements the fmt.Stringer interface.
func (u *User) String() string {
	return u.Name
}
