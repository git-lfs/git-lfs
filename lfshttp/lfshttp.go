package lfshttp

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/tr"
)

var (
	lfsMediaTypeRE  = regexp.MustCompile(`\Aapplication/vnd\.git\-lfs\+json(;|\z)`)
	jsonMediaTypeRE = regexp.MustCompile(`\Aapplication/json(;|\z)`)
)

type Context interface {
	GitConfig() *git.Configuration
	OSEnv() config.Environment
	GitEnv() config.Environment
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
		return errors.Wrapf(err, tr.Tr.Get("Unable to parse HTTP response for %s %s", res.Request.Method, res.Request.URL))
	}

	return nil
}
