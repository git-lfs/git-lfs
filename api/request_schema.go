// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

// RequestSchema provides a schema from which to generate sendable requests.
type RequestSchema struct {
	// Method is the method that should be used when making a particular API
	// call.
	Method string
	// Path is the relative path that this API call should be made against.
	Path string
	// Operation is the operation used to determine which endpoint to make
	// the request against (see github.com/github/git-lfs/config).
	Operation Operation
	// Query is the query parameters used in the request URI.
	Query map[string]string
	// Body is the body of the request.
	Body interface{}
	// Into is an optional field used to represent the data structure into
	// which a response should be serialized.
	Into interface{}
}
