// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

import (
	"net/url"

	"github.com/github/git-lfs/api"
)

const (
	MediaType = "application/vnd.git-lfs+json; charset=utf-8"
)

// Client exposes the LFS API to callers through a multitude of different
// services and transport mechanisms. Callers can make a *RequestSchema using
// any service that is attached to the Client, and then execute a request based
// on that schema using the `Do()` method.
//
// A prototypical example follows:
// ```
//   apiResponse, schema := client.Locks.Lock(request)
//   resp, err := client.Do(schema)
//   if err != nil {
//       handleErr(err)
//   }
//
//   fmt.Println(apiResponse.Lock)
// ```
type Client struct {
	// base is root URL that all requests will be made against. It is
	// initialized when the client is constructed, and remains immutable
	// throughout the duration of the *Client.
	base *url.URL
	// http is the lifecycle used by all requests through this client.
	http *HttpLifecycle
}

// NewClient instantiates and returns a new instance of *Client with a base path
// initialized to the given `root`. If `root` is unable to be parsed according
// to the rules of `url.Parse`, then a `nil` client will be returned, and the
// parse error will be returned instead.
//
// Assuming all goes well, a *Client is returned as expected, along with a `nil`
// error.
func NewClient(root string) (*Client, error) {
	base, err := url.Parse(root)
	if err != nil {
		return nil, err
	}

	return &Client{
		base: base,
		http: NewHttpLifecycle(base),
	}, nil
}

// Do preforms the request assosicated with the given *RequestSchema by
// delegating into the Lifecycle in use.
//
// If any error was encountered while either building, executing or cleaning up
// the request, then it will be returned immediately, and the request can be
// treated as invalid.
//
// If no error occured, an some api.Response implementation will be returned,
// along with a `nil` error. At this point, the body of the response has been
// serialized into `schema.Into`, and the body is closed.
func (c *Client) Do(schema *api.RequestSchema) (api.Response, error) {
	req, err := c.http.Build(schema)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Execute(req, schema.Into)
	if err != nil {
		return nil, err
	}

	return c.http.Cleanup(resp)
}
