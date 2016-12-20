package lfsapi

import (
	"encoding/json"
	"net"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/errors"
)

var (
	lfsMediaTypeRE  = regexp.MustCompile(`\Aapplication/vnd\.git\-lfs\+json(;|\z)`)
	jsonMediaTypeRE = regexp.MustCompile(`\Aapplication/json(;|\z)`)
)

type Client struct {
	Endpoints   EndpointFinder
	Credentials CredentialHelper
	Netrc       NetrcFinder

	DialTimeout         int `git:"lfs.dialtimeout"`
	KeepaliveTimeout    int `git:"lfs.keepalive"`
	TLSTimeout          int `git:"lfs.tlstimeout"`
	ConcurrentTransfers int `git:"lfs.concurrenttransfers"`

	hostClients map[string]*http.Client
	clientMu    sync.Mutex
}

func NewClient(osEnv env, gitEnv env) (*Client, error) {
	netrc, err := ParseNetrc(osEnv)
	if err != nil {
		return nil, err
	}

	return &Client{
		Endpoints: NewEndpointFinder(gitEnv),
		Credentials: &CommandCredentialHelper{
			SkipPrompt: !osEnv.Bool("GIT_TERMINAL_PROMPT", true),
		},
		Netrc: netrc,
	}, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	res, err := c.httpClient(req.Host).Do(req)
	if err != nil {
		return res, err
	}

	return res, c.handleResponse(res)
}

func (c *Client) httpClient(host string) *http.Client {
	c.clientMu.Lock()
	defer c.clientMu.Unlock()

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
		Dial: (&net.Dialer{
			Timeout:   time.Duration(dialtime) * time.Second,
			KeepAlive: time.Duration(keepalivetime) * time.Second,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(tlstime) * time.Second,
		MaxIdleConnsPerHost: concurrentTransfers,
	}

	httpClient := &http.Client{
		Transport: tr,
	}

	c.hostClients[host] = httpClient

	return httpClient
}

func decodeResponse(res *http.Response, obj interface{}) error {
	ctype := res.Header.Get("Content-Type")
	if !(lfsMediaTypeRE.MatchString(ctype) || jsonMediaTypeRE.MatchString(ctype)) {
		return nil
	}

	err := json.NewDecoder(res.Body).Decode(obj)
	res.Body.Close()

	if err != nil {
		return errors.Wrapf(err, "Unable to parse HTTP response for %s %s", res.Request.Method, res.Request.URL)
	}

	return nil
}
