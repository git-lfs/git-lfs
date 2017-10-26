package lfsapi

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/git-lfs/git-lfs/config"

	"fmt"
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

		if !useProxy(noProxy, canonicalAddr(req.URL)) {
			return nil, nil
		}

		proxyURL, err := url.Parse(proxy)
		if err != nil || !strings.HasPrefix(proxyURL.Scheme, "http") {
			// proxy was bogus. Try prepending "http://" to it and
			// see if that parses correctly. If not, we fall
			// through and complain about the original one.
			if httpProxyURL, httpErr := url.Parse("http://" + proxy); httpErr == nil {
				return httpProxyURL, nil
			}
		}
		if err != nil {
			return nil, fmt.Errorf("invalid proxy address: %q: %v", proxy, err)
		}
		return proxyURL, nil
	}
}

func getProxyServers(u *url.URL, urlCfg *config.URLConfig, osEnv config.Environment) (httpsProxy string, httpProxy string, noProxy string) {
	if urlCfg != nil {
		httpProxy, _ = urlCfg.Get("http", u.String(), "proxy")
		if strings.HasPrefix(httpProxy, "https://") {
			httpsProxy = httpProxy
		}
	}

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

	noProxy, _ = osEnv.Get("NO_PROXY")
	if len(noProxy) == 0 {
		noProxy, _ = osEnv.Get("no_proxy")
	}

	return
}

// canonicalAddr returns url.Host but always with a ":port" suffix
// Copied from "net/http".ProxyFromEnvironment in the go std lib.
func canonicalAddr(url *url.URL) string {
	addr := url.Host
	if !hasPort(addr) {
		return addr + ":" + portMap[url.Scheme]
	}
	return addr
}

// useProxy reports whether requests to addr should use a proxy,
// according to the noProxy or noProxy environment variable.
// addr is always a canonicalAddr with a host and port.
// Copied from "net/http".ProxyFromEnvironment in the go std lib
// and adapted to allow proxy usage even for localhost.
func useProxy(noProxy, addr string) bool {
	if len(addr) == 0 {
		return true
	}

	if noProxy == "*" {
		return false
	}

	addr = strings.ToLower(strings.TrimSpace(addr))
	if hasPort(addr) {
		addr = addr[:strings.LastIndex(addr, ":")]
	}

	for _, p := range strings.Split(noProxy, ",") {
		p = strings.ToLower(strings.TrimSpace(p))
		if len(p) == 0 {
			continue
		}
		if hasPort(p) {
			p = p[:strings.LastIndex(p, ":")]
		}
		if addr == p {
			return false
		}
		if p[0] == '.' && (strings.HasSuffix(addr, p) || addr == p[1:]) {
			// noProxy ".foo.com" matches "bar.foo.com" or "foo.com"
			return false
		}
		if p[0] != '.' && strings.HasSuffix(addr, p) && addr[len(addr)-len(p)-1] == '.' {
			// noProxy "foo.com" matches "bar.foo.com"
			return false
		}
	}
	return true
}

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
// Copied from "net/http".ProxyFromEnvironment in the go std lib.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

var (
	portMap = map[string]string{
		"http":  "80",
		"https": "443",
	}
)
