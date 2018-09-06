package lfshttp

import (
	"net/http"
)

func (c *Client) doWithAuth(remote string, req *http.Request, via []*http.Request) (*http.Response, error) {
	return c.do(req, remote, via)
}
