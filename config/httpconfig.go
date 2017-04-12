package config

import (
	"fmt"
	"net/url"
	"strings"
)

type HTTPConfig struct {
	git Environment
}

func NewHTTPConfig(git Environment) *HTTPConfig {
	return &HTTPConfig{
		git: git,
	}
}

// Get retrieves a `http.{url}.{key}` for the given key and urls, following the
// rules in https://git-scm.com/docs/git-config#git-config-httplturlgt.
// The value for `http.{key}` is returned as a fallback if no config keys are
// set for the given urls.
func (c *HTTPConfig) Get(key string, rawurl string) (string, bool) {
	key = strings.ToLower(key)
	if v, ok := c.get(key, rawurl); ok {
		return v, ok
	}
	return c.git.Get(fmt.Sprintf("http.%s", key))
}

func (c *HTTPConfig) get(key, rawurl string) (string, bool) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", false
	}

	hosts := make([]string, 0)
	if u.User != nil {
		hosts = append(hosts, fmt.Sprintf("%s://%s@%s", u.Scheme, u.User.Username(), u.Host))
	}
	hosts = append(hosts, fmt.Sprintf("%s://%s", u.Scheme, u.Host))

	pLen := len(u.Path)
	if pLen > 2 {
		end := pLen
		if strings.HasSuffix(u.Path, "/") {
			end -= 1
		}

		paths := strings.Split(u.Path[1:end], "/")
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
