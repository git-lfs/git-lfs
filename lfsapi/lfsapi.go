package lfsapi

import (
	"encoding/json"
	"net"
	"net/http"
	"regexp"
	"strings"
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

	HTTPSProxy string
	HTTPProxy  string
	NoProxy    string

	hostClients map[string]*http.Client
	clientMu    sync.Mutex
}

func NewClient(osEnv env, gitEnv env) (*Client, error) {
	if osEnv == nil {
		osEnv = make(testEnv)
	}

	if gitEnv == nil {
		gitEnv = make(testEnv)
	}

	netrc, err := ParseNetrc(osEnv)
	if err != nil {
		return nil, err
	}

	httpsProxy, httpProxy, noProxy := getProxyServers(osEnv, gitEnv)

	return &Client{
		Endpoints: NewEndpointFinder(gitEnv),
		Credentials: &CommandCredentialHelper{
			SkipPrompt: !osEnv.Bool("GIT_TERMINAL_PROMPT", true),
		},
		Netrc:      netrc,
		HTTPSProxy: httpsProxy,
		HTTPProxy:  httpProxy,
		NoProxy:    noProxy,
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
		Proxy: ProxyFromClient(c),
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

func getProxyServers(osEnv env, gitEnv env) (string, string, string) {
	var httpsProxy string
	httpProxy, _ := gitEnv.Get("http.proxy")
	if strings.HasPrefix(httpProxy, "https://") {
		httpsProxy = httpProxy
	}

	if len(httpsProxy) == 0 {
		httpsProxy, _ = osEnv.Get("HTTPS_PROXY")
	}

	if len(httpsProxy) == 0 {
		httpsProxy, _ = osEnv.Get("https_proxy")
	}

	if len(httpProxy) == 0 {
		httpProxy, _ = osEnv.Get("HTTP_PROXY")
	}

	if len(httpProxy) == 0 {
		httpProxy, _ = osEnv.Get("http_proxy")
	}

	noProxy, _ := osEnv.Get("NO_PROXY")
	if len(noProxy) == 0 {
		noProxy, _ = osEnv.Get("no_proxy")
	}

	return httpsProxy, httpProxy, noProxy
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

// basic config.Environment implementation. Only used in tests, or as a zero
// value to NewClient().
type testEnv map[string]string

func (e testEnv) Get(key string) (string, bool) {
	v, ok := e[key]
	return v, ok
}

func (e testEnv) Bool(key string, def bool) (val bool) {
	s, _ := e.Get(key)
	if len(s) == 0 {
		return def
	}

	switch strings.ToLower(s) {
	case "true", "1", "on", "yes", "t":
		return true
	case "false", "0", "off", "no", "f":
		return false
	default:
		return false
	}
}

func (e testEnv) All() map[string]string {
	return e
}
