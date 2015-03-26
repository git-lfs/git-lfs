package lfs

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/rubyist/tracerx"
	"io"
	"net/http"
	"os"
	"strings"
)

func DoHTTP(c *Configuration, req *http.Request) (*http.Response, error) {
	traceHttpRequest(c, req)
	res, err := c.HttpClient().Do(req)
	if res == nil {
		res = &http.Response{StatusCode: 0, Header: make(http.Header), Request: req}
	}
	traceHttpResponse(c, res)
	return res, err
}

func (c *Configuration) HttpClient() *http.Client {
	if c.httpClient == nil {
		tr := &http.Transport{}
		sslVerify, _ := c.GitConfig("http.sslverify")
		if sslVerify == "false" || len(os.Getenv("GIT_SSL_NO_VERIFY")) > 0 {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		c.httpClient = &http.Client{
			Transport:     tr,
			CheckRedirect: checkRedirect,
		}
	}
	return c.httpClient
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 3 {
		return errors.New("stopped after 3 redirects")
	}

	oldest := via[0]
	for key, _ := range oldest.Header {
		if key == "Authorization" {
			if req.URL.Scheme != oldest.URL.Scheme || req.URL.Host != oldest.URL.Host {
				continue
			}
		}
		req.Header.Set(key, oldest.Header.Get(key))
	}

	tracerx.Printf("api: redirect %s %s to %s", oldest.Method, oldest.URL, req.URL)

	return nil
}

var tracedTypes = []string{"json", "text", "xml", "html"}

func traceHttpRequest(c *Configuration, req *http.Request) {
	tracerx.Printf("HTTP: %s %s", req.Method, req.URL.String())

	if c.isTracingHttp == false {
		return
	}

	if req.Body != nil {
		req.Body = newCountedRequest(req)
	}

	fmt.Fprintf(os.Stderr, "> %s %s %s\n", req.Method, req.URL.RequestURI(), req.Proto)
	for key, _ := range req.Header {
		fmt.Fprintf(os.Stderr, "> %s: %s\n", key, req.Header.Get(key))
	}
}

func traceHttpResponse(c *Configuration, res *http.Response) {
	if res == nil {
		return
	}

	tracerx.Printf("HTTP: %d", res.StatusCode)

	if c.isTracingHttp == false {
		return
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "< %s %s\n", res.Proto, res.Status)
	for key, _ := range res.Header {
		fmt.Fprintf(os.Stderr, "< %s: %s\n", key, res.Header.Get(key))
	}

	traceBody := false
	ctype := strings.ToLower(strings.SplitN(res.Header.Get("Content-Type"), ";", 2)[0])
	for _, tracedType := range tracedTypes {
		if strings.Contains(ctype, tracedType) {
			traceBody = true
		}
	}

	res.Body = newCountedResponse(res)
	if traceBody {
		res.Body = newTracedBody(res.Body)
	}

	fmt.Fprintf(os.Stderr, "\n")
}

const (
	countingUpload = iota
	countingDownload
)

type countingBody struct {
	Direction int
	Size      int64
	io.ReadCloser
}

func (r *countingBody) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.Size += int64(n)
	return n, err
}

func (r *countingBody) Close() error {
	if r.Direction == countingUpload {
		fmt.Fprintf(os.Stderr, "* uploaded %d bytes\n", r.Size)
	} else {
		fmt.Fprintf(os.Stderr, "* downloaded %d bytes\n", r.Size)
	}
	return r.ReadCloser.Close()
}

func newCountedResponse(res *http.Response) *countingBody {
	return &countingBody{countingDownload, 0, res.Body}
}

func newCountedRequest(req *http.Request) *countingBody {
	return &countingBody{countingUpload, 0, req.Body}
}

type tracedBody struct {
	io.ReadCloser
}

func (r *tracedBody) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	fmt.Fprintf(os.Stderr, "%s\n", string(p[0:n]))
	return n, err
}

func (r *tracedBody) Close() error {
	return r.ReadCloser.Close()
}

func newTracedBody(body io.ReadCloser) *tracedBody {
	return &tracedBody{body}
}
