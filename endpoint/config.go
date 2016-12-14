package endpoint

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
)

type Config struct {
	gitProtocol string
	aliases     map[string]string
	aliasMu     sync.Mutex
}

func NewConfig(git env) *Config {
	c := &Config{
		gitProtocol: "https",
		aliases:     make(map[string]string),
	}

	if git != nil {
		if v, ok := git.Get("lfs.gitprotocol"); ok {
			c.gitProtocol = v
		}
		initAliases(c, git)
	}

	return c
}

func (c *Config) NewEndpointFromCloneURL(rawurl string) Endpoint {
	e := c.NewEndpoint(rawurl)
	if e.Url == UrlUnknown {
		return e
	}

	if strings.HasSuffix(rawurl, "/") {
		e.Url = rawurl[0 : len(rawurl)-1]
	}

	// When using main remote URL for HTTP, append info/lfs
	if path.Ext(e.Url) == ".git" {
		e.Url += "/info/lfs"
	} else {
		e.Url += ".git/info/lfs"
	}

	return e
}

func (c *Config) NewEndpoint(rawurl string) Endpoint {
	rawurl = c.ReplaceUrlAlias(rawurl)
	u, err := url.Parse(rawurl)
	if err != nil {
		return endpointFromBareSshUrl(rawurl)
	}

	switch u.Scheme {
	case "ssh":
		return endpointFromSshUrl(u)
	case "http", "https":
		return endpointFromHttpUrl(u)
	case "git":
		return endpointFromGitUrl(u, c)
	case "":
		return endpointFromBareSshUrl(u.String())
	default:
		// Just passthrough to preserve
		return Endpoint{Url: rawurl}
	}
}

func (c *Config) GitProtocol() string {
	return c.gitProtocol
}

func (c *Config) ReplaceUrlAlias(rawurl string) string {
	c.aliasMu.Lock()
	defer c.aliasMu.Unlock()

	var longestalias string
	for alias, _ := range c.aliases {
		if !strings.HasPrefix(rawurl, alias) {
			continue
		}

		if longestalias < alias {
			longestalias = alias
		}
	}

	if len(longestalias) > 0 {
		return c.aliases[longestalias] + rawurl[len(longestalias):]
	}

	return rawurl
}

func initAliases(c *Config, git env) {
	prefix := "url."
	suffix := ".insteadof"
	for gitkey, gitval := range git.All() {
		if !(strings.HasPrefix(gitkey, prefix) && strings.HasSuffix(gitkey, suffix)) {
			continue
		}
		if _, ok := c.aliases[gitval]; ok {
			fmt.Fprintf(os.Stderr, "WARNING: Multiple 'url.*.insteadof' keys with the same alias: %q\n", gitval)
		}
		c.aliases[gitval] = gitkey[len(prefix) : len(gitkey)-len(suffix)]
	}
}

type env interface {
	Get(string) (string, bool)
	All() map[string]string
}
