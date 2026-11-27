package lfshttp

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

func (c *Client) traceRequest(req *http.Request) (*tracedRequest, error) {
	tracerx.Printf("HTTP: %s", traceReq(req))

	if c.Verbose {
		if dump, err := httputil.DumpRequestOut(req, false); err == nil {
			c.traceHTTPDump(">", dump)
		}
	}

	body, ok := req.Body.(ReadSeekCloser)
	if body != nil && !ok {
		return nil, errors.New(tr.Tr.Get("Request body must implement io.ReadCloser and io.Seeker: %T", body))
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

func (c *Client) traceResponse(req *http.Request, tracedReq *tracedRequest, res *http.Response) {
	if tracedReq != nil {
		c.httpLogger.LogRequest(req, tracedReq.BodySize)
	}

	if res == nil {
		c.httpLogger.LogResponse(req, -1, 0)
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
	eof        bool
	io.ReadCloser
}

func (r *tracedResponse) Read(b []byte) (int, error) {
	n, err := tracedRead(r.ReadCloser, b, r.verboseOut, r.gitTrace, r.verbose)
	r.BodySize += int64(n)

	if err == io.EOF && !r.eof {
		r.httpLogger.LogResponse(r.response.Request, r.response.StatusCode, r.BodySize)
		r.eof = true
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

type traceTLSKey struct{}

func (c *Client) traceTLS(req *http.Request) *http.Request {
	if !c.TraceTLS || req.Context().Value(traceTLSKey{}) != nil {
		return req
	}

	ctx := httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			for _, line := range traceTLSState(cs, err) {
				tracerx.Printf("tls: %s", line)
			}
		},
	})

	return req.WithContext(context.WithValue(ctx, traceTLSKey{}, true))
}

func traceTLSState(cs tls.ConnectionState, err error) []string {
	if err != nil {
		return []string{fmt.Sprintf("handshake failed: %s", err)}
	}

	lines := []string{traceNegotiated(cs)}

	if cs.DidResume {
		lines = append(lines, "connection resumed")
	}

	if len(cs.PeerCertificates) == 0 {
		return lines
	}

	return append(lines, traceTLSCertificates(cs)...)
}

func traceNegotiated(cs tls.ConnectionState) string {
	negotiated := fmt.Sprintf(
		"negotiated %s / %s",
		tls.VersionName(cs.Version),
		tls.CipherSuiteName(cs.CipherSuite),
	)

	if len(cs.NegotiatedProtocol) > 0 {
		negotiated += fmt.Sprintf(" (ALPN %q)", cs.NegotiatedProtocol)
	}

	return negotiated
}

func traceTLSCertificates(cs tls.ConnectionState) []string {
	leaf := cs.PeerCertificates[0]

	lines := []string{
		fmt.Sprintf("server certificate subject: %s", leaf.Subject.String()),
		fmt.Sprintf("server certificate issuer: %s", leaf.Issuer.String()),
		fmt.Sprintf("server certificate algorithms: %s / %s", leaf.PublicKeyAlgorithm.String(), leaf.SignatureAlgorithm.String()),
	}

	chain := make([]string, 0, len(cs.PeerCertificates))
	for _, cert := range cs.PeerCertificates {
		chain = append(chain, cert.Subject.String())
	}

	lines = append(lines, "server certificate chain: "+strings.Join(chain, " -> "))

	if len(cs.VerifiedChains) > 0 {
		lines = append(lines, "server certificate verified")
	} else {
		lines = append(lines, "server certificate not verified")
	}

	return lines
}
