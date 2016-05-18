// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/github/git-lfs/api"
)

// HttpLifecycle serves as the default implementation of the Lifecycle interface
// for HTTP requests. Internally, it leverages the *http.Client type to execute
// HTTP requests against a root *url.URL, as given in `NewHttpLifecycle`.
type HttpLifecycle struct {
	// root is the root of the API server, from which all other sub-paths are
	// relativized
	root *url.URL
	// client is the *http.Client used to execute these requests.
	client *http.Client
}

var _ Lifecycle = new(HttpLifecycle)

// NewHttpLifecycle initializes a new instance of the *HttpLifecycle type with a
// new *http.Client, and the given root (see above).
func NewHttpLifecycle(root *url.URL) *HttpLifecycle {
	return &HttpLifecycle{
		root:   root,
		client: new(http.Client),
	}
}

// Build implements the Lifecycle.Build function.
//
// HttpLifecycle in particular, builds an absolute path by parsing and then
// relativizing the `schema.Path` with respsect to the `HttpLifecycle.root`. If
// there was an error in determining this URL, then that error will be returned,
//
// After this is complete, a body is attached to the request if the
// schema contained one. If a body was present, and there an error occurred while
// serializing it into JSON, then that error will be returned and the
// *http.Request will not be generated.
//
// Finally, all of these components are combined together and the resulting
// request is returned.
func (l *HttpLifecycle) Build(schema *api.RequestSchema) (*http.Request, error) {
	path, err := l.AbsolutePath(schema.Path)
	if err != nil {
		return nil, err
	}

	body, err := l.Body(schema)
	if err != nil {
		return nil, err
	}

	// TODO(taylor): attach creds!
	return http.NewRequest(schema.Method, path.String(), body)
}

// Execute implements the Lifecycle.Execute function.
//
// Internally, the *http.Client is used to execute the underlying *http.Request.
// If the client returned an error corresponding to a failure to make the
// request, then that error will be returned immediately, and the response is
// guaranteed not to be serialized.
//
// Once the response has been gathered from the server, it is unmarshled into
// the given `into interface{}` which is identical to the one provided in the
// original RequestSchema. If an error occured while decoding, then that error
// is returned.
//
// Otherwise, the api.Response is returned, along with no error, signaling that
// the request completed successfully.
func (l *HttpLifecycle) Execute(req *http.Request, into interface{}) (api.Response, error) {
	resp, err := l.c.Do(req)
	if err != nil {
		return nil, err
	}

	if into != nil {
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(&into); err != nil {
			return nil, err
		}
	}

	return WrapHttpResponse(resp), nil
}

// Cleanup implements the Lifecycle.Cleanup function by closing the Body
// attached to the repsonse.
func (l *HttpLifecycle) Cleanup(resp api.Response) error {
	return resp.Body().Close()
}

// AbsolutePath returns the absolute path made by combining a given relative
// path with the owned "base" path. If there was an error in parsing the
// relative path, then that error will be returned.
func (l *HttpLifecycle) AbsolutePath(path string) (*url.URL, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err

	}

	return l.root.ResolveReference(rel), nil
}

// Body returns an io.Reader which reads out a JSON-encoded copy of the payload
// attached to a given *RequestSchema, if it is present. If no body is present
// in the request, then nil is returned instead.
//
// If an error was encountered while attempting to marshal the body, then that
// will be returned instead, along with a nil io.Reader.
func (l *HttpLifecycle) Body(schema *api.RequestSchema) (io.ReadCloser, error) {
	if schema.Body == nil {
		return nil, nil
	}

	body, err := json.Marshal(schema.Body)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(bytes.NewReader(body)), nil
}
