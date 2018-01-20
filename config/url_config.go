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
	if git == nil {
		git = EnvironmentOf(make(mapFetcher))
	}

	return &URLConfig{
		git: git,
	}
}

// Get retrieves a `http.{url}.{key}` for the given key and urls, following the
// rules in https://git-scm.com/docs/git-config#git-config-httplturlgt.
// The value for `http.{key}` is returned as a fallback if no config keys are
// set for the given urls.
func (c *URLConfig) Get(prefix, rawurl, key string) (string, bool) {
	if c == nil {
		return "", false
	}

	key = strings.ToLower(key)
	prefix = strings.ToLower(prefix)
	if v := c.getAll(prefix, rawurl, key); len(v) > 0 {
		return v[len(v)-1], true
	}
	return c.git.Get(strings.Join([]string{prefix, key}, "."))
}

func (c *URLConfig) GetAll(prefix, rawurl, key string) []string {
	if c == nil {
		return nil
	}

	key = strings.ToLower(key)
	prefix = strings.ToLower(prefix)
	if v := c.getAll(prefix, rawurl, key); len(v) > 0 {
		return v
	}
	return c.git.GetAll(strings.Join([]string{prefix, key}, "."))
}

func (c *URLConfig) Bool(prefix, rawurl, key string, def bool) bool {
	s, _ := c.Get(prefix, rawurl, key)
	return Bool(s, def)
}

func (c *URLConfig) getAll(prefix, rawurl, key string) []string {
	hosts, paths := c.hostsAndPaths(rawurl)

	for i := len(paths); i > 0; i-- {
		for _, host := range hosts {
			path := strings.Join(paths[:i], slash)
			if v := c.git.GetAll(fmt.Sprintf("%s.%s/%s.%s", prefix, host, path, key)); len(v) > 0 {
				return v
			}
			if v := c.git.GetAll(fmt.Sprintf("%s.%s/%s/.%s", prefix, host, path, key)); len(v) > 0 {
				return v
			}

			if isDefaultLFSUrl(path, paths, i) {
				path = path[0 : len(path)-4]
				if v := c.git.GetAll(fmt.Sprintf("%s.%s/%s.%s", prefix, host, path, key)); len(v) > 0 {
					return v
				}
			}
		}
	}

	for _, host := range hosts {
		if v := c.git.GetAll(fmt.Sprintf("%s.%s.%s", prefix, host, key)); len(v) > 0 {
			return v
		}
		if v := c.git.GetAll(fmt.Sprintf("%s.%s/.%s", prefix, host, key)); len(v) > 0 {
			return v
		}
	}
	return nil

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
	if strings.HasSuffix(path, slash) {
		end--
	}
	return strings.Split(path[1:end], slash)
}

const (
	gitExt   = ".git"
	infoPart = "info"
	lfsPart  = "lfs"
	slash    = "/"
)

func isDefaultLFSUrl(path string, parts []string, index int) bool {
	if len(path) < 5 {
		return false // shorter than ".git"
	}

	if !strings.HasSuffix(path, gitExt) {
		return false
	}

	if index > len(parts)-2 {
		return false
	}

	return parts[index] == infoPart && parts[index+1] == lfsPart
}
