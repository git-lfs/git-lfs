// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

import "io"

// Response is an interface that represents a response returned as a result of
// executing an API call. It is designed to represent itself across multiple
// response type, be it HTTP, SSH, or something else.
//
// The Response interface is meant to be small enough such that it can be
// sufficiently general, but is easily accessible if a caller needs more
// information specific to a particular protocol.
type Response interface {
	// Status is a human-readable string representing the status the
	// response was returned with.
	Status() string
	// StatusCode is the numeric code associated with a particular status.
	StatusCode() int
	// Proto is the protocol with which the response was delivered.
	Proto() string
	// Body returns an io.ReadCloser containg the contents of the response's
	// body.
	Body() io.ReadCloser
}
