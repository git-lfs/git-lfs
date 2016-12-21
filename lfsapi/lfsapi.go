package lfsapi

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

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

	DialTimeout         int
	KeepaliveTimeout    int
	TLSTimeout          int
	ConcurrentTransfers int
	HTTPSProxy          string
	HTTPProxy           string
	NoProxy             string
	SkipSSLVerify       bool

	hostClients map[string]*http.Client
	clientMu    sync.Mutex

	// only used for per-host ssl certs
	gitEnv env
	osEnv  env
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

	c := &Client{
		Endpoints: NewEndpointFinder(gitEnv),
		Credentials: &CommandCredentialHelper{
			SkipPrompt: !osEnv.Bool("GIT_TERMINAL_PROMPT", true),
		},
		Netrc:               netrc,
		DialTimeout:         gitEnv.Int("lfs.dialtimeout", 0),
		KeepaliveTimeout:    gitEnv.Int("lfs.keepalive", 0),
		TLSTimeout:          gitEnv.Int("lfs.tlstimeout", 0),
		ConcurrentTransfers: gitEnv.Int("lfs.concurrenttransfers", 0),
		SkipSSLVerify:       !gitEnv.Bool("http.sslverify", true) || osEnv.Bool("GIT_SSL_NO_VERIFY", false),
		HTTPSProxy:          httpsProxy,
		HTTPProxy:           httpProxy,
		NoProxy:             noProxy,
		gitEnv:              gitEnv,
		osEnv:               osEnv,
	}

	return c, nil
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

type env interface {
	Get(string) (string, bool)
	Int(string, int) int
	Bool(string, bool) bool
	All() map[string]string
}

// basic config.Environment implementation. Only used in tests, or as a zero
// value to NewClient().
type testEnv map[string]string

func (e testEnv) Get(key string) (string, bool) {
	v, ok := e[key]
	return v, ok
}

func (e testEnv) Int(key string, def int) (val int) {
	s, _ := e.Get(key)
	if len(s) == 0 {
		return def
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}

	return i
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
