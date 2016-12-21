package lfsapi

import (
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

func (c *Client) LogResponse(key string, res *http.Response) {
	c.transferBucketMu.Lock()
	defer c.transferBucketMu.Unlock()

	if c.transferBuckets == nil {
		c.transferBuckets = make(map[string][]*http.Response)
	}

	c.transferBuckets[key] = append(c.transferBuckets[key], res)
}

func (c *Client) StartResponseStats(res *http.Response, start time.Time) {
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

func (c *Client) FinishResponseStats(res *http.Response, bodySize int64) {
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
