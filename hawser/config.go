package hawser

import (
	"fmt"
	"github.com/hawser/git-hawser/git"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

type Configuration struct {
	CurrentRemote         string
	OutputWriter          io.Writer
	gitConfig             map[string]string
	remotes               []string
	httpClient            *http.Client
	redirectingHttpClient *http.Client
	isTracingHttp         bool
}

var (
	Config        = NewConfig()
	RedirectError = fmt.Errorf("Unexpected redirection")
	httpPrefixRe  = regexp.MustCompile("\\Ahttps?://")
	defaultRemote = "origin"
)

func NewConfig() *Configuration {
	c := &Configuration{
		CurrentRemote: defaultRemote,
		OutputWriter:  os.Stderr,
	}
	if len(os.Getenv("GIT_CURL_VERBOSE")) > 0 || len(os.Getenv("GIT_HTTP_VERBOSE")) > 0 {
		c.isTracingHttp = true
	}
	return c
}

func (c *Configuration) Endpoint() string {
	if url, ok := c.GitConfig("hawser.url"); ok {
		return url
	}

	if len(c.CurrentRemote) > 0 && c.CurrentRemote != defaultRemote {
		if endpoint := c.RemoteEndpoint(c.CurrentRemote); len(endpoint) > 0 {
			return endpoint
		}
	}

	return c.RemoteEndpoint(defaultRemote)
}

func (c *Configuration) RemoteEndpoint(remote string) string {
	if len(remote) == 0 {
		remote = defaultRemote
	}

	if url, ok := c.GitConfig("remote." + remote + ".hawser"); ok {
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
	c.loadGitConfig()
	return c.remotes
}

func (c *Configuration) GitConfig(key string) (string, bool) {
	c.loadGitConfig()
	value, ok := c.gitConfig[strings.ToLower(key)]
	return value, ok
}

func (c *Configuration) SetConfig(key, value string) {
	c.loadGitConfig()
	c.gitConfig[key] = value
}

func (c *Configuration) ObjectUrl(oid string) *url.URL {
	u, _ := url.Parse(c.Endpoint())
	u.Path = path.Join(u.Path, "objects", oid)
	return u
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
	if c.gitConfig != nil {
		return
	}

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
		key := strings.ToLower(pieces[0])
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
