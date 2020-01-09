package lfshttp

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/git-lfs/git-lfs/config"
	"golang.org/x/net/http/httpproxy"
)

// Logic is copied, with small changes, from "net/http".ProxyFromEnvironment in the go std lib.
func proxyFromClient(c *Client) func(req *http.Request) (*url.URL, error) {
	return func(req *http.Request) (*url.URL, error) {
		httpsProxy, httpProxy, noProxy := getProxyServers(req.URL, c.uc, c.osEnv)

		var proxy string
		if req.URL.Scheme == "https" {
			proxy = httpsProxy
		}

		if len(proxy) == 0 {
			proxy = httpProxy
		}

		if len(proxy) == 0 {
			return nil, nil
		}

		cfg := &httpproxy.Config{
			HTTPProxy:  proxy,
			HTTPSProxy: proxy,
			NoProxy:    noProxy,
			CGI:        false,
		}

		// We want to use the standard logic except that we want to
		// allow proxies for localhost, which the standard library does
		// not.  Since the proxy code looks only at the URL, we
		// synthesize a fake URL except that we rewrite "localhost" to
		// "127.0.0.1" for purposes of looking up the proxy.
		u := *(req.URL)
		if u.Host == "localhost" {
			u.Host = "127.0.0.1"
		}
		return cfg.ProxyFunc()(&u)
	}
}

func getProxyServers(u *url.URL, urlCfg *config.URLConfig, osEnv config.Environment) (httpsProxy string, httpProxy string, noProxy string) {
	if osEnv == nil {
		return
	}

	if len(httpsProxy) == 0 {
		httpsProxy, _ = osEnv.Get("HTTPS_PROXY")
	}

	if len(httpsProxy) == 0 {
		httpsProxy, _ = osEnv.Get("https_proxy")
	}

	if len(httpProxy) == 0 {
		httpProxy, _ = osEnv.Get("HTTP_PROXY")
	}

	if len(httpProxy) == 0 {
		httpProxy, _ = osEnv.Get("http_proxy")
	}

	if urlCfg != nil {
		gitProxy, ok := urlCfg.Get("http", u.String(), "proxy")
		if len(gitProxy) > 0 && ok {
			if strings.HasPrefix(gitProxy, "https://") {
				httpsProxy = gitProxy
			}
			httpProxy = gitProxy
		}
	}

	noProxy, _ = osEnv.Get("NO_PROXY")
	if len(noProxy) == 0 {
		noProxy, _ = osEnv.Get("no_proxy")
	}

	return
}
