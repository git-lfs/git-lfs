package lfsapi

import (
	"context"
	"net/http"
)

// ckey is a type that wraps a string for package-unique context.Context keys.
type ckey string

const (
	// contextKeyRetries is a context.Context key for storing the desired
	// number of retries for a given request.
	contextKeyRetries ckey = "retries"

	// defaultRequestRetries is the default number of retries to perform on
	// a given HTTP request.
	defaultRequestRetries = 0
)

// WithRetries stores the desired number of retries "n" on the given
// http.Request, and causes it to be retried "n" times in the case of a non-nil
// network related error.
func WithRetries(req *http.Request, n int) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, contextKeyRetries, n)

	return req.WithContext(ctx)
}

// Retries returns the number of retries requested for a given http.Request.
func Retries(req *http.Request) (int, bool) {
	n, ok := req.Context().Value(contextKeyRetries).(int)

	return n, ok
}
