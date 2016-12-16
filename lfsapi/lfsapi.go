package lfsapi

import "net/http"

type Client struct {
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}
