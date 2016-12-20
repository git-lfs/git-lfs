package lfsapi

import (
	"net/http"
	"net/url"
)

func (c *Client) doWithNTLM(req *http.Request, creds Creds, credsURL *url.URL) (*http.Response, error) {
	return c.Do(req)
}
