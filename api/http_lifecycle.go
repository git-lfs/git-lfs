// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/github/git-lfs/auth"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/httputil"
)

var (
	// ErrNoOperationGiven is an error which is returned when no operation
	// is provided in a RequestSchema object.
	ErrNoOperationGiven = errors.New("lfs/api: no operation provided in schema")
)

// EndpointSource is an interface which encapsulates the behavior of returning
// `config.Endpoint`s based on a particular operation.
type EndpointSource interface {
	// Endpoint returns the `config.Endpoint` assosciated with a given
	// operation.
	Endpoint(operation string) config.Endpoint
}

// HttpLifecycle serves as the default implementation of the Lifecycle interface
// for HTTP requests. Internally, it leverages the *http.Client type to execute
// HTTP requests against a root *url.URL, as given in `NewHttpLifecycle`.
type HttpLifecycle struct {
	endpoints EndpointSource
}

var _ Lifecycle = new(HttpLifecycle)

// NewHttpLifecycle initializes a new instance of the *HttpLifecycle type with a
// new *http.Client, and the given root (see above).
func NewHttpLifecycle(endpoints EndpointSource) *HttpLifecycle {
	return &HttpLifecycle{
		endpoints: endpoints,
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
// In all cases, credentials are attached to the HTTP request as described in
// the `auth` package (see github.com/github/git-lfs/auth#GetCreds).
//
// Finally, all of these components are combined together and the resulting
// request is returned.
func (l *HttpLifecycle) Build(schema *RequestSchema) (*http.Request, error) {
	path, err := l.absolutePath(schema.Operation, schema.Path)
	if err != nil {
		return nil, err
	}

	body, err := l.body(schema)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(schema.Method, path.String(), body)
	if err != nil {
		return nil, err
	}

	if _, err = auth.GetCreds(req); err != nil {
		return nil, err
	}

	req.URL.RawQuery = l.queryParameters(schema).Encode()

	return req, nil
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
func (l *HttpLifecycle) Execute(req *http.Request, into interface{}) (Response, error) {
	resp, err := httputil.DoHttpRequestWithRedirects(req, []*http.Request{}, true)
	if err != nil {
		return nil, err
	}

	if into != nil {
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(into); err != nil {
			return nil, err
		}
	}

	return WrapHttpResponse(resp), nil
}

// Cleanup implements the Lifecycle.Cleanup function by closing the Body
// attached to the response.
func (l *HttpLifecycle) Cleanup(resp Response) error {
	return resp.Body().Close()
}

// absolutePath returns the absolute path made by combining a given relative
// path with the root URL of the endpoint corresponding to the given operation.
//
// If there was an error in parsing the relative path, then that error will be
// returned.
func (l *HttpLifecycle) absolutePath(operation Operation, path string) (*url.URL, error) {
	if len(operation) == 0 {
		return nil, ErrNoOperationGiven
	}

	root, err := url.Parse(l.endpoints.Endpoint(string(operation)).Url)
	if err != nil {
		return nil, err
	}

	rel, err := url.Parse(path)
	if err != nil {
		return nil, err

	}

	return root.ResolveReference(rel), nil
}

// body returns an io.Reader which reads out a JSON-encoded copy of the payload
// attached to a given *RequestSchema, if it is present. If no body is present
// in the request, then nil is returned instead.
//
// If an error was encountered while attempting to marshal the body, then that
// will be returned instead, along with a nil io.Reader.
func (l *HttpLifecycle) body(schema *RequestSchema) (io.ReadCloser, error) {
	if schema.Body == nil {
		return nil, nil
	}

	body, err := json.Marshal(schema.Body)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(bytes.NewReader(body)), nil
}

// queryParameters returns a url.Values containing all of the provided query
// parameters as given in the *RequestSchema. If no query parameters were given,
// then an empty url.Values is returned instead.
func (l *HttpLifecycle) queryParameters(schema *RequestSchema) url.Values {
	vals := url.Values{}
	if schema.Query != nil {
		for k, v := range schema.Query {
			vals.Add(k, v)
		}
	}

	return vals
}
