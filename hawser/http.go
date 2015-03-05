package hawser

import (
	"crypto/tls"
	"github.com/rubyist/tracerx"
	"net/http"
	"os"
)

func DoHTTP(c *Configuration, req *http.Request) (*http.Response, error) {
	var res *http.Response
	var err error

	tracerx.Printf("HTTP: %s %s", req.Method, req.URL.String())

	switch req.Method {
	case "GET", "HEAD":
		res, err = c.RedirectingHttpClient().Do(req)
	default:
		res, err = c.HttpClient().Do(req)
	}

	tracerx.Printf("HTTP: %d", res.StatusCode)

	return res, err
}

func (c *Configuration) HttpClient() *http.Client {
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Transport: c.RedirectingHttpClient().Transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return RedirectError
			},
		}
	}
	return c.httpClient
}

func (c *Configuration) RedirectingHttpClient() *http.Client {
	if c.redirectingHttpClient == nil {
		tr := &http.Transport{}
		sslVerify, _ := c.GitConfig("http.sslverify")
		if sslVerify == "false" || len(os.Getenv("GIT_SSL_NO_VERIFY")) > 0 {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		c.redirectingHttpClient = &http.Client{Transport: tr}
	}
	return c.redirectingHttpClient
}
