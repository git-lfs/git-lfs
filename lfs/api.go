package lfs

import (
	"io"
	"sync"
)

// An abstract interface providing state & resource management for a specific Endpoint across
// potentially multiple requests
type ApiContext interface {
	// Get the endpoint this context was constructed from
	Endpoint() Endpoint
	// Close the context & any resources it's using
	Close() error

	// Download a single object, return reader for data, size and any error
	// Essentially the same as calling DownloadCheck() then DownloadObject()
	Download(oid string) (io.ReadCloser, int64, *WrappedError)
	// Check whether an object is available for download and return an object resource if so
	DownloadCheck(oid string) (*ObjectResource, *WrappedError)
	// Download a single object from an already identified resource from DownloadCheck(), return reader for data, size and any error
	DownloadObject(obj *ObjectResource) (io.ReadCloser, int64, *WrappedError)
	// Check whether an upload would be accepted for an object and return the resource to use if so
	UploadCheck(oid string, sz int64) (*ObjectResource, *WrappedError)
	// Perform the actual upload of an object having identified it will be accepted and the resource to use
	UploadObject(o *ObjectResource, reader io.Reader) *WrappedError
	// Perform a batch request for a number of objects to determine what can be uploaded/downloaded
	Batch(objects []*ObjectResource) ([]*ObjectResource, *WrappedError)

	// TODO - add download/upload resume
	// TODO - add binary delta support
}

var (
	// Cache can contain many contexts for the same ID / connection, for concurrent transfers
	contextCache     []ApiContext
	contextCacheLock sync.Mutex
)

// Return an API context appropriate for a given Endpoint
// Once this context is returned it is made *unavailable* to subsequent callers,
// until ReleaseApiContext is called. This is necessary to ensure that contexts
// which maintain state are only available to be used by one goroutine at a time.
// If multiple goroutines request a context for the same endpoint at once, they
// will receive separate instances which implies separate connections for stateful contexts.
func GetApiContext(endpoint Endpoint) ApiContext {
	// construct a string identifier for the Endpoint
	isSSH := len(endpoint.SshUserAndHost) > 0
	contextCacheLock.Lock()
	defer contextCacheLock.Unlock()
	var ctx ApiContext
	for i, c := range contextCache {
		cendpoint := c.Endpoint()

		if cendpoint.SshPath == endpoint.SshPath &&
			cendpoint.SshPort == endpoint.SshPort &&
			cendpoint.SshUserAndHost == endpoint.SshUserAndHost &&
			cendpoint.Url == endpoint.Url {
			ctx = c
			// remove this item
			contextCache = append(contextCache[:i], contextCache[i+1:]...)
			break
		}
	}
	if ctx == nil {
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

// Release an API context for use by other callers later. You should call this
// sometime after GetApiContext once you are done with the context. It allows
// stateful contexts to re-use resources such as connections between subsequent
// operations.
func ReleaseApiContext(ctx ApiContext) {
	contextCacheLock.Lock()
	defer contextCacheLock.Unlock()

	contextCache = append(contextCache, ctx)
}

// Shut down any open API contexts
func ShutdownApiContexts() {
	contextCacheLock.Lock()
	defer contextCacheLock.Unlock()
	for _, ctx := range contextCache {
		ctx.Close()
	}
	contextCache = nil
}
