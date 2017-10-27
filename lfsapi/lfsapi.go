package lfsapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"

	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
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

	commandCredHelper *commandCredentialHelper
	askpassCredHelper *AskPassCredentialHelper
	cachingCredHelper *credentialCacher
	gitEnv            config.Environment
	osEnv             config.Environment
	uc                *config.URLConfig
}

type Context interface {
	GitConfig() *git.Configuration
	OSEnv() config.Environment
	GitEnv() config.Environment
}

func NewClient(ctx Context) (*Client, error) {
	if ctx == nil {
		ctx = NewContext(nil, nil, nil)
	}

	gitEnv := ctx.GitEnv()
	osEnv := ctx.OSEnv()
	netrc, netrcfile, err := ParseNetrc(osEnv)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("bad netrc file %s", netrcfile))
	}

	cacheCreds := gitEnv.Bool("lfs.cachecredentials", true)
	var sshResolver SSHResolver = &sshAuthClient{os: osEnv}
	if cacheCreds {
		sshResolver = withSSHCache(sshResolver)
	}

	c := &Client{
		Endpoints:           NewEndpointFinder(ctx),
		SSH:                 sshResolver,
		Netrc:               netrc,
		DialTimeout:         gitEnv.Int("lfs.dialtimeout", 0),
		KeepaliveTimeout:    gitEnv.Int("lfs.keepalive", 0),
		TLSTimeout:          gitEnv.Int("lfs.tlstimeout", 0),
		ConcurrentTransfers: gitEnv.Int("lfs.concurrenttransfers", 3),
		SkipSSLVerify:       !gitEnv.Bool("http.sslverify", true) || osEnv.Bool("GIT_SSL_NO_VERIFY", false),
		Verbose:             osEnv.Bool("GIT_CURL_VERBOSE", false),
		DebuggingVerbose:    osEnv.Bool("LFS_DEBUG_HTTP", false),
		commandCredHelper: &commandCredentialHelper{
			SkipPrompt: osEnv.Bool("GIT_TERMINAL_PROMPT", false),
		},
		gitEnv: gitEnv,
		osEnv:  osEnv,
		uc:     config.NewURLConfig(gitEnv),
	}

	askpass, ok := osEnv.Get("GIT_ASKPASS")
	if !ok {
		askpass, ok = gitEnv.Get("core.askpass")
	}
	if !ok {
		askpass, _ = osEnv.Get("SSH_ASKPASS")
	}
	if len(askpass) > 0 {
		c.askpassCredHelper = &AskPassCredentialHelper{
			Program: askpass,
		}
	}

	if cacheCreds {
		c.cachingCredHelper = newCredentialCacher()
	}

	return c, nil
}

func (c *Client) GitEnv() config.Environment {
	return c.gitEnv
}

func (c *Client) OSEnv() config.Environment {
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

type testContext struct {
	gitConfig *git.Configuration
	osEnv     config.Environment
	gitEnv    config.Environment
}

func (c *testContext) GitConfig() *git.Configuration {
	return c.gitConfig
}

func (c *testContext) OSEnv() config.Environment {
	return c.osEnv
}

func (c *testContext) GitEnv() config.Environment {
	return c.gitEnv
}

func NewContext(gitConf *git.Configuration, osEnv, gitEnv map[string]string) Context {
	c := &testContext{gitConfig: gitConf}
	if c.gitConfig == nil {
		c.gitConfig = git.NewConfig("", "")
	}
	if osEnv != nil {
		c.osEnv = testEnv(osEnv)
	} else {
		c.osEnv = make(testEnv)
	}

	if gitEnv != nil {
		c.gitEnv = testEnv(gitEnv)
	} else {
		c.gitEnv = make(testEnv)
	}
	return c
}

type testEnv map[string]string

func (e testEnv) Get(key string) (v string, ok bool) {
	v, ok = e[key]
	return
}

func (e testEnv) GetAll(key string) []string {
	if v, ok := e.Get(key); ok {
		return []string{v}
	}
	return make([]string, 0)
}

func (e testEnv) Int(key string, def int) int {
	s, _ := e.Get(key)
	return config.Int(s, def)
}

func (e testEnv) Bool(key string, def bool) bool {
	s, _ := e.Get(key)
	return config.Bool(s, def)
}

func (e testEnv) All() map[string][]string {
	m := make(map[string][]string)
	for k, _ := range e {
		m[k] = e.GetAll(k)
	}
	return m
}
