package endpoint

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/git"
)

const defaultRemote = "origin"

type Config struct {
	git         env
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
		c.git = git
		if v, ok := git.Get("lfs.gitprotocol"); ok {
			c.gitProtocol = v
		}
		initAliases(c, git)
	}

	return c
}

func (c *Config) Endpoint(operation, remote string) Endpoint {
	if c.git == nil {
		return Endpoint{}
	}

	if operation == "upload" {
		if url, ok := c.git.Get("lfs.pushurl"); ok {
			return c.NewEndpoint(url)
		}
	}

	if url, ok := c.git.Get("lfs.url"); ok {
		return c.NewEndpoint(url)
	}

	if len(remote) > 0 && remote != defaultRemote {
		if e := c.RemoteEndpoint(operation, remote); len(e.Url) > 0 {
			return e
		}
	}

	return c.RemoteEndpoint(operation, defaultRemote)
}

func (c *Config) RemoteEndpoint(operation, remote string) Endpoint {
	if c.git == nil {
		return Endpoint{}
	}

	if len(remote) == 0 {
		remote = defaultRemote
	}

	// Support separate push URL if specified and pushing
	if operation == "upload" {
		if url, ok := c.git.Get("remote." + remote + ".lfspushurl"); ok {
			return c.NewEndpoint(url)
		}
	}
	if url, ok := c.git.Get("remote." + remote + ".lfsurl"); ok {
		return c.NewEndpoint(url)
	}

	// finally fall back on git remote url (also supports pushurl)
	if url := c.GitRemoteURL(remote, operation == "upload"); url != "" {
		return c.NewEndpointFromCloneURL(url)
	}

	return Endpoint{}
}

func (c *Config) GitRemoteURL(remote string, forpush bool) string {
	if c.git != nil {
		if forpush {
			if u, ok := c.git.Get("remote." + remote + ".pushurl"); ok {
				return u
			}
		}

		if u, ok := c.git.Get("remote." + remote + ".url"); ok {
			return u
		}
	}

	if err := git.ValidateRemote(remote); err == nil {
		return remote
	}

	return ""
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
