package lfsapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type httpTransfer struct {
	Key             string
	RequestBodySize int64
	Start           time.Time
}

type statsContextKey string

const transferKey = statsContextKey("transfer")

func (c *Client) LogHTTPStats(w io.WriteCloser) {
	fmt.Fprintf(w, "concurrent=%d time=%d version=%s\n", c.ConcurrentTransfers, time.Now().Unix(), UserAgent)
	c.httpLogger = w
}

// LogStats is intended to be called after all HTTP operations for the
// commmand have finished. It dumps k/v logs, one line per httpTransfer into
// a log file with the current timestamp.
//
// DEPRECATED: Call LogHTTPStats() before the first HTTP request.
func (c *Client) LogStats(out io.Writer) {}

// LogRequest tells the client to log the request's stats to the http log
// after the response body has been read.
func (c *Client) LogRequest(r *http.Request, reqKey string) *http.Request {
	ctx := context.WithValue(r.Context(), transferKey, httpTransfer{Key: reqKey})
	return r.WithContext(ctx)
}

// LogResponse sends the current response stats to the http log.
//
// DEPRECATED: Use LogRequest() instead.
func (c *Client) LogResponse(key string, res *http.Response) {}

func (c *Client) startResponseStats(res *http.Response, start time.Time) {
	if v := res.Request.Context().Value(transferKey); v != nil {
		t := v.(httpTransfer)
		t.Start = start
		t.RequestBodySize = res.Request.ContentLength
	}
}

func (c *Client) finishResponseStats(res *http.Response, bodySize int64) {
	if c.httpLogger == nil {
		return
	}

	if v := res.Request.Context().Value(transferKey); v != nil {
		writeHTTPStats(c.httpLogger, res, v.(httpTransfer), bodySize, time.Now())
	}
}

func writeHTTPStats(w io.Writer, res *http.Response, tr httpTransfer, bodySize int64, t time.Time) {
	fmt.Fprintf(w, "key=%s url=%s status=%d reqbody=%d resbody=%d restime=%d\n",
		tr.Key,
		strings.SplitN(res.Request.URL.String(), "?", 2)[0],
		res.StatusCode,
		tr.RequestBodySize,
		bodySize,
		t.Sub(tr.Start).Nanoseconds(),
	)
}
