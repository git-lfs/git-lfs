package gitmedia

import (
	"fmt"
	"github.com/github/git-media/git"
	"os"
	"path"
	"regexp"
	"strings"
)

type Configuration struct {
	gitConfig map[string]string
	remotes   []string
}

var (
	Config       = &Configuration{}
	httpPrefixRe = regexp.MustCompile("\\Ahttps?://")
)

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
		if !httpPrefixRe.MatchString(url) {
			pieces := strings.SplitN(url, ":", 2)
			hostPieces := strings.SplitN(pieces[0], "@", 2)
			if len(hostPieces) < 2 {
				return "unknown"
			}
			url = fmt.Sprintf("https://%s/%s", hostPieces[1], pieces[1])
		}

		if path.Ext(url) == ".git" {
			return url + "/info/media"
		}
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

type AltConfig struct {
	Remote map[string]*struct {
		Media string
	}

	Media struct {
		Url string
	}
}

func (c *Configuration) loadGitConfig() {
	uniqRemotes := make(map[string]bool)

	c.gitConfig = make(map[string]string)

	var output string
	listOutput, err := git.Config.List()
	if err != nil {
		panic(fmt.Errorf("Error listing git config: %s", err))
	}

	fileOutput, err := git.Config.ListFromFile()
	if err != nil {
		panic(fmt.Errorf("Error listing git config from file: %s", err))
	}

	output = listOutput + "\n" + fileOutput

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		pieces := strings.SplitN(line, "=", 2)
		if len(pieces) < 2 {
			continue
		}
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

func configFileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}
