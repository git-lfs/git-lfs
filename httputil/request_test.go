package httputil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type AuthenticateHeaderTestCase struct {
	ExpectedAuthType string
	Headers          map[string][]string
}

func (c *AuthenticateHeaderTestCase) Assert(t *testing.T) {
	t.Log("lfs/httputil: asserting auth type: %q for: %v", c.ExpectedAuthType, c.Headers)

	assert.Equal(t, c.ExpectedAuthType, GetAuthType(c.HttpResponse()))
}

func (c *AuthenticateHeaderTestCase) HttpResponse() *http.Response {
	res := &http.Response{Header: make(http.Header)}

	for k, vv := range c.Headers {
		for _, v := range vv {
			res.Header.Add(k, v)
		}
	}

	return res
}

func TestGetAuthType(t *testing.T) {
	for _, c := range []AuthenticateHeaderTestCase{
		{BasicAuthType, map[string][]string{}},
		{NtlmAuthType, map[string][]string{"WWW-Authenticate": {"Basic", "NTLM", "Bearer"}}},
		{NtlmAuthType, map[string][]string{"LFS-Authenticate": {"Basic", "NTLM", "Bearer"}}},
		{NtlmAuthType, map[string][]string{"LFS-Authenticate": {"Basic", "Ntlm"}}},
		{NtlmAuthType, map[string][]string{"Www-Authenticate": {"Basic", "Ntlm"}}},
		{NtlmAuthType, map[string][]string{"WWW-Authenticate": {"Basic"},
			"LFS-Authenticate": {"Ntlm"}}},
	} {
		c.Assert(t)
	}
}
