package lfs

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/httputil"
	"github.com/stretchr/testify/assert"
)

func TestProxyFromGitConfig(t *testing.T) {
	cfg := config.NewFromValues(map[string]string{
		"http.proxy": "https://proxy-from-git-config:8080",
	})

	u, err := url.Parse("http://some-host.com:123/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		URL:    u,
		Header: http.Header{},
	}

	proxyURL, err := httputil.ProxyFromGitConfigOrEnvironment(cfg)(req)

	assert.Equal(t, "proxy-from-git-config:8080", proxyURL.Host)
	assert.Equal(t, nil, err)
}

func TestProxyIsNil(t *testing.T) {
	cfg := config.NewConfig()
	u, err := url.Parse("http://some-host.com:123/foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	req := &http.Request{
		URL:    u,
		Header: http.Header{},
	}

	proxyURL, err := httputil.ProxyFromGitConfigOrEnvironment(cfg)(req)

	assert.Nil(t, proxyURL)
	assert.Nil(t, err)
}
