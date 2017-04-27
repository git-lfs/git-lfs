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
	URL             string
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
	ctx := context.WithValue(r.Context(), transferKey, httpTransfer{
		URL: strings.SplitN(r.URL.String(), "?", 2)[0],
		Key: reqKey,
	})
	return r.WithContext(ctx)
}

// LogResponse sends the current response stats to the http log.
//
// DEPRECATED: Use LogRequest() instead.
func (c *Client) LogResponse(key string, res *http.Response) {}

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

func (l *syncLogger) Log(req *http.Request, event, extra string) {
	if l == nil {
		return
	}

	v := req.Context().Value(transferKey)
	if v == nil {
		return
	}

	t := v.(httpTransfer)
	l.wg.Add(1)
	l.ch <- fmt.Sprintf("key=%s event=%s url=%s method=%s %ssince=%d\n",
		t.Key,
		event,
		t.URL,
		req.Method,
		extra,
		time.Since(t.Start).Nanoseconds(),
	)
}

func (l *syncLogger) Close() error {
	if l == nil {
		return nil
	}

	l.wg.Done()
	l.wg.Wait()
	return l.w.Close()
}
