package lfsapi

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/rubyist/tracerx"
)

func (c *Client) traceRequest(req *http.Request) {
	tracerx.Printf("HTTP: %s", traceReq(req))

	if !c.IsTracing {
		return
	}

	dump, err := httputil.DumpRequest(req, false)
	if err != nil {
		return
	}

	c.traceHTTPDump(">", dump)
}

func (c *Client) traceResponse(res *http.Response) {
	if res == nil {
		return
	}

	tracerx.Printf("HTTP: %d", res.StatusCode)

	if c.IsTracing == false {
		return
	}

	dump, err := httputil.DumpResponse(res, false)
	if err != nil {
		return
	}

	if isTraceableContent(res.Header) {
		fmt.Fprintf(os.Stderr, "\n\n")
	} else {
		fmt.Fprintf(os.Stderr, "\n")
	}

	c.traceHTTPDump("<", dump)
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
