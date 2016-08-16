package httputil

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"fmt"

	"github.com/github/git-lfs/config"
)

// Logic is copied, with small changes, from "net/http".ProxyFromEnvironment in the go std lib.
func ProxyFromGitConfigOrEnvironment(c *config.Configuration) func(req *http.Request) (*url.URL, error) {
	var https_proxy string
	http_proxy, _ := c.Git.Get("http.proxy")
	if strings.HasPrefix(http_proxy, "https://") {
		https_proxy = http_proxy
	}

	if len(https_proxy) == 0 {
		https_proxy, _ = c.Os.Get("HTTPS_PROXY")
	}

	if len(https_proxy) == 0 {
		https_proxy, _ = c.Os.Get("https_proxy")
	}

	if len(http_proxy) == 0 {
		http_proxy, _ = c.Os.Get("HTTP_PROXY")
	}

	if len(http_proxy) == 0 {
		http_proxy, _ = c.Os.Get("http_proxy")
	}

	no_proxy, _ := c.Os.Get("NO_PROXY")
	if len(no_proxy) == 0 {
		no_proxy, _ = c.Os.Get("no_proxy")
	}

	return func(req *http.Request) (*url.URL, error) {
		var proxy string
		if req.URL.Scheme == "https" {
			proxy = https_proxy
		}

		if len(proxy) == 0 {
			proxy = http_proxy
		}

		if len(proxy) == 0 {
			return nil, nil
		}

		if !useProxy(no_proxy, canonicalAddr(req.URL)) {
			return nil, nil
		}

		proxyURL, err := url.Parse(proxy)
		if err != nil || !strings.HasPrefix(proxyURL.Scheme, "http") {
			// proxy was bogus. Try prepending "http://" to it and
			// see if that parses correctly. If not, we fall
			// through and complain about the original one.
			if proxyURL, err := url.Parse("http://" + proxy); err == nil {
				return proxyURL, nil
			}
		}
		if err != nil {
			return nil, fmt.Errorf("invalid proxy address: %q: %v", proxy, err)
		}
		return proxyURL, nil
	}
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
// according to the NO_PROXY or no_proxy environment variable.
// addr is always a canonicalAddr with a host and port.
// Copied from "net/http".ProxyFromEnvironment in the go std lib.
func useProxy(no_proxy, addr string) bool {
	if len(addr) == 0 {
		return true
	}

	if no_proxy == "*" {
		return false
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return false
		}
	}

	addr = strings.ToLower(strings.TrimSpace(addr))
	if hasPort(addr) {
		addr = addr[:strings.LastIndex(addr, ":")]
	}

	for _, p := range strings.Split(no_proxy, ",") {
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
			// no_proxy ".foo.com" matches "bar.foo.com" or "foo.com"
			return false
		}
		if p[0] != '.' && strings.HasSuffix(addr, p) && addr[len(addr)-len(p)-1] == '.' {
			// no_proxy "foo.com" matches "bar.foo.com"
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
