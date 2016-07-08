// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

import "github.com/github/git-lfs/config"

type Operation string

const (
	UploadOperation   Operation = "upload"
	DownloadOperation Operation = "download"
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
	// Locks is the LockService used to interact with the Git LFS file-
	// locking API.
	Locks LockService

	// lifecycle is the lifecycle used by all requests through this client.
	lifecycle Lifecycle
}

// NewClient instantiates and returns a new instance of *Client, with the given
// lifecycle.
//
// If no lifecycle is given, a HttpLifecycle is used by default.
func NewClient(lifecycle Lifecycle) *Client {
	if lifecycle == nil {
		lifecycle = NewHttpLifecycle(config.Config)
	}

	return &Client{lifecycle: lifecycle}
}

// Do preforms the request assosicated with the given *RequestSchema by
// delegating into the Lifecycle in use.
//
// If any error was encountered while either building, executing or cleaning up
// the request, then it will be returned immediately, and the request can be
// treated as invalid.
//
// If no error occured, an api.Response will be returned, along with a `nil`
// error. At this point, the body of the response has been serialized into
// `schema.Into`, and the body has been closed.
func (c *Client) Do(schema *RequestSchema) (Response, error) {
	req, err := c.lifecycle.Build(schema)
	if err != nil {
		return nil, err
	}

	resp, err := c.lifecycle.Execute(req, schema.Into)
	if err != nil {
		return nil, err
	}

	if err = c.lifecycle.Cleanup(resp); err != nil {
		return nil, err
	}

	return resp, nil
}
