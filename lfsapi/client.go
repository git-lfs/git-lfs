package lfsapi

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	res, err := c.doWithRedirects(c.httpClient(req.Host), req, nil)
	if err != nil {
		return res, err
	}

	return res, c.handleResponse(res)
}

func (c *Client) doWithRedirects(cli *http.Client, req *http.Request, via []*http.Request) (*http.Response, error) {
	if seeker, ok := req.Body.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	res, err := cli.Do(req)
	if err != nil {
		return res, err
	}

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
	redirectedReq, err := http.NewRequest(req.Method, redirectTo, nil)
	if err != nil {
		return res, errors.Wrapf(err, err.Error())
	}

	redirectedReq.Body = req.Body
	redirectedReq.ContentLength = req.ContentLength

	if err = checkRedirect(redirectedReq, via); err != nil {
		return res, errors.Wrapf(err, err.Error())
	}

	return c.doWithRedirects(cli, redirectedReq, via)
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
		Proxy: ProxyFromClient(c),
		Dial: (&net.Dialer{
			Timeout:   time.Duration(dialtime) * time.Second,
			KeepAlive: time.Duration(keepalivetime) * time.Second,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(tlstime) * time.Second,
		MaxIdleConnsPerHost: concurrentTransfers,
	}

	tr.TLSClientConfig = &tls.Config{}
	if isCertVerificationDisabledForHost(c, host) {
		tr.TLSClientConfig.InsecureSkipVerify = true
	} else {
		tr.TLSClientConfig.RootCAs = getRootCAsForHost(c, host)
	}

	httpClient := &http.Client{
		Transport:     tr,
		CheckRedirect: checkRedirect,
	}

	c.hostClients[host] = httpClient

	return httpClient
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 3 {
		return errors.New("stopped after 3 redirects")
	}

	oldest := via[0]
	for key := range oldest.Header {
		if key == "Authorization" {
			if req.URL.Scheme != oldest.URL.Scheme || req.URL.Host != oldest.URL.Host {
				continue
			}
		}
		req.Header.Set(key, oldest.Header.Get(key))
	}

	oldestURL := strings.SplitN(oldest.URL.String(), "?", 2)[0]
	newURL := strings.SplitN(req.URL.String(), "?", 2)[0]
	tracerx.Printf("api: redirect %s %s to %s", oldest.Method, oldestURL, newURL)

	return nil
}
