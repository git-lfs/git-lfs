package lfsapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyFromGitConfig(t *testing.T) {
	c, err := NewClient(TestEnv(map[string]string{
		"HTTPS_PROXY": "https://proxy-from-env:8080",
	}), TestEnv(map[string]string{
		"http.proxy": "https://proxy-from-git-config:8080",
	}))
	require.Nil(t, err)

	req, err := http.NewRequest("GET", "https://some-host.com:123/foo/bar", nil)
	require.Nil(t, err)

	proxyURL, err := proxyFromClient(c)(req)
	assert.Equal(t, "proxy-from-git-config:8080", proxyURL.Host)
	assert.Nil(t, err)
}

func TestHttpProxyFromGitConfig(t *testing.T) {
	c, err := NewClient(TestEnv(map[string]string{
		"HTTPS_PROXY": "https://proxy-from-env:8080",
	}), TestEnv(map[string]string{
		"http.proxy": "http://proxy-from-git-config:8080",
	}))
	require.Nil(t, err)

	req, err := http.NewRequest("GET", "https://some-host.com:123/foo/bar", nil)
	require.Nil(t, err)

	proxyURL, err := proxyFromClient(c)(req)
	assert.Equal(t, "proxy-from-env:8080", proxyURL.Host)
	assert.Nil(t, err)
}

func TestProxyFromEnvironment(t *testing.T) {
	c, err := NewClient(TestEnv(map[string]string{
		"HTTPS_PROXY": "https://proxy-from-env:8080",
	}), nil)
	require.Nil(t, err)

	req, err := http.NewRequest("GET", "https://some-host.com:123/foo/bar", nil)
	require.Nil(t, err)

	proxyURL, err := proxyFromClient(c)(req)
	assert.Equal(t, "proxy-from-env:8080", proxyURL.Host)
	assert.Nil(t, err)
}

func TestProxyIsNil(t *testing.T) {
	c := &Client{}

	req, err := http.NewRequest("GET", "http://some-host.com:123/foo/bar", nil)
	require.Nil(t, err)

	proxyURL, err := proxyFromClient(c)(req)
	assert.Nil(t, proxyURL)
	assert.Nil(t, err)
}

func TestProxyNoProxy(t *testing.T) {
	c, err := NewClient(TestEnv(map[string]string{
		"NO_PROXY": "some-host",
	}), TestEnv(map[string]string{
		"http.proxy": "https://proxy-from-git-config:8080",
	}))
	require.Nil(t, err)

	req, err := http.NewRequest("GET", "https://some-host:8080", nil)
	require.Nil(t, err)

	proxyURL, err := proxyFromClient(c)(req)
	assert.Nil(t, proxyURL)
	assert.Nil(t, err)
}
