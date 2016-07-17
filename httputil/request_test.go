package httputil

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetAutTypeNoAuthHeaders(t *testing.T) {
	headers := map[string][]string{}

	authType := GetAuthType(httpResponseFromHeaders(headers))

	assert.Equal(t, BasicAuthType, authType)
}

func TestGetAutTypewwAuthenticateBasicNtlmBearer(t *testing.T) {
	headers := map[string][]string{
		"WWW-Authenticate": {"Basic", "NTLM", "Bearer"},
	}

	authType := GetAuthType(httpResponseFromHeaders(headers))

	assert.Equal(t, NtlmAuthType, authType)
}

func TestGetAutTypeNoWwwAuthenticateLfsAuthenticateNtlm(t *testing.T) {
	headers := map[string][]string{
		"LFS-Authenticate": {"Basic", "NTLM", "Bearer"},
	}

	authType := GetAuthType(httpResponseFromHeaders(headers))

	assert.Equal(t, NtlmAuthType, authType)
}

func TestGetAutTypeNoWwwAuthenticateLfsAuthenticateBasicNtlm(t *testing.T) {
	headers := map[string][]string{
		"LFS-Authenticate": {"Basic", "Ntlm"},
	}

	authType := GetAuthType(httpResponseFromHeaders(headers))

	assert.Equal(t, NtlmAuthType, authType)
}

func TestGetAutTypeWwwAuthenticateBasicNegotiate(t *testing.T) {
	headers := map[string][]string{
		"Www-Authenticate": {"Basic", "Negotiate"},
	}

	authType := GetAuthType(httpResponseFromHeaders(headers))

	assert.Equal(t, NtlmAuthType, authType)
}

func TestGetAutTypeWwwAuthenticateBasicLfsAuthenticateNtlm(t *testing.T) {
	headers := map[string][]string{
		"WWW-Authenticate": {"Basic"},
		"LFS-Authenticate": {"Ntlm"},
	}

	authType := GetAuthType(httpResponseFromHeaders(headers))

	assert.Equal(t, NtlmAuthType, authType)
}

func httpResponseFromHeaders(headers map[string][]string) *http.Response {
	res := &http.Response{Header: make(http.Header)}
	for key, values := range headers {
		for _, val := range values {
			res.Header.Add(key, val)
		}
	}
	return res
}
