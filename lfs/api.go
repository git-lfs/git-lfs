package lfs

import (
	"fmt"
	"io"
)

// An abstract interface providing state & resource management for a specific Endpoint across
// potentially multiple requests
type ApiContext interface {
	// Get the endpoint this context was constructed from
	Endpoint() Endpoint
	// Close the context & any resources it's using
	Close() error

	// Download a single object, return reader for data, size and any error
	Download(oid string) (io.ReadCloser, int64, *WrappedError)
	// Upload a single object
	Upload(oid string, sz int64, content io.Reader, cb CopyCallback) *WrappedError
}

var (
	contextCache map[string]ApiContext
)

// Return an API context appropriate for a given Endpoint
// This may return a new context, or an existing one which is compatible with the endpoint
func GetApiContext(endpoint Endpoint) ApiContext {
	// construct a string identifier for the Endpoint
	isSSH := false
	var id string
	if len(endpoint.SshUserAndHost) > 0 {
		isSSH = true
		// SSH will use a unique connection per path as well as user/host (passed as param)
		id = fmt.Sprintf("%s:%s", endpoint.SshUserAndHost, endpoint.SshPath)
	} else {
		// We'll use the same HTTP context for all
		id = "HTTP"
	}
	ctx, ok := contextCache[id]
	if !ok {
		// Construct new
		if isSSH {
			ctx = NewSshApiContext(endpoint)
		}
		// If not SSH, OR if full SSH server isn't supported, use HTTPS with SSH auth only
		if ctx == nil {
			ctx = NewHttpApiContext(endpoint)
		}
	}

	return ctx
}
