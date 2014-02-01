package gitmedia

import (
	"github.com/pelletier/go-toml"
	"os"
	"path/filepath"
	"strings"
)

type Configuration struct {
	Endpoint  string
	gitConfig map[string]string
	remotes   []string
}

var config *Configuration

// Config gets the git media configuration for the current repository.  It
// reads .gitmedia, which is a toml file.
//
// https://github.com/mojombo/toml
func Config() *Configuration {
	if config == nil {
		config = &Configuration{}
		readToml(config)
	}

	return config
}

func (c *Configuration) RemoteEndpoint(remote string) string {
	if url, ok := c.GitConfig("remote." + remote + ".mediaUrl"); ok {
		return url
	}

	if url, ok := c.GitConfig("remote." + remote + ".url"); ok {
		return url + ".git/info/media"
	}

	return "<unknown>"
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
			uniqRemotes[keyParts[1]] = true
		}
	}

	c.remotes = make([]string, len(uniqRemotes))
	i := 0
	for remote, _ := range uniqRemotes {
		c.remotes[i] = remote
		i += 1
	}
}

func readToml(config *Configuration) {
	tomlPath := filepath.Join(LocalWorkingDir, ".gitmedia")
	stat, _ := os.Stat(tomlPath)
	if stat != nil {
		readTomlFile(tomlPath, config)
	}
}

func readTomlFile(path string, config *Configuration) {
	tomlConfig, err := toml.LoadFile(path)
	if err != nil {
		Panic(err, "Error reading TOML file: %s", path)
	}

	if endpoint, ok := tomlConfig.Get("endpoint").(string); ok {
		config.Endpoint = endpoint
	}
}
