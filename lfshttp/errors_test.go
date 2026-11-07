package lfshttp

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorMessagesRedactCredentials(t *testing.T) {
	u, err := url.Parse("https://user:Secr3t@example.com/repo.git/info/lfs/objects/batch")
	require.NoError(t, err)
	req := &http.Request{Method: "POST", URL: u}

	for _, code := range []int{400, 401, 403, 404, 422, 429, 500, 501, 502, 507, 509} {
		res := &http.Response{StatusCode: code, Request: req}
		msg := defaultError(res).Error()

		assert.NotContains(t, msg, "Secr3t", "status %d leaked the password: %s", code, msg)
		assert.NotContains(t, msg, "user:Secr3t", "status %d leaked the credentials: %s", code, msg)
		assert.Contains(t, msg, "https://example.com/repo.git/info/lfs/objects/batch",
			"status %d lost the URL: %s", code, msg)

		msg = NewStatusCodeError(res).Error()
		assert.NotContains(t, msg, "Secr3t", "status %d leaked the password: %s", code, msg)
		assert.NotContains(t, msg, "user:Secr3t", "status %d leaked the credentials: %s", code, msg)
	}
}

func TestRedactedURL(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"https://user:Secr3t@example.com/r.git", "https://example.com/r.git"},
		{"https://token@example.com/r.git", "https://example.com/r.git"},
		{"https://example.com/r.git?x=1", "https://example.com/r.git?x=1"},
		{"http://127.0.0.1:8052/r.git", "http://127.0.0.1:8052/r.git"},
		{"/local/path", "/local/path"},
	} {
		u, err := url.Parse(tc.in)
		require.NoError(t, err)
		assert.Equal(t, tc.want, RedactedURL(u))
	}
	assert.Equal(t, "", RedactedURL(nil))
}
