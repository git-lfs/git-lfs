package lfs

import (
	"net/http"
	"net/url"
	// "os"
	"testing"

	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

// --- FAIL: TestProxyFromEnv (0.00s)
// panic: runtime error: invalid memory address or nil pointer dereference [recovered]
// panic: runtime error: invalid memory address or nil pointer dereference
//
// func TestProxyFromEnv(t *testing.T) {
// 	os.Setenv("HTTPS_PROXY", "https://proxy-from-env:8080")
//
// 	u, err := url.Parse("http://some-host.com:123/foo/bar")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req := &http.Request{
// 		URL:    u,
// 		Header: http.Header{},
// 	}
// 	proxyURL, err := proxyFromGitConfigOrEnvironment(req)
//
// 	assert.Equal(t, "proxy-from-env:8080", proxyURL.Host)
// 	assert.Equal(t, nil, err)
// }

func TestProxyFromGitConfig(t *testing.T) {
	oldGitConfig := Config.gitConfig
	defer func() {
		Config.gitConfig = oldGitConfig
	}()
	Config.gitConfig = map[string]string{"http.proxy": "https://proxy-from-git-config:8080"}

	u, err := url.Parse("http://some-host.com:123/foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	req := &http.Request{
		URL:    u,
		Header: http.Header{},
	}
	proxyURL, err := proxyFromGitConfigOrEnvironment(req)

	assert.Equal(t, "proxy-from-git-config:8080", proxyURL.Host)
	assert.Equal(t, nil, err)
}

// --- FAIL: TestProxyIsNil (0.00s)
// panic: runtime error: invalid memory address or nil pointer dereference [recovered]
// panic: runtime error: invalid memory address or nil pointer dereference
//
// func TestProxyIsNil(t *testing.T) {
// 	u, err := url.Parse("http://some-host.com:123/foo/bar")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req := &http.Request{
// 		URL:    u,
// 		Header: http.Header{},
// 	}
// 	proxyURL, err := proxyFromGitConfigOrEnvironment(req)
//
// 	assert.Equal(t, nil, proxyURL.Host)
// 	assert.Equal(t, nil, err)
// }
