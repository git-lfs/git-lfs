package lfs

import (
	"fmt"
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

// Holds limited number of stateful ApiContext instances & parcels them out via queues
type StatefulApiContextHolder struct {
	endpoint    Endpoint
	contextChan chan ApiContext
}

var (
	// Cache of stateful (e.g. SSH) contexts so we limit parallel access to them
	statefulContextCache = make(map[string]*StatefulApiContextHolder)
	// stateless contexts we can re-use whenever (items are not checked out)
	statelessContextCache = make(map[string]ApiContext)
	contextCacheLock      sync.Mutex
)

func getContextKey(endpoint Endpoint) (key string, isSSH bool) {
	if len(endpoint.SshUserAndHost) > 0 {
		return fmt.Sprintf("%v:%v:%v", endpoint.SshUserAndHost, endpoint.SshPort, endpoint.SshPath), true
	} else {
		return endpoint.Url, false
	}
}

// Return an API context appropriate for a given Endpoint
// Once this context is returned it may be made *unavailable* to subsequent callers,
// until ReleaseApiContext is called. This is to ensure that contexts
// which maintain state are only available to be used by one goroutine at a time.
func GetApiContext(endpoint Endpoint) ApiContext {
	// construct a string identifier for the Endpoint
	key, isSSH := getContextKey(endpoint)

	var ctx ApiContext
	contextCacheLock.Lock()
	if isSSH {
		hld, ok := statefulContextCache[key]
		if !ok {
			hld = &StatefulApiContextHolder{endpoint, make(chan ApiContext, Config.ConcurrentTransfers())}
			// Immediately add number of connections equal to the concurrent transfers
			for i := 0; i < Config.ConcurrentTransfers(); i++ {
				hld.contextChan <- NewSshApiContext(endpoint)
			}
			statefulContextCache[key] = hld
		}
		// Need to manually unlock in this path because channel access might block
		contextCacheLock.Unlock()
		ctx = <-hld.contextChan

	} else {
		defer contextCacheLock.Unlock()
		var ok bool
		ctx, ok = statelessContextCache[key]
		if !ok {
			ctx = NewHttpApiContext(endpoint)
			// Stateless get added immediately & may be used in parallel
			statelessContextCache[key] = ctx
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

	key, isSSH := getContextKey(ctx.Endpoint())
	if isSSH {
		hld, ok := statefulContextCache[key]
		if ok {
			hld.contextChan <- ctx
		}
	}
	// Do nothing for HTTP/stateless as they are not checked out exclusively
}

// Shut down any open API contexts
func ShutdownApiContexts() {
	contextCacheLock.Lock()
	defer contextCacheLock.Unlock()
	for _, hld := range statefulContextCache {
		for i := 0; i < Config.ConcurrentTransfers(); i++ {
			ctx := <-hld.contextChan
			ctx.Close()
		}
	}
	statefulContextCache = make(map[string]*StatefulApiContextHolder)
}
