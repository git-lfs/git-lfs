// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

import "net/http"

// TODO: extract interface for *http.Request; update methods. This will be in a
// later iteration of the API client.

// A Lifecycle represents and encapsulates the behavior on an API request from
// inception to cleanup.
//
// At a high level, it turns an *api.RequestSchema into an
// api.Response (and optionally an error). Lifecycle does so by providing
// several more fine-grained methods that are used by the client to manage the
// lifecycle of a request in a platform-agnostic fashion.
type Lifecycle interface {
	// Build creates a sendable request by using the given RequestSchema as
	// a model.
	Build(req *RequestSchema) (*http.Request, error)

	// Execute transforms generated request into a wrapped repsonse, (and
	// optionally an error, if the request failed), and serializes the
	// response into the `into interface{}`, if one was provided.
	Execute(req *http.Request, into interface{}) (Response, error)

	// Cleanup is called after the request has been completed and its
	// response has been processed. It is meant to preform any post-request
	// actions necessary, like closing or resetting the connection. If an
	// error was encountered in doing this operation, it should be returned
	// from this method, otherwise nil.
	Cleanup(resp Response) error
}
