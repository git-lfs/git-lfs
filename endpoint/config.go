package endpoint

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type Config struct {
	aliases map[string]string
	aliasMu sync.Mutex
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

func NewConfig(git env) *Config {
	c := &Config{
		aliases: make(map[string]string),
	}

	if git != nil {
		initAliases(c, git)
	}

	return c
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
	All() map[string]string
}
