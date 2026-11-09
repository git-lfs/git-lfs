package lfshttp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"regexp"
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
	n, err := tracedRead(r.ReadSeekCloser, b, r.verboseOut, false, r.verbose, nil)
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
	tracedRes := &tracedResponse{
		httpLogger: c.httpLogger,
		response:   res,
		gitTrace:   verboseBody,
		verbose:    verboseBody && c.Verbose,
		verboseOut: c.VerboseOut,
		ReadCloser: res.Body,
	}
	if isJSONContent(res.Header) && !c.DebuggingVerbose {
		tracedRes.redactor = &authRedactor{}
	}
	res.Body = tracedRes

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
	redactor   *authRedactor
	eof        bool
	io.ReadCloser
}

func (r *tracedResponse) Read(b []byte) (int, error) {
	n, err := tracedRead(r.ReadCloser, b, r.verboseOut, r.gitTrace, r.verbose, r.redactor)
	r.BodySize += int64(n)

	if err == io.EOF && !r.eof {
		r.httpLogger.LogResponse(r.response.Request, r.response.StatusCode, r.BodySize)
		r.eof = true
	}
	return n, err
}

func (r *tracedResponse) Close() error {
	if r.redactor != nil {
		traceBody(r.redactor.redact(nil, true), r.verboseOut, r.gitTrace, r.verbose)
	}
	return r.ReadCloser.Close()
}

func tracedRead(r io.Reader, b []byte, verboseOut io.Writer, gitTrace, verbose bool, redactor *authRedactor) (int, error) {
	n, err := r.Read(b)
	if err == nil || err == io.EOF {
		if gitTrace || verbose {
			chunk := b[0:n]
			if redactor != nil {
				chunk = redactor.redact(chunk, err == io.EOF)
			}
			traceBody(chunk, verboseOut, gitTrace, verbose)
		}
	}

	return n, err
}

func traceBody(b []byte, verboseOut io.Writer, gitTrace, verbose bool) {
	if len(b) == 0 {
		return
	}

	chunk := string(b)
	if gitTrace {
		tracerx.Printf("HTTP: %s", chunk)
	}

	if verbose {
		fmt.Fprint(verboseOut, chunk)
	}
}

// authRedactorMaxWithheld is the maximum amount of data an authRedactor
// withholds while waiting for the remainder of an incomplete member, which
// is larger than any header value an HTTP server would reasonably accept.
const authRedactorMaxWithheld = 64 * 1024

var (
	// authorizationKey is the quoted key, in lowercase, of the JSON
	// object members whose values authRedactor anonymizes.
	authorizationKey = []byte(`"authorization"`)

	// authorizationValueRE matches the remainder of a JSON object member
	// following its key, capturing the content of the member's string
	// value after any leading authentication scheme, up to either the
	// end of the string or the end of the available data.
	authorizationValueRE = regexp.MustCompile(`\A\s*:\s*"(?:[^"\\\s]+\s+)?((?:[^"\\]|\\.)*)`)

	// authorizationPrefixRE matches the remainder of a JSON object member
	// following its key when the available data ends before the start of
	// the member's value.
	authorizationPrefixRE = regexp.MustCompile(`\A\s*:?\s*\z`)
)

// authRedactor anonymizes the credentials in the values of any
// "Authorization" members of JSON objects in a JSON HTTP response body,
// such as those in the "header" objects of the actions in a Batch API
// response, while the body is read and traced.
//
// Because a member may be split across several reads of the body, any
// data at the end of one read which might belong to an incomplete member
// is withheld and only traced once enough subsequent data has been read,
// or once the amount of withheld data exceeds authRedactorMaxWithheld.
type authRedactor struct {
	withheld []byte
}

// redact returns any previously withheld data followed by the data in b,
// with the credentials in any "Authorization" members anonymized.  Unless
// final is true, any data at the end which might belong to an incomplete
// member is withheld from the result until the next call.
func (a *authRedactor) redact(b []byte, final bool) []byte {
	data := b
	if len(a.withheld) > 0 {
		data = append(a.withheld, b...)
		a.withheld = nil
	}

	var out []byte
	start := 0
	for i := 0; ; {
		j := bytes.IndexByte(data[i:], '"')
		if j < 0 {
			break
		}
		i += j

		rest := data[i:]
		flush := final || len(rest) > authRedactorMaxWithheld
		incomplete := false
		if len(rest) < len(authorizationKey) {
			// The data may end partway through a key.
			incomplete = bytes.EqualFold(rest, authorizationKey[:len(rest)])
		} else if bytes.EqualFold(rest[:len(authorizationKey)], authorizationKey) {
			offset := i + len(authorizationKey)
			value := data[offset:]
			if m := authorizationValueRE.FindSubmatchIndex(value); m == nil {
				// The data may end before the start of the value.
				incomplete = authorizationPrefixRE.Match(value)
			} else if flush || (m[1] < len(value) && value[m[1]] == '"') {
				out = append(out, data[start:offset+m[2]]...)
				out = append(out, "* * * * *"...)
				start = offset + m[3]
				i = start
				continue
			} else {
				// The data ends partway through the value.
				incomplete = true
			}
		}

		if incomplete && !flush {
			a.withheld = bytes.Clone(rest)
			return append(out, data[start:i]...)
		}
		i++
	}

	return append(out, data[start:]...)
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

func isJSONContent(h http.Header) bool {
	ctype := strings.ToLower(strings.SplitN(h.Get("Content-Type"), ";", 2)[0])
	return strings.Contains(ctype, "json")
}

func traceReq(req *http.Request) string {
	return fmt.Sprintf("%s %s", req.Method, strings.SplitN(req.URL.String(), "?", 2)[0])
}
