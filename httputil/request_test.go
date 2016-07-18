package httputil

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetAutTypeNoAuthHeaders(t *testing.T) {
	res := &http.Response{Header: make(http.Header)}

	authType := GetAuthType(res)

	assert.Equal(t, BasicAuthType, authType)
}

func TestGetAutTypewwAuthenticateBasicNtlmBearer(t *testing.T) {

	res := &http.Response{Header: make(http.Header)}
	// Intentionally using header name in the non-canonical form.
	res.Header.Add("WWW-Authenticate", "Basic")
	res.Header.Add("WWW-Authenticate", "NTLM")
	res.Header.Add("WWW-Authenticate", "Bearer")

	authType := GetAuthType(res)

	assert.Equal(t, NtlmAuthType, authType)
}

func TestGetAutTypeNoWwwAuthenticateLfsAuthenticateNtlm(t *testing.T) {

	res := &http.Response{Header: make(http.Header)}
	res.Header.Add("LFS-Authenticate", "NTLM")

	authType := GetAuthType(res)

	assert.Equal(t, NtlmAuthType, authType)
}

func TestGetAutTypeNoWwwAuthenticateLfsAuthenticateBasicNtlm(t *testing.T) {

	res := &http.Response{Header: make(http.Header)}
	res.Header.Add("LFS-Authenticate", "Basic")
	res.Header.Add("LFS-Authenticate", "NTLM")

	authType := GetAuthType(res)

	assert.Equal(t, NtlmAuthType, authType)
}

func TestGetAutTypeWwwAuthenticateBasicLfsAuthenticateNtlm(t *testing.T) {

	res := &http.Response{Header: make(http.Header)}
	res.Header.Add("WWW-Authenticate", "Basic")
	res.Header.Add("LFS-Authenticate", "NTLM")

	authType := GetAuthType(res)

	assert.Equal(t, NtlmAuthType, authType)
}
