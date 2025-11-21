package lfshttp

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git/core"
	"github.com/git-lfs/git-lfs/v3/tr"
)

var (
	lfsMediaTypeRE  = regexp.MustCompile(`\Aapplication/vnd\.git\-lfs\+json(;|\z)`)
	jsonMediaTypeRE = regexp.MustCompile(`\Aapplication/json(;|\z)`)
)

type Context interface {
	GitConfig() *core.Configuration
	OSEnv() core.Environment
	GitEnv() core.Environment
}

func NewContext(gitConf *core.Configuration, osEnv, gitEnv map[string]string) Context {
	c := &testContext{gitConfig: gitConf}
	if c.gitConfig == nil {
		c.gitConfig = core.NewConfig("", "")
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

type testContext struct {
	gitConfig *core.Configuration
	osEnv     core.Environment
	gitEnv    core.Environment
}

func (c *testContext) GitConfig() *core.Configuration {
	return c.gitConfig
}

func (c *testContext) OSEnv() core.Environment {
	return c.osEnv
}

func (c *testContext) GitEnv() core.Environment {
	return c.gitEnv
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
	return tr.Tr.Get("Expected JSON type, got: %q", e.Type)
}

func DecodeJSON(res *http.Response, obj interface{}) error {
	ctype := res.Header.Get("Content-Type")
	if !(lfsMediaTypeRE.MatchString(ctype) || jsonMediaTypeRE.MatchString(ctype)) {
		return &decodeTypeError{Type: ctype}
	}

	err := json.NewDecoder(res.Body).Decode(obj)
	res.Body.Close()

	if err != nil {
		return errors.Wrap(err, tr.Tr.Get("Unable to parse HTTP response for %s %s", res.Request.Method, res.Request.URL))
	}

	return nil
}
