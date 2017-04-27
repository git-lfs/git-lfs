package lfsapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"
)

type httpTransferStats struct {
	HeaderSize int
	BodySize   int64
	Start      time.Time
	Stop       time.Time
}

type httpTransfer struct {
	Key           string
	RequestStats  httpTransferStats
	ResponseStats httpTransferStats
}

type statsContextKey string

const statsKeytransferKey = statsContextKey("transfer")

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
	ctx := context.WithValue(r.Context(), transferKey, httpTransfer{
		Key:           reqKey,
		RequestStats:  httpTransferStats{},
		ResponseStats: httpTransferStats{},
	})
	return r.WithContext(ctx)
}

// LogResponse sends the current response stats to the http log.
//
// DEPRECATED: Use LogRequest() instead.
func (c *Client) LogResponse(key string, res *http.Response) {}

func (c *Client) startResponseStats(res *http.Response, start time.Time) {
	v := res.Request.Context().Value(transferKey)
	if v == nil {
		return
	}

	t := v.(httpTransfer)

	t.RequestStats.BodySize = res.Request.ContentLength
	if dump, err := httputil.DumpRequest(res.Request, false); err == nil {
		t.RequestStats.HeaderSize = len(dump)
	}

	// Response body size cannot be figured until it is read. Do not rely on a Content-Length
	// header because it may not exist or be -1 in the case of chunked responses.
	t.ResponseStats.Start = start
	if dump, err := httputil.DumpResponse(res, false); err == nil {
		t.ResponseStats.HeaderSize = len(dump)
	}
}

func (c *Client) finishResponseStats(res *http.Response, bodySize int64) {
	v := res.Request.Context().Value(transferKey)
	if v == nil {
		return
	}

	t := v.(httpTransfer)
	t.ResponseStats.BodySize = bodySize
	t.ResponseStats.Stop = time.Now()
	if c.httpLogger != nil {
		writeHTTPStats(c.httpLogger, res, t)
	}
}

func writeHTTPStats(w io.Writer, res *http.Response, t httpTransfer) {
	fmt.Fprintf(w, "key=%s reqheader=%d reqbody=%d resheader=%d resbody=%d restime=%d status=%d url=%s\n",
		t.Key,
		t.RequestStats.HeaderSize,
		t.RequestStats.BodySize,
		t.ResponseStats.HeaderSize,
		t.ResponseStats.BodySize,
		t.ResponseStats.Stop.Sub(t.ResponseStats.Start).Nanoseconds(),
		res.StatusCode,
		res.Request.URL,
	)
}
