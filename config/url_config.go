package config

import (
	"fmt"
	"net/url"
	"strings"
)

type URLConfig struct {
	git Environment
}

func NewURLConfig(git Environment) *URLConfig {
	return &URLConfig{
		git: git,
	}
}

// Get retrieves a `http.{url}.{key}` for the given key and urls, following the
// rules in https://git-scm.com/docs/git-config#git-config-httplturlgt.
// The value for `http.{key}` is returned as a fallback if no config keys are
// set for the given urls.
func (c *URLConfig) Get(prefix, key string, rawurl string) (string, bool) {
	key = strings.ToLower(key)
	prefix = strings.ToLower(prefix)
	if v, ok := c.get(key, rawurl); ok {
		return v, ok
	}
	return c.git.Get(strings.Join([]string{prefix, key}, "."))
}

func (c *URLConfig) get(key, rawurl string) (string, bool) {
	hosts, paths := c.hostsAndPaths(rawurl)

	for i := len(paths); i > 0; i-- {
		for _, host := range hosts {
			path := strings.Join(paths[:i], "/")
			if v, ok := c.git.Get(fmt.Sprintf("http.%s/%s.%s", host, path, key)); ok {
				return v, ok
			}
			if v, ok := c.git.Get(fmt.Sprintf("http.%s/%s/.%s", host, path, key)); ok {
				return v, ok
			}
		}
	}

	for _, host := range hosts {
		if v, ok := c.git.Get(fmt.Sprintf("http.%s.%s", host, key)); ok {
			return v, ok
		}
		if v, ok := c.git.Get(fmt.Sprintf("http.%s/.%s", host, key)); ok {
			return v, ok
		}
	}
	return "", false

}

func (c *URLConfig) hostsAndPaths(rawurl string) (hosts, paths []string) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, nil
	}

	return c.hosts(u), c.paths(u.Path)
}

func (c *URLConfig) hosts(u *url.URL) []string {
	hosts := make([]string, 0, 1)

	if u.User != nil {
		hosts = append(hosts, fmt.Sprintf("%s://%s@%s", u.Scheme, u.User.Username(), u.Host))
	}
	hosts = append(hosts, fmt.Sprintf("%s://%s", u.Scheme, u.Host))

	return hosts
}

func (c *URLConfig) paths(path string) []string {
	pLen := len(path)
	if pLen <= 2 {
		return nil
	}

	end := pLen
	if strings.HasSuffix(path, "/") {
		end -= 1
	}
	return strings.Split(path[1:end], "/")
}
