package lfsapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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
	c.httpLogger = newSyncLogger(w)
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
	if v := res.Request.Context().Value(transferKey); v != nil {
		t := v.(httpTransfer)
		c.httpLogger.Write(fmt.Sprintf("key=%s url=%s status=%d reqbody=%d resbody=%d restime=%d\n",
			t.Key,
			strings.SplitN(res.Request.URL.String(), "?", 2)[0],
			res.StatusCode,
			t.RequestBodySize,
			bodySize,
			time.Since(t.Start).Nanoseconds(),
		))
	}
}

func newSyncLogger(w io.WriteCloser) *syncLogger {
	ch := make(chan string, 100)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func(c chan string, w io.Writer, wg *sync.WaitGroup) {
		for l := range c {
			w.Write([]byte(l))
			wg.Done()
		}
	}(ch, w, wg)

	return &syncLogger{w: w, ch: ch, wg: wg}
}

type syncLogger struct {
	w  io.WriteCloser
	ch chan string
	wg *sync.WaitGroup
}

func (l *syncLogger) Write(line string) {
	if l != nil {
		l.wg.Add(1)
		l.ch <- line
	}
}

func (l *syncLogger) Close() error {
	if l == nil {
		return nil
	}

	l.wg.Done()
	l.wg.Wait()
	return l.w.Close()
}
