package lfsapi

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

var UserAgent = "git-lfs"

const MediaType = "application/vnd.git-lfs+json; charset=utf-8"

func (c *Client) NewRequest(method string, e Endpoint, suffix string, body interface{}) (*http.Request, error) {
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

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", UserAgent)

	res, err := c.doWithRedirects(c.httpClient(req.Host), req, nil)
	if err != nil {
		return res, err
	}

	return res, c.handleResponse(res)
}

func (c *Client) setExtraHeaders(req *http.Request) http.Header {
	var copy http.Header = make(http.Header)
	for k, vs := range req.Header {
		copy[k] = vs
	}

	for k, vs := range c.extraHeaders(req.URL) {
		for _, v := range vs {
			copy[k] = append(copy[k], v)
		}
	}
	return copy
}

func (c *Client) extraHeaders(u *url.URL) map[string][]string {
	if c.uc == nil {
		c.uc = config.NewURLConfig(c.GitEnv())
	}

	hdrs := c.uc.GetAll("http", u.String(), "extraHeader")
	m := make(map[string][]string, len(hdrs))

	for _, hdr := range hdrs {
		parts := strings.SplitN(hdr, ": ", 2)
		if len(parts) < 2 {
			continue
		}

		k, v := parts[0], parts[1]

		m[k] = append(m[k], v)
	}
	return m
}

func (c *Client) doWithRedirects(cli *http.Client, req *http.Request, via []*http.Request) (*http.Response, error) {
	c.traceRequest(req)
	if err := c.prepareRequestBody(req); err != nil {
		return nil, err
	}

	start := time.Now()
	res, err := cli.Do(req)
	if err != nil {
		return res, err
	}

	c.traceResponse(res)
	c.startResponseStats(res, start)

	if res.StatusCode != 307 {
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

	return c.doWithRedirects(cli, redirectedReq, via)
}

func (c *Client) httpClient(host string) *http.Client {
	c.clientMu.Lock()
	defer c.clientMu.Unlock()

	if c.gitEnv == nil {
		c.gitEnv = make(TestEnv)
	}

	if c.osEnv == nil {
		c.osEnv = make(TestEnv)
	}

	if c.hostClients == nil {
		c.hostClients = make(map[string]*http.Client)
	}

	if client, ok := c.hostClients[host]; ok {
		return client
	}

	concurrentTransfers := c.ConcurrentTransfers
	if concurrentTransfers < 1 {
		concurrentTransfers = 3
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
		Proxy: proxyFromClient(c),
		Dial: (&net.Dialer{
			Timeout:   time.Duration(dialtime) * time.Second,
			KeepAlive: time.Duration(keepalivetime) * time.Second,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(tlstime) * time.Second,
		MaxIdleConnsPerHost: concurrentTransfers,
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

	for key := range req.Header {
		if key == "Authorization" {
			if req.URL.Scheme != newReq.URL.Scheme || req.URL.Host != newReq.URL.Host {
				continue
			}
		}
		newReq.Header.Set(key, req.Header.Get(key))
	}

	oldestURL := strings.SplitN(req.URL.String(), "?", 2)[0]
	newURL := strings.SplitN(newReq.URL.String(), "?", 2)[0]
	tracerx.Printf("api: redirect %s %s to %s", req.Method, oldestURL, newURL)

	newReq.Body = req.Body
	newReq.ContentLength = req.ContentLength
	return newReq, nil
}

func init() {
	UserAgent = config.VersionDesc
}
