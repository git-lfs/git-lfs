package lfs

import (
	"net/http"
	"testing"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/httputil"
	"github.com/stretchr/testify/assert"
)

func TestProxyFromEnvironment(t *testing.T) {
	cfg := config.NewFromValues(map[string]string{
		"http.proxy": "https://proxy-from-git-config:8080",
	})
	cfg.SetAllEnv(map[string]string{
		"HTTPS_PROXY": "https://proxy-from-env:8080",
	})

	req, err := http.NewRequest("GET", "https://some-host.com:123/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxyURL, err := httputil.ProxyFromGitConfigOrEnvironment(cfg)(req)

	assert.Equal(t, "proxy-from-env:8080", proxyURL.Host)
	assert.Equal(t, nil, err)
}

func TestProxyFromGitConfig(t *testing.T) {
	cfg := config.NewFromValues(map[string]string{
		"http.proxy": "https://proxy-from-git-config:8080",
	})

	req, err := http.NewRequest("GET", "http://some-host.com:123/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxyURL, err := httputil.ProxyFromGitConfigOrEnvironment(cfg)(req)

	assert.Equal(t, "proxy-from-git-config:8080", proxyURL.Host)
	assert.Equal(t, nil, err)
}

func TestProxyIsNil(t *testing.T) {
	cfg := config.NewConfig()

	req, err := http.NewRequest("GET", "http://some-host.com:123/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxyURL, err := httputil.ProxyFromGitConfigOrEnvironment(cfg)(req)

	assert.Nil(t, proxyURL)
	assert.Nil(t, err)
}
