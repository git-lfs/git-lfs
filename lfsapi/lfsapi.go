package lfsapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
)

var (
	lfsMediaTypeRE  = regexp.MustCompile(`\Aapplication/vnd\.git\-lfs\+json(;|\z)`)
	jsonMediaTypeRE = regexp.MustCompile(`\Aapplication/json(;|\z)`)
)

type Client struct {
	Endpoints   EndpointFinder
	Credentials CredentialHelper
	SSH         SSHResolver
	Netrc       NetrcFinder

	DialTimeout         int
	KeepaliveTimeout    int
	TLSTimeout          int
	ConcurrentTransfers int
	HTTPSProxy          string
	HTTPProxy           string
	NoProxy             string
	SkipSSLVerify       bool

	Verbose          bool
	DebuggingVerbose bool
	VerboseOut       io.Writer

	hostClients map[string]*http.Client
	clientMu    sync.Mutex

	ntlmSessions map[string]ntlm.ClientSession
	ntlmMu       sync.Mutex

	httpLogger *syncLogger

	LoggingStats bool // DEPRECATED

	// only used for per-host ssl certs
	gitEnv Env
	osEnv  Env
	uc     *config.URLConfig
}

func NewClient(osEnv Env, gitEnv Env) (*Client, error) {
	if osEnv == nil {
		osEnv = make(TestEnv)
	}

	if gitEnv == nil {
		gitEnv = make(TestEnv)
	}

	netrc, err := ParseNetrc(osEnv)
	if err != nil {
		return nil, err
	}

	httpsProxy, httpProxy, noProxy := getProxyServers(osEnv, gitEnv)

	var creds CredentialHelper = &commandCredentialHelper{
		SkipPrompt: !osEnv.Bool("GIT_TERMINAL_PROMPT", true),
	}
	var sshResolver SSHResolver = &sshAuthClient{os: osEnv, git: gitEnv}

	if gitEnv.Bool("lfs.cachecredentials", false) {
		creds = withCredentialCache(creds)
		sshResolver = withSSHCache(sshResolver)
	}

	c := &Client{
		Endpoints:           NewEndpointFinder(gitEnv),
		Credentials:         creds,
		SSH:                 sshResolver,
		Netrc:               netrc,
		DialTimeout:         gitEnv.Int("lfs.dialtimeout", 0),
		KeepaliveTimeout:    gitEnv.Int("lfs.keepalive", 0),
		TLSTimeout:          gitEnv.Int("lfs.tlstimeout", 0),
		ConcurrentTransfers: gitEnv.Int("lfs.concurrenttransfers", 3),
		SkipSSLVerify:       !gitEnv.Bool("http.sslverify", true) || osEnv.Bool("GIT_SSL_NO_VERIFY", false),
		Verbose:             osEnv.Bool("GIT_CURL_VERBOSE", false),
		DebuggingVerbose:    osEnv.Bool("LFS_DEBUG_HTTP", false),
		HTTPSProxy:          httpsProxy,
		HTTPProxy:           httpProxy,
		NoProxy:             noProxy,
		gitEnv:              gitEnv,
		osEnv:               osEnv,
		uc:                  config.NewURLConfig(gitEnv),
	}

	return c, nil
}

func (c *Client) GitEnv() Env {
	return c.gitEnv
}

func (c *Client) OSEnv() Env {
	return c.osEnv
}

func IsDecodeTypeError(err error) bool {
	_, ok := err.(*decodeTypeError)
	return ok
}

type decodeTypeError struct {
	Type string
}

func (e *decodeTypeError) TypeError() {}

func (e *decodeTypeError) Error() string {
	return fmt.Sprintf("Expected json type, got: %q", e.Type)
}

func DecodeJSON(res *http.Response, obj interface{}) error {
	ctype := res.Header.Get("Content-Type")
	if !(lfsMediaTypeRE.MatchString(ctype) || jsonMediaTypeRE.MatchString(ctype)) {
		return &decodeTypeError{Type: ctype}
	}

	err := json.NewDecoder(res.Body).Decode(obj)
	res.Body.Close()

	if err != nil {
		return errors.Wrapf(err, "Unable to parse HTTP response for %s %s", res.Request.Method, res.Request.URL)
	}

	return nil
}

// Env is an interface for the config.Environment methods that this package
// relies on.
type Env interface {
	Get(string) (string, bool)
	GetAll(string) []string
	Int(string, int) int
	Bool(string, bool) bool
	All() map[string][]string
}

type UniqTestEnv map[string]string

func (e UniqTestEnv) Get(key string) (v string, ok bool) {
	v, ok = e[key]
	return
}

func (e UniqTestEnv) GetAll(key string) []string {
	if v, ok := e.Get(key); ok {
		return []string{v}
	}
	return make([]string, 0)
}

func (e UniqTestEnv) Int(key string, def int) (val int) {
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

func (e UniqTestEnv) Bool(key string, def bool) (val bool) {
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

func (e UniqTestEnv) All() map[string][]string {
	m := make(map[string][]string)
	for k, _ := range e {
		m[k] = e.GetAll(k)
	}
	return m
}

// TestEnv is a basic config.Environment implementation. Only used in tests, or
// as a zero value to NewClient().
type TestEnv map[string][]string

func (e TestEnv) Get(key string) (string, bool) {
	all := e.GetAll(key)

	if len(all) == 0 {
		return "", false
	}
	return all[len(all)-1], true
}

func (e TestEnv) GetAll(key string) []string {
	return e[key]
}

func (e TestEnv) Int(key string, def int) (val int) {
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

func (e TestEnv) Bool(key string, def bool) (val bool) {
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

func (e TestEnv) All() map[string][]string {
	return e
}
