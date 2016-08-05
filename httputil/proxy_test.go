package httputil

import (
	"net/http"
	"testing"

	"github.com/github/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestProxyFromGitConfig(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"http.proxy": "https://proxy-from-git-config:8080",
		},
		Os: map[string]string{
			"HTTPS_PROXY": "https://proxy-from-env:8080",
		},
	})

	req, err := http.NewRequest("GET", "https://some-host.com:123/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxyURL, err := ProxyFromGitConfigOrEnvironment(cfg)(req)

	assert.Equal(t, "proxy-from-git-config:8080", proxyURL.Host)
	assert.Nil(t, err)
}

func TestHttpProxyFromGitConfig(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"http.proxy": "http://proxy-from-git-config:8080",
		},
		Os: map[string]string{
			"HTTPS_PROXY": "https://proxy-from-env:8080",
		},
	})

	req, err := http.NewRequest("GET", "https://some-host.com:123/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxyURL, err := ProxyFromGitConfigOrEnvironment(cfg)(req)

	assert.Equal(t, "proxy-from-env:8080", proxyURL.Host)
	assert.Nil(t, err)
}

func TestProxyFromEnvironment(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Os: map[string]string{
			"HTTPS_PROXY": "https://proxy-from-env:8080",
		},
	})

	req, err := http.NewRequest("GET", "https://some-host.com:123/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxyURL, err := ProxyFromGitConfigOrEnvironment(cfg)(req)

	assert.Equal(t, "proxy-from-env:8080", proxyURL.Host)
	assert.Nil(t, err)
}

func TestProxyIsNil(t *testing.T) {
	cfg := config.New()

	req, err := http.NewRequest("GET", "http://some-host.com:123/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxyURL, err := ProxyFromGitConfigOrEnvironment(cfg)(req)

	assert.Nil(t, proxyURL)
	assert.Nil(t, err)
}

func TestProxyNoProxy(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string]string{
			"http.proxy": "https://proxy-from-git-config:8080",
		},
		Os: map[string]string{
			"NO_PROXY": "some-host",
		},
	})

	req, err := http.NewRequest("GET", "https://some-host:8080", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxyUrl, err := ProxyFromGitConfigOrEnvironment(cfg)(req)

	assert.Nil(t, proxyUrl)
	assert.Nil(t, err)
}
