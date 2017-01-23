package locking

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/git-lfs/git-lfs/lfsapi"
)

type lockClient struct {
	*lfsapi.Client
}

// LockRequest encapsulates the payload sent across the API when a client would
// like to obtain a lock against a particular path on a given remote.
type lockRequest struct {
	// Path is the path that the client would like to obtain a lock against.
	Path string `json:"path"`
	// LatestRemoteCommit is the SHA of the last known commit from the
	// remote that we are trying to create the lock against, as found in
	// `.git/refs/origin/<name>`.
	LatestRemoteCommit string `json:"latest_remote_commit"`
	// Committer is the individual that wishes to obtain the lock.
	Committer *Committer `json:"committer"`
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
	// CommitNeeded holds the minimum commit SHA that client must have to
	// obtain the lock.
	CommitNeeded string `json:"commit_needed,omitempty"`
	// Err is the optional error that was encountered while trying to create
	// the above lock.
	Err string `json:"error,omitempty"`
}

func (c *lockClient) Lock(remote string, lockReq *lockRequest) (*lockResponse, *http.Response, error) {
	e := c.Endpoints.Endpoint("upload", remote)
	req, err := c.NewRequest("POST", e, "locks", lockReq)
	if err != nil {
		return nil, nil, err
	}

	res, err := c.DoWithAuth(remote, req)
	if err != nil {
		return nil, res, err
	}

	lockRes := &lockResponse{}
	return lockRes, res, lfsapi.DecodeJSON(res, lockRes)
}

// UnlockRequest encapsulates the data sent in an API request to remove a lock.
type unlockRequest struct {
	// Id is the Id of the lock that the user wishes to unlock.
	Id string `json:"id"`

	// Force determines whether or not the lock should be "forcibly"
	// unlocked; that is to say whether or not a given individual should be
	// able to break a different individual's lock.
	Force bool `json:"force"`
}

// UnlockResponse is the result sent back from the API when asked to remove a
// lock.
type unlockResponse struct {
	// Lock is the lock corresponding to the asked-about lock in the
	// `UnlockPayload` (see above). If no matching lock was found, this
	// field will take the zero-value of Lock, and Err will be non-nil.
	Lock *Lock `json:"lock"`
	// Err is an optional field which holds any error that was experienced
	// while removing the lock.
	Err string `json:"error,omitempty"`
}

func (c *lockClient) Unlock(remote, id string, force bool) (*unlockResponse, *http.Response, error) {
	e := c.Endpoints.Endpoint("upload", remote)
	suffix := fmt.Sprintf("locks/%s/unlock", id)
	req, err := c.NewRequest("POST", e, suffix, &unlockRequest{Id: id, Force: force})
	if err != nil {
		return nil, nil, err
	}

	res, err := c.DoWithAuth(remote, req)
	if err != nil {
		return nil, res, err
	}

	unlockRes := &unlockResponse{}
	err = lfsapi.DecodeJSON(res, unlockRes)
	return unlockRes, res, err
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
	// Err populates any error that was encountered during the search. If no
	// error was encountered and the operation was succesful, then a value
	// of nil will be passed here.
	Err string `json:"error,omitempty"`
}

func (c *lockClient) Search(remote string, searchReq *lockSearchRequest) (*lockList, *http.Response, error) {
	e := c.Endpoints.Endpoint("upload", remote)
	req, err := c.NewRequest("GET", e, "locks", nil)
	if err != nil {
		return nil, nil, err
	}

	q := req.URL.Query()
	for key, value := range searchReq.QueryValues() {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	res, err := c.DoWithAuth(remote, req)
	if err != nil {
		return nil, res, err
	}

	locks := &lockList{}
	if res.StatusCode == http.StatusOK {
		err = lfsapi.DecodeJSON(res, locks)
	}

	return locks, res, err
}

// Committer represents a "First Last <email@domain.com>" pair.
type Committer struct {
	// Name is the name of the individual who would like to obtain the
	// lock, for instance: "Rick Olson".
	Name string `json:"name"`
	// Email is the email assopsicated with the individual who would
	// like to obtain the lock, for instance: "rick@github.com".
	Email string `json:"email"`
}

func NewCommitter(name, email string) *Committer {
	return &Committer{Name: name, Email: email}
}

// String implements the fmt.Stringer interface by returning a string
// representation of the Committer in the format "First Last <email>".
func (c *Committer) String() string {
	return fmt.Sprintf("%s <%s>", c.Name, c.Email)
}
