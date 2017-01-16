package lfsapi

import (
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
	requestStats  *httpTransferStats
	responseStats *httpTransferStats
}

// LogStats is intended to be called after all HTTP operations for the
// commmand have finished. It dumps k/v logs, one line per httpTransfer into
// a log file with the current timestamp.
func (c *Client) LogStats(out io.Writer) {
	if !c.LoggingStats {
		return
	}

	fmt.Fprintf(out, "concurrent=%d time=%d version=%s\n", c.ConcurrentTransfers, time.Now().Unix(), UserAgent)

	for key, responses := range c.transferBuckets {
		for _, response := range responses {
			stats := c.transfers[response]
			fmt.Fprintf(out, "key=%s reqheader=%d reqbody=%d resheader=%d resbody=%d restime=%d status=%d url=%s\n",
				key,
				stats.requestStats.HeaderSize,
				stats.requestStats.BodySize,
				stats.responseStats.HeaderSize,
				stats.responseStats.BodySize,
				stats.responseStats.Stop.Sub(stats.responseStats.Start).Nanoseconds(),
				response.StatusCode,
				response.Request.URL)
		}
	}
}

func (c *Client) LogResponse(key string, res *http.Response) {
	if !c.LoggingStats {
		return
	}

	c.transferBucketMu.Lock()
	defer c.transferBucketMu.Unlock()

	if c.transferBuckets == nil {
		c.transferBuckets = make(map[string][]*http.Response)
	}

	c.transferBuckets[key] = append(c.transferBuckets[key], res)
}

func (c *Client) startResponseStats(res *http.Response, start time.Time) {
	if !c.LoggingStats {
		return
	}

	reqHeaderSize := 0
	resHeaderSize := 0

	if dump, err := httputil.DumpRequest(res.Request, false); err == nil {
		reqHeaderSize = len(dump)
	}

	if dump, err := httputil.DumpResponse(res, false); err == nil {
		resHeaderSize = len(dump)
	}

	reqstats := &httpTransferStats{HeaderSize: reqHeaderSize, BodySize: res.Request.ContentLength}

	// Response body size cannot be figured until it is read. Do not rely on a Content-Length
	// header because it may not exist or be -1 in the case of chunked responses.
	resstats := &httpTransferStats{HeaderSize: resHeaderSize, Start: start}
	t := &httpTransfer{requestStats: reqstats, responseStats: resstats}

	c.transferMu.Lock()
	if c.transfers == nil {
		c.transfers = make(map[*http.Response]*httpTransfer)
	}
	c.transfers[res] = t
	c.transferMu.Unlock()
}

func (c *Client) finishResponseStats(res *http.Response, bodySize int64) {
	if !c.LoggingStats || res == nil {
		return
	}

	c.transferMu.Lock()
	defer c.transferMu.Unlock()

	if c.transfers == nil {
		return
	}

	if transfer, ok := c.transfers[res]; ok {
		transfer.responseStats.BodySize = bodySize
		transfer.responseStats.Stop = time.Now()
	}
}
