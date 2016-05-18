// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

import (
	"io"
	"net/http"
)

// HttpResponse is an implementation of the Response interface capable of
// handling HTTP responses. At its core, it works by wrapping an *http.Response.
type HttpResponse struct {
	// r is the underlying *http.Response that is being wrapped.
	r *http.Response
}

// WrapHttpResponse returns a wrapped *HttpResponse implementing the Repsonse
// type by using the given *http.Response.
func WrapHttpResponse(r *http.Response) *HttpResponse {
	return &HttpResponse{
		r: r,
	}
}

var _ Response = new(HttpResponse)

// Status implements the Response.Status function, and returns the status given
// by the underlying *http.Response.
func (h *HttpResponse) Status() string {
	return h.r.Status
}

// StatusCode implements the Response.StatusCode function, and returns the
// status code given by the underlying *http.Response.
func (h *HttpResponse) StatusCode() int {
	return h.r.StatusCode
}

// Proto implements the Response.Proto function, and returns the proto given by
// the underlying *http.Response.
func (h *HttpResponse) Proto() string {
	return h.r.Proto
}

// Body implements the Response.Body function, and returns the body as given by
// the underlying *http.Response.
func (h *HttpResponse) Body() io.ReadCloser {
	return h.r.Body
}

// Header returns the underlying *http.Response's header.
func (h *HttpResponse) Header() http.Header {
	return h.r.Header
}
