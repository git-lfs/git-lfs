package gitmedia

import (
	"strings"
)

type Configuration struct {
	gitConfig map[string]string
	remotes   []string
}

var Config = &Configuration{}

func (c *Configuration) Endpoint() string {
	if url, ok := c.GitConfig("media.url"); ok {
		return url
	}

	return c.RemoteEndpoint("origin")
}

func (c *Configuration) RemoteEndpoint(remote string) string {
	if url, ok := c.GitConfig("remote." + remote + ".media"); ok {
		return url
	}

	if url, ok := c.GitConfig("remote." + remote + ".url"); ok {
		return url + ".git/info/media"
	}

	return ""
}

func (c *Configuration) Remotes() []string {
	if c.remotes == nil {
		c.loadGitConfig()
	}
	return c.remotes
}

func (c *Configuration) GitConfig(key string) (string, bool) {
	if c.gitConfig == nil {
		c.loadGitConfig()
	}
	value, ok := c.gitConfig[key]
	return value, ok
}

func (c *Configuration) SetConfig(key, value string) {
	if c.gitConfig == nil {
		c.loadGitConfig()
	}
	c.gitConfig[key] = value
}

func (c *Configuration) loadGitConfig() {
	uniqRemotes := make(map[string]bool)

	c.gitConfig = make(map[string]string)
	lines := strings.Split(SimpleExec("git", "config", "-l"), "\n")
	for _, line := range lines {
		pieces := strings.SplitN(line, "=", 2)
		key := pieces[0]
		c.gitConfig[key] = pieces[1]

		keyParts := strings.Split(key, ".")
		if len(keyParts) > 1 && keyParts[0] == "remote" {
			remote := keyParts[1]
			uniqRemotes[remote] = remote == "origin"
		}
	}

	c.remotes = make([]string, 0, len(uniqRemotes))
	for remote, isOrigin := range uniqRemotes {
		if isOrigin {
			continue
		}
		c.remotes = append(c.remotes, remote)
	}
}
