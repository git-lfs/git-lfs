package lfsapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

const MediaType = "application/vnd.git-lfs+json; charset=utf-8"

var (
	UserAgent = "git-lfs"
	httpRE    = regexp.MustCompile(`\Ahttps?://`)
)

var hintFileUrl = strings.TrimSpace(`
hint: The remote resolves to a file:// URL, which can only work with a
hint: standalone transfer agent.  See section "Using a Custom Transfer Type
hint: without the API server" in custom-transfers.md for details.
`)

func (c *Client) NewRequest(method string, e Endpoint, suffix string, body interface{}) (*http.Request, error) {
	if strings.HasPrefix(e.Url, "file://") {
		// Initial `\n` to avoid overprinting `Downloading LFS...`.
		fmt.Fprintf(os.Stderr, "\n%s\n", hintFileUrl)
	}

	sshRes, err := c.SSH.Resolve(e, method)
	if err != nil {
		tracerx.Printf("ssh: %s failed, error: %s, message: %s",
			e.SshUserAndHost, err.Error(), sshRes.Message,
		)

		if len(sshRes.Message) > 0 {
			return nil, errors.Wrap(err, sshRes.Message)
		}
		return nil, err
	}

	prefix := e.Url
	if len(sshRes.Href) > 0 {
		prefix = sshRes.Href
	}

	if !httpRE.MatchString(prefix) {
		urlfragment := strings.SplitN(prefix, "?", 2)[0]
		return nil, fmt.Errorf("missing protocol: %q", urlfragment)
	}

	req, err := http.NewRequest(method, joinURL(prefix, suffix), nil)
	if err != nil {
		return req, err
	}

	for key, value := range sshRes.Header {
		req.Header.Set(key, value)
	}
	req.Header.Set("Accept", MediaType)

	if body != nil {
		if merr := MarshalToRequest(req, body); merr != nil {
			return req, merr
		}
		req.Header.Set("Content-Type", MediaType)
	}

	return req, err
}

const slash = "/"

func joinURL(prefix, suffix string) string {
	if strings.HasSuffix(prefix, slash) {
		return prefix + suffix
	}
	return prefix + slash + suffix
}

// Do sends an HTTP request to get an HTTP response. It wraps net/http, adding
// extra headers, redirection handling, and error reporting.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header = c.extraHeadersFor(req)

	return c.do(req, "", nil)
}

// do performs an *http.Request respecting redirects, and handles the response
// as defined in c.handleResponse. Notably, it does not alter the headers for
// the request argument in any way.
func (c *Client) do(req *http.Request, remote string, via []*http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", UserAgent)

	res, err := c.doWithRedirects(c.httpClient(req.Host), req, remote, via)
	if err != nil {
		return res, err
	}

	return res, c.handleResponse(res)
}

// Close closes any resources that this client opened.
func (c *Client) Close() error {
	return c.httpLogger.Close()
}

func (c *Client) extraHeadersFor(req *http.Request) http.Header {
	extraHeaders := c.extraHeaders(req.URL)
	if len(extraHeaders) == 0 {
		return req.Header
	}

	copy := make(http.Header, len(req.Header))
	for k, vs := range req.Header {
		copy[k] = vs
	}

	for k, vs := range extraHeaders {
		for _, v := range vs {
			copy[k] = append(copy[k], v)
		}
	}
	return copy
}

func (c *Client) extraHeaders(u *url.URL) map[string][]string {
	hdrs := c.uc.GetAll("http", u.String(), "extraHeader")
	m := make(map[string][]string, len(hdrs))

	for _, hdr := range hdrs {
		parts := strings.SplitN(hdr, ":", 2)
		if len(parts) < 2 {
			continue
		}

		k, v := parts[0], strings.TrimSpace(parts[1])
		// If header keys are given in non-canonicalized form (e.g.,
		// "AUTHORIZATION" as opposed to "Authorization") they will not
		// be returned in calls to net/http.Header.Get().
		//
		// So, we avoid this problem by first canonicalizing header keys
		// for extra headers.
		k = textproto.CanonicalMIMEHeaderKey(k)

		m[k] = append(m[k], v)
	}
	return m
}

func (c *Client) doWithRedirects(cli *http.Client, req *http.Request, remote string, via []*http.Request) (*http.Response, error) {
	tracedReq, err := c.traceRequest(req)
	if err != nil {
		return nil, err
	}

	var retries int
	if n, ok := Retries(req); ok {
		retries = n
	} else {
		retries = defaultRequestRetries
	}

	var res *http.Response

	requests := tools.MaxInt(0, retries) + 1
	for i := 0; i < requests; i++ {
		res, err = cli.Do(req)
		if err == nil {
			break
		}

		if seek, ok := req.Body.(io.Seeker); ok {
			seek.Seek(0, io.SeekStart)
		}

		c.traceResponse(req, tracedReq, nil)
	}

	if err != nil {
		c.traceResponse(req, tracedReq, nil)
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	c.traceResponse(req, tracedReq, res)

	if res.StatusCode != 301 &&
		res.StatusCode != 302 &&
		res.StatusCode != 303 &&
		res.StatusCode != 307 &&
		res.StatusCode != 308 {

		// Above are the list of 3xx status codes that we know
		// how to handle below. If the status code contained in
		// the HTTP response was none of them, return the (res,
		// err) tuple as-is, otherwise handle the redirect.
		return res, err
	}

	redirectTo := res.Header.Get("Location")
	locurl, err := url.Parse(redirectTo)
	if err == nil && !locurl.IsAbs() {
		locurl = req.URL.ResolveReference(locurl)
		redirectTo = locurl.String()
	}

	via = append(via, req)
	if len(via) >= 3 {
		return res, errors.New("too many redirects")
	}

	redirectedReq, err := newRequestForRetry(req, redirectTo)
	if err != nil {
		return res, err
	}

	if len(req.Header.Get("Authorization")) > 0 {
		// If the original request was authenticated (noted by the
		// presence of the Authorization header), then recur through
		// doWithAuth, retaining the requests via but only after
		// authenticating the redirected request.
		return c.doWithAuth(remote, redirectedReq, via)
	}
	return c.doWithRedirects(cli, redirectedReq, remote, via)
}

func (c *Client) httpClient(host string) *http.Client {
	c.clientMu.Lock()
	defer c.clientMu.Unlock()

	if c.gitEnv == nil {
		c.gitEnv = make(testEnv)
	}

	if c.osEnv == nil {
		c.osEnv = make(testEnv)
	}

	if c.hostClients == nil {
		c.hostClients = make(map[string]*http.Client)
	}

	if client, ok := c.hostClients[host]; ok {
		return client
	}

	concurrentTransfers := c.ConcurrentTransfers
	if concurrentTransfers < 1 {
		concurrentTransfers = 8
	}

	dialtime := c.DialTimeout
	if dialtime < 1 {
		dialtime = 30
	}

	keepalivetime := c.KeepaliveTimeout
	if keepalivetime < 1 {
		keepalivetime = 1800
	}

	tlstime := c.TLSTimeout
	if tlstime < 1 {
		tlstime = 30
	}

	tr := &http.Transport{
		Proxy:               proxyFromClient(c),
		TLSHandshakeTimeout: time.Duration(tlstime) * time.Second,
		MaxIdleConnsPerHost: concurrentTransfers,
	}

	activityTimeout := 30
	if v, ok := c.uc.Get("lfs", fmt.Sprintf("https://%v", host), "activitytimeout"); ok {
		if i, err := strconv.Atoi(v); err == nil {
			activityTimeout = i
		} else {
			activityTimeout = 0
		}
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(dialtime) * time.Second,
		KeepAlive: time.Duration(keepalivetime) * time.Second,
		DualStack: true,
	}

	if activityTimeout > 0 {
		activityDuration := time.Duration(activityTimeout) * time.Second
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			c, err := dialer.DialContext(ctx, network, addr)
			if c == nil {
				return c, err
			}
			if tc, ok := c.(*net.TCPConn); ok {
				tc.SetKeepAlive(true)
				tc.SetKeepAlivePeriod(dialer.KeepAlive)
			}
			return &deadlineConn{Timeout: activityDuration, Conn: c}, err
		}
	} else {
		tr.DialContext = dialer.DialContext
	}

	tr.TLSClientConfig = &tls.Config{}

	if isClientCertEnabledForHost(c, host) {
		tracerx.Printf("http: client cert for %s", host)
		tr.TLSClientConfig.Certificates = []tls.Certificate{getClientCertForHost(c, host)}
		tr.TLSClientConfig.BuildNameToCertificate()
	}

	if isCertVerificationDisabledForHost(c, host) {
		tr.TLSClientConfig.InsecureSkipVerify = true
	} else {
		tr.TLSClientConfig.RootCAs = getRootCAsForHost(c, host)
	}

	httpClient := &http.Client{
		Transport: tr,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	c.hostClients[host] = httpClient
	if c.VerboseOut == nil {
		c.VerboseOut = os.Stderr
	}

	return httpClient
}

func (c *Client) CurrentUser() (string, string) {
	userName, _ := c.gitEnv.Get("user.name")
	userEmail, _ := c.gitEnv.Get("user.email")
	return userName, userEmail
}

func newRequestForRetry(req *http.Request, location string) (*http.Request, error) {
	newReq, err := http.NewRequest(req.Method, location, nil)
	if err != nil {
		return nil, err
	}

	if req.URL.Scheme == "https" && newReq.URL.Scheme == "http" {
		return nil, errors.New("lfsapi/client: refusing insecure redirect, https->http")
	}

	sameHost := req.URL.Host == newReq.URL.Host
	for key := range req.Header {
		if key == "Authorization" {
			if !sameHost {
				continue
			}
		}
		newReq.Header.Set(key, req.Header.Get(key))
	}

	oldestURL := strings.SplitN(req.URL.String(), "?", 2)[0]
	newURL := strings.SplitN(newReq.URL.String(), "?", 2)[0]
	tracerx.Printf("api: redirect %s %s to %s", req.Method, oldestURL, newURL)

	// This body will have already been rewound from a call to
	// lfsapi.Client.traceRequest().
	newReq.Body = req.Body
	newReq.ContentLength = req.ContentLength

	// Copy the request's context.Context, if any.
	newReq = newReq.WithContext(req.Context())

	return newReq, nil
}

type deadlineConn struct {
	Timeout time.Duration
	net.Conn
}

func (c *deadlineConn) Read(b []byte) (int, error) {
	if err := c.Conn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return 0, err
	}
	return c.Conn.Read(b)
}

func (c *deadlineConn) Write(b []byte) (int, error) {
	if err := c.Conn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return 0, err
	}

	return c.Conn.Write(b)
}

func init() {
	UserAgent = config.VersionDesc
}
