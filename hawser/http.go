package hawser

import (
	"crypto/tls"
	"fmt"
	"github.com/rubyist/tracerx"
	"io"
	"net/http"
	"os"
	"strings"
)

func DoHTTP(c *Configuration, req *http.Request) (*http.Response, error) {
	var res *http.Response
	var err error

	traceHttpRequest(c, req)

	switch req.Method {
	case "GET", "HEAD":
		res, err = c.RedirectingHttpClient().Do(req)
	default:
		res, err = c.HttpClient().Do(req)
	}

	traceHttpResponse(c, res)

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

var tracedTypes = []string{"json", "text", "xml", "html"}

func traceHttpRequest(c *Configuration, req *http.Request) {
	tracerx.Printf("HTTP: %s %s", req.Method, req.URL.String())

	if c.isTracingHttp == false {
		return
	}

	fmt.Println(">", req.Method, req.URL.RequestURI(), req.Proto)
	for key, _ := range req.Header {
		fmt.Printf("> %s: %s\n", key, req.Header.Get(key))
	}
}

func traceHttpResponse(c *Configuration, res *http.Response) {
	tracerx.Printf("HTTP: %d", res.StatusCode)

	if c.isTracingHttp == false {
		return
	}

	fmt.Println("<", res.Proto, res.Status)
	for key, _ := range res.Header {
		fmt.Printf("< %s: %s\n", key, res.Header.Get(key))
	}

	traceBody := false
	ctype := strings.ToLower(strings.SplitN(res.Header.Get("Content-Type"), ";", 2)[0])
	for _, tracedType := range tracedTypes {
		if strings.Contains(ctype, tracedType) {
			traceBody = true
		}
	}

	if traceBody {
		fmt.Println()
		res.Body = newTracedBody(res.Body)
	}
}

type tracedBody struct {
	body io.ReadCloser
}

func (r *tracedBody) Read(p []byte) (int, error) {
	n, err := r.body.Read(p)
	fmt.Println(string(p[0:n]))
	return n, err
}

func (r *tracedBody) Close() error {
	return r.body.Close()
}

func newTracedBody(body io.ReadCloser) *tracedBody {
	return &tracedBody{body}
}
