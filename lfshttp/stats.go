package lfshttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/git-lfs/git-lfs/tools"
)

type httpTransfer struct {
	// members managed via sync/atomic must be aligned at the top of this
	// structure (see: https://github.com/git-lfs/git-lfs/pull/2880).

	RequestBodySize int64
	Start           int64
	ResponseStart   int64
	ConnStart       int64
	ConnEnd         int64
	DNSStart        int64
	DNSEnd          int64
	TLSStart        int64
	TLSEnd          int64
	URL             string
	Method          string
	Key             string
}

type statsContextKey string

const transferKey = statsContextKey("transfer")

func (c *Client) LogHTTPStats(w io.WriteCloser) {
	fmt.Fprintf(w, "concurrent=%d time=%d version=%s\n", c.ConcurrentTransfers, time.Now().Unix(), UserAgent)
	c.httpLogger = newSyncLogger(w)
}

// LogStats is intended to be called after all HTTP operations for the
// command have finished. It dumps k/v logs, one line per httpTransfer into
// a log file with the current timestamp.
//
// DEPRECATED: Call LogHTTPStats() before the first HTTP request.
func (c *Client) LogStats(out io.Writer) {}

// LogRequest tells the client to log the request's stats to the http log
// after the response body has been read.
func (c *Client) LogRequest(r *http.Request, reqKey string) *http.Request {
	if c.httpLogger == nil {
		return r
	}

	t := &httpTransfer{
		URL:    strings.SplitN(r.URL.String(), "?", 2)[0],
		Method: r.Method,
		Key:    reqKey,
	}

	ctx := httptrace.WithClientTrace(r.Context(), &httptrace.ClientTrace{
		GetConn: func(_ string) {
			atomic.CompareAndSwapInt64(&t.Start, 0, time.Now().UnixNano())
		},
		DNSStart: func(_ httptrace.DNSStartInfo) {
			atomic.CompareAndSwapInt64(&t.DNSStart, 0, time.Now().UnixNano())
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			atomic.CompareAndSwapInt64(&t.DNSEnd, 0, time.Now().UnixNano())
		},
		ConnectStart: func(_, _ string) {
			atomic.CompareAndSwapInt64(&t.ConnStart, 0, time.Now().UnixNano())
		},
		ConnectDone: func(_, _ string, _ error) {
			atomic.CompareAndSwapInt64(&t.ConnEnd, 0, time.Now().UnixNano())
		},
		TLSHandshakeStart: func() {
			atomic.CompareAndSwapInt64(&t.TLSStart, 0, time.Now().UnixNano())
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			atomic.CompareAndSwapInt64(&t.TLSEnd, 0, time.Now().UnixNano())
		},
		GotFirstResponseByte: func() {
			atomic.CompareAndSwapInt64(&t.ResponseStart, 0, time.Now().UnixNano())
		},
	})

	return r.WithContext(context.WithValue(ctx, transferKey, t))
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

func (l *syncLogger) LogRequest(req *http.Request, bodySize int64) {
	if l == nil {
		return
	}

	if v := req.Context().Value(transferKey); v != nil {
		l.logTransfer(v.(*httpTransfer), "request", fmt.Sprintf(" body=%d", bodySize))
	}
}

func (l *syncLogger) LogResponse(req *http.Request, status int, bodySize int64) {
	if l == nil {
		return
	}

	if v := req.Context().Value(transferKey); v != nil {
		t := v.(*httpTransfer)
		now := time.Now().UnixNano()
		l.logTransfer(t, "response",
			fmt.Sprintf(" status=%d body=%d conntime=%d dnstime=%d tlstime=%d restime=%d time=%d",
				status,
				bodySize,
				tools.MaxInt64(t.ConnEnd-t.ConnStart, 0),
				tools.MaxInt64(t.DNSEnd-t.DNSStart, 0),
				tools.MaxInt64(t.TLSEnd-t.TLSStart, 0),
				tools.MaxInt64(now-t.ResponseStart, 0),
				tools.MaxInt64(now-t.Start, 0),
			))
	}
}

func (l *syncLogger) logTransfer(t *httpTransfer, event, extra string) {
	l.wg.Add(1)
	l.ch <- fmt.Sprintf("key=%s event=%s url=%s method=%s%s\n",
		t.Key,
		event,
		t.URL,
		t.Method,
		extra,
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
