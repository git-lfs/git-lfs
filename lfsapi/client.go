package lfsapi

import (
	"io"
	"net/http"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/creds"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
)

func (c *Client) NewRequest(method string, e lfshttp.Endpoint, suffix string, body interface{}) (*http.Request, error) {
	return c.client.NewRequest(method, e, suffix, body)
}

// Do sends an HTTP request to get an HTTP response. It wraps net/http, adding
// extra headers, redirection handling, and error reporting.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func (c *Client) doWithAccess(req *http.Request, remote string, via []*http.Request, mode creds.AccessMode) (*http.Response, error) {
	return c.client.DoWithAccess(req, mode)
}

func (c *Client) LogRequest(r *http.Request, reqKey string) *http.Request {
	return c.client.LogRequest(r, reqKey)
}

func (c *Client) GitEnv() config.Environment {
	return c.client.GitEnv()
}

func (c *Client) OSEnv() config.Environment {
	return c.client.OSEnv()
}

func (c *Client) ConcurrentTransfers() int {
	return c.client.ConcurrentTransfers
}

func (c *Client) LogHTTPStats(w io.WriteCloser) {
	c.client.LogHTTPStats(w)
}

func (c *Client) Close() error {
	return c.client.Close()
}
