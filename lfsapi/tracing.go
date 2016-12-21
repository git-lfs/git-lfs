package lfsapi

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/rubyist/tracerx"
)

func (c *Client) traceRequest(req *http.Request) error {
	tracerx.Printf("HTTP: %s", traceReq(req))

	traced := &tracedRequest{
		isTraceableType: c.IsTracing && isTraceableContent(req.Header),
	}

	body, ok := req.Body.(ReadSeekCloser)
	if body != nil && !ok {
		return fmt.Errorf("Request body must implement io.ReadCloser and io.Seeker. Got: %T", body)
	}

	if ok {
		traced.ReadSeekCloser = body
	}

	if !c.IsTracing {
		return nil
	}

	dump, err := httputil.DumpRequest(req, false)
	if err != nil {
		return err
	}

	c.traceHTTPDump(">", dump)
	return nil
}

type tracedRequest struct {
	Count           int
	isTraceableType bool
	useStderrTrace  bool
	ReadSeekCloser
}

func (r *tracedRequest) Read(b []byte) (int, error) {
	n, err := tracedRead(r.ReadSeekCloser, b, r.isTraceableType, false, r.useStderrTrace)
	r.Count += n
	return n, err
}

func (c *Client) traceResponse(res *http.Response) *tracedResponse {
	isTraceable := isTraceableContent(res.Header)
	traced := &tracedResponse{
		client:          c,
		response:        res,
		isLogging:       c.IsLogging,
		isTraceableType: isTraceable,
		useStderrTrace:  c.IsTracing,
	}

	if res == nil {
		return traced
	}

	traced.ReadCloser = res.Body
	tracerx.Printf("HTTP: %d", res.StatusCode)

	if c.IsTracing == false {
		return traced
	}

	dump, err := httputil.DumpResponse(res, false)
	if err != nil {
		return traced
	}

	if isTraceable {
		fmt.Fprintf(os.Stderr, "\n\n")
	} else {
		fmt.Fprintf(os.Stderr, "\n")
	}

	c.traceHTTPDump("<", dump)
	return traced
}

type tracedResponse struct {
	Count           int
	client          *Client
	response        *http.Response
	isLogging       bool
	isTraceableType bool
	useStderrTrace  bool
	io.ReadCloser
}

func (r *tracedResponse) Read(b []byte) (int, error) {
	n, err := tracedRead(r.ReadCloser, b, r.isTraceableType, true, r.useStderrTrace)
	r.Count += n

	if err == io.EOF && r.isLogging && r.response != nil {
		r.client.FinishResponseStats(r.response, int64(r.Count))
	}
	return n, err
}

func tracedRead(r io.Reader, b []byte, isTraceable, useGitTrace, useStderrTrace bool) (int, error) {
	n, err := r.Read(b)
	if err == nil || err == io.EOF {
		if n > 0 && isTraceable {
			chunk := string(b[0:n])
			if useGitTrace {
				tracerx.Printf("HTTP: %s", chunk)
			}

			if useStderrTrace {
				fmt.Fprint(os.Stderr, chunk)
			}
		}
	}

	return n, err
}

func (c *Client) traceHTTPDump(direction string, dump []byte) {
	scanner := bufio.NewScanner(bytes.NewBuffer(dump))

	for scanner.Scan() {
		line := scanner.Text()
		if !c.IsDebugging && strings.HasPrefix(strings.ToLower(line), "authorization: basic") {
			fmt.Fprintf(os.Stderr, "%s Authorization: Basic * * * * *\n", direction)
		} else {
			fmt.Fprintf(os.Stderr, "%s %s\n", direction, line)
		}
	}
}

var tracedTypes = []string{"json", "text", "xml", "html"}

func isTraceableContent(h http.Header) bool {
	ctype := strings.ToLower(strings.SplitN(h.Get("Content-Type"), ";", 2)[0])
	for _, tracedType := range tracedTypes {
		if strings.Contains(ctype, tracedType) {
			return true
		}
	}
	return false
}

func traceReq(req *http.Request) string {
	return fmt.Sprintf("%s %s", req.Method, strings.SplitN(req.URL.String(), "?", 2)[0])
}
