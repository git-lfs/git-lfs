package api

import (
	"fmt"
	"strconv"
	"time"
)

// LockService is an API service which encapsulates the Git LFS Locking API.
type LockService struct{}

// Lock generates a *RequestSchema that is used to preform the "attempt lock"
// API method.
//
// If a lock is already present, or if the server was unable to generate the
// lock, the Err field of the LockResponse type will be populated with a more
// detailed error describing the situation.
//
// If the caller does not have the minimum commit necessary to obtain the lock
// on that file, then the CommitNeeded field will be populated in the
// LockResponse, signaling that more commits are needed.
//
// In the successful case, a new Lock will be returned and granted to the
// caller.
func (s *LockService) Lock(req *LockRequest) (*RequestSchema, *LockResponse) {
	var resp LockResponse

	return &RequestSchema{
		Method:    "POST",
		Path:      "/locks",
		Operation: UploadOperation,
		Body:      req,
		Into:      &resp,
	}, &resp
}

// Search generates a *RequestSchema that is used to preform the "search for
// locks" API method.
//
// Searches can be scoped to match specific parameters by using the Filters
// field in the given LockSearchRequest. If no matching Locks were found, then
// the Locks field of the response will be empty.
//
// If the client expects that the server will return many locks, then the client
// can choose to paginate that response. Pagination is preformed by limiting the
// amount of results per page, and the server will inform the client of the ID
// of the last returned lock. Since the server is guaranteed to return results
// in reverse chronological order, the client simply sends the last ID it
// processed along with the next request, and the server will continue where it
// left off.
//
// If the server was unable to process the lock search request, then the Error
// field will be populated in the response.
//
// In the successful case, one or more locks will be returned as a part of the
// response.
func (s *LockService) Search(req *LockSearchRequest) (*RequestSchema, *LockList) {
	var resp LockList

	query := make(map[string]string)
	for _, filter := range req.Filters {
		query[filter.Property] = filter.Value
	}

	if req.Cursor != "" {
		query["cursor"] = req.Cursor
	}

	if req.Limit != 0 {
		query["limit"] = strconv.Itoa(req.Limit)
	}

	return &RequestSchema{
		Method:    "GET",
		Path:      "/locks",
		Operation: UploadOperation,
		Query:     query,
		Into:      &resp,
	}, &resp
}

// Unlock generates a *RequestSchema that is used to preform the "unlock" API
// method, against a particular lock potentially with --force.
//
// This method's corresponding response type will either contain a reference to
// the lock that was unlocked, or an error that was experienced by the server in
// unlocking it.
func (s *LockService) Unlock(id string, force bool) (*RequestSchema, *UnlockResponse) {
	var resp UnlockResponse

	return &RequestSchema{
		Method:    "POST",
		Path:      fmt.Sprintf("/locks/%s/unlock", id),
		Operation: UploadOperation,
		Body:      &UnlockRequest{id, force},
		Into:      &resp,
	}, &resp
}

// Lock represents a single lock that against a particular path.
//
// Locks returned from the API may or may not be currently active, according to
// the Expired flag.
type Lock struct {
	// Id is the unique identifier corresponding to this particular Lock. It
	// must be consistent with the local copy, and the server's copy.
	Id string `json:"id"`
	// Path is an absolute path to the file that is locked as a part of this
	// lock.
	Path string `json:"path"`
	// Committer is the author who initiated this lock.
	Committer Committer `json:"committer"`
	// CommitSHA is the commit that this Lock was created against. It is
	// strictly equal to the SHA of the minimum commit negotiated in order
	// to create this lock.
	CommitSHA string `json:"commit_sha"`
	// LockedAt is a required parameter that represents the instant in time
	// that this lock was created. For most server implementations, this
	// should be set to the instant at which the lock was initially
	// received.
	LockedAt time.Time `json:"locked_at"`
	// ExpiresAt is an optional parameter that represents the instant in
	// time that the lock stopped being active. If the lock is still active,
	// the server can either a) not send this field, or b) send the
	// zero-value of time.Time.
	UnlockedAt time.Time `json:"unlocked_at,omitempty"`
}

// Active returns whether or not the given lock is still active against the file
// that it is protecting.
func (l *Lock) Active() bool {
	return l.UnlockedAt.IsZero()
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

func NewCommitter(name, email string) Committer {
	return Committer{Name: name, Email: email}
}

// LockRequest encapsulates the payload sent across the API when a client would
// like to obtain a lock against a particular path on a given remote.
type LockRequest struct {
	// Path is the path that the client would like to obtain a lock against.
	Path string `json:"path"`
	// LatestRemoteCommit is the SHA of the last known commit from the
	// remote that we are trying to create the lock against, as found in
	// `.git/refs/origin/<name>`.
	LatestRemoteCommit string `json:"latest_remote_commit"`
	// Committer is the individual that wishes to obtain the lock.
	Committer Committer `json:"committer"`
}

// LockResponse encapsulates the information sent over the API in response to
// a `LockRequest`.
type LockResponse struct {
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

// UnlockRequest encapsulates the data sent in an API request to remove a lock.
type UnlockRequest struct {
	// Id is the Id of the lock that the user wishes to unlock.
	Id string `json:"id"`
	// Force determines whether or not the lock should be "forcibly"
	// unlocked; that is to say whether or not a given individual should be
	// able to break a different individual's lock.
	Force bool `json:"force"`
}

// UnlockResponse is the result sent back from the API when asked to remove a
// lock.
type UnlockResponse struct {
	// Lock is the lock corresponding to the asked-about lock in the
	// `UnlockPayload` (see above). If no matching lock was found, this
	// field will take the zero-value of Lock, and Err will be non-nil.
	Lock *Lock `json:"lock"`
	// Err is an optional field which holds any error that was experienced
	// while removing the lock.
	Err string `json:"error,omitempty"`
}

// Filter represents a single qualifier to apply against a set of locks.
type Filter struct {
	// Property is the property to search against.
	// Value is the value that the property must take.
	Property, Value string
}

// LockSearchRequest encapsulates the request sent to the server when the client
// would like a list of locks that match the given criteria.
type LockSearchRequest struct {
	// Filters is the set of filters to query against. If the client wishes
	// to obtain a list of all locks, an empty array should be passed here.
	Filters []Filter
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

// LockList encapsulates a set of Locks.
type LockList struct {
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
