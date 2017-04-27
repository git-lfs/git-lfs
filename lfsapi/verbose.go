package lfsapi

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/rubyist/tracerx"
)

func (c *Client) traceRequest(req *http.Request) (*tracedRequest, error) {
	tracerx.Printf("HTTP: %s", traceReq(req))

	if c.Verbose {
		if dump, err := httputil.DumpRequest(req, false); err == nil {
			c.traceHTTPDump(">", dump)
		}
	}

	body, ok := req.Body.(ReadSeekCloser)
	if body != nil && !ok {
		return nil, fmt.Errorf("Request body must implement io.ReadCloser and io.Seeker. Got: %T", body)
	}

	if body != nil && ok {
		body.Seek(0, io.SeekStart)
		tr := &tracedRequest{
			verbose:        c.Verbose && isTraceableContent(req.Header),
			verboseOut:     c.VerboseOut,
			ReadSeekCloser: body,
		}
		req.Body = tr
		return tr, nil
	}

	return nil, nil
}

type tracedRequest struct {
	BodySize   int64
	verbose    bool
	verboseOut io.Writer
	ReadSeekCloser
}

func (r *tracedRequest) Read(b []byte) (int, error) {
	n, err := tracedRead(r.ReadSeekCloser, b, r.verboseOut, false, r.verbose)
	r.BodySize += int64(n)
	return n, err
}

func (c *Client) traceResponse(tracedReq *tracedRequest, res *http.Response) {
	if tracedReq != nil {
		c.httpLogger.Log(res.Request, "request",
			fmt.Sprintf("body=%d ", tracedReq.BodySize))
	}

	if res == nil {
		return
	}

	tracerx.Printf("HTTP: %d", res.StatusCode)

	verboseBody := isTraceableContent(res.Header)
	res.Body = &tracedResponse{
		httpLogger: c.httpLogger,
		response:   res,
		gitTrace:   verboseBody,
		verbose:    verboseBody && c.Verbose,
		verboseOut: c.VerboseOut,
		ReadCloser: res.Body,
	}

	if !c.Verbose {
		return
	}

	if dump, err := httputil.DumpResponse(res, false); err == nil {
		if verboseBody {
			fmt.Fprintf(c.VerboseOut, "\n\n")
		} else {
			fmt.Fprintf(c.VerboseOut, "\n")
		}
		c.traceHTTPDump("<", dump)
	}
}

type tracedResponse struct {
	BodySize   int64
	httpLogger *syncLogger
	response   *http.Response
	verbose    bool
	gitTrace   bool
	verboseOut io.Writer
	io.ReadCloser
}

func (r *tracedResponse) Read(b []byte) (int, error) {
	n, err := tracedRead(r.ReadCloser, b, r.verboseOut, r.gitTrace, r.verbose)
	r.BodySize += int64(n)

	if err == io.EOF {
		r.httpLogger.Log(r.response.Request, "response",
			fmt.Sprintf("status=%d body=%d ", r.response.StatusCode, r.BodySize))
	}
	return n, err
}

func tracedRead(r io.Reader, b []byte, verboseOut io.Writer, gitTrace, verbose bool) (int, error) {
	n, err := r.Read(b)
	if err == nil || err == io.EOF {
		if n > 0 && (gitTrace || verbose) {
			chunk := string(b[0:n])
			if gitTrace {
				tracerx.Printf("HTTP: %s", chunk)
			}

			if verbose {
				fmt.Fprint(verboseOut, chunk)
			}
		}
	}

	return n, err
}

func (c *Client) traceHTTPDump(direction string, dump []byte) {
	scanner := bufio.NewScanner(bytes.NewBuffer(dump))

	for scanner.Scan() {
		line := scanner.Text()
		if !c.DebuggingVerbose && strings.HasPrefix(strings.ToLower(line), "authorization: basic") {
			fmt.Fprintf(c.VerboseOut, "%s Authorization: Basic * * * * *\n", direction)
		} else {
			fmt.Fprintf(c.VerboseOut, "%s %s\n", direction, line)
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

func annotateReqStart(r *http.Request) *http.Request {
	ctx := r.Context()
	v := ctx.Value(transferKey)
	if v == nil {
		return r
	}

	t := v.(httpTransfer)
	t.Start = time.Now()
	return r.WithContext(context.WithValue(ctx, transferKey, t))
}
