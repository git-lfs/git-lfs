package lfs

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/github/git-lfs/git"
)

type Configuration struct {
	CurrentRemote         string
	httpClient            *HttpClient
	redirectingHttpClient *http.Client
	envVars               map[string]string
	isTracingHttp         bool
	isLoggingStats        bool

	loading   sync.Mutex // guards initialization of gitConfig and remotes
	gitConfig map[string]string
	remotes   []string
}

type Endpoint struct {
	Url            string
	SshUserAndHost string
	SshPath        string
}

var (
	Config        = NewConfig()
	httpPrefixRe  = regexp.MustCompile("\\Ahttps?://")
	defaultRemote = "origin"
)

func NewConfig() *Configuration {
	c := &Configuration{
		CurrentRemote: defaultRemote,
		envVars:       make(map[string]string),
	}
	c.isTracingHttp = len(c.Getenv("GIT_CURL_VERBOSE")) > 0
	c.isLoggingStats = len(c.Getenv("GIT_LOG_STATS")) > 0
	return c
}

func (c *Configuration) Getenv(key string) string {
	if i, ok := c.envVars[key]; ok {
		return i
	}

	v := os.Getenv(key)
	c.envVars[key] = v
	return v
}

func (c *Configuration) Endpoint() Endpoint {
	if url, ok := c.GitConfig("lfs.url"); ok {
		return Endpoint{Url: url}
	}

	if len(c.CurrentRemote) > 0 && c.CurrentRemote != defaultRemote {
		if endpoint := c.RemoteEndpoint(c.CurrentRemote); len(endpoint.Url) > 0 {
			return endpoint
		}
	}

	return c.RemoteEndpoint(defaultRemote)
}

func (c *Configuration) ConcurrentTransfers() int {
	uploads := 3

	if v, ok := c.GitConfig("lfs.concurrenttransfers"); ok {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			uploads = n
		}
	}

	return uploads
}

func (c *Configuration) BatchTransfer() bool {
	if v, ok := c.GitConfig("lfs.batch"); ok {
		if v == "true" || v == "" {
			return true
		}

		// Any numeric value except 0 is considered true
		if n, err := strconv.Atoi(v); err == nil && n != 0 {
			return true
		}
	}
	return false
}

func (c *Configuration) RemoteEndpoint(remote string) Endpoint {
	if len(remote) == 0 {
		remote = defaultRemote
	}

	if url, ok := c.GitConfig("remote." + remote + ".lfsurl"); ok {
		return Endpoint{Url: url}
	}

	if url, ok := c.GitConfig("remote." + remote + ".url"); ok {
		return remoteEndpointFromUrl(url)
	}

	return Endpoint{}
}

func remoteEndpointFromUrl(url string) Endpoint {
	e := Endpoint{Url: url}

	if !httpPrefixRe.MatchString(url) {
		pieces := strings.SplitN(url, ":", 2)
		hostPieces := strings.SplitN(pieces[0], "@", 2)
		if len(hostPieces) < 2 {
			e.Url = "<unknown>"
			return e
		}

		e.SshUserAndHost = pieces[0]
		e.SshPath = pieces[1]
		e.Url = fmt.Sprintf("https://%s/%s", hostPieces[1], pieces[1])
	}

	if path.Ext(url) == ".git" {
		e.Url += "/info/lfs"
	} else {
		e.Url += ".git/info/lfs"
	}

	return e
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

func (c *Configuration) ObjectUrl(oid string) (*url.URL, error) {
	return ObjectUrl(c.Endpoint(), oid)
}

func ObjectUrl(endpoint Endpoint, oid string) (*url.URL, error) {
	u, err := url.Parse(endpoint.Url)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "objects")
	if len(oid) > 0 {
		u.Path = path.Join(u.Path, oid)
	}
	return u, nil
}

func (c *Configuration) loadGitConfig() {
	c.loading.Lock()
	defer c.loading.Unlock()

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

	configFile := filepath.Join(LocalWorkingDir, ".gitconfig")
	fileOutput, err := git.Config.ListFromFile(configFile)
	if err != nil {
		panic(fmt.Errorf("Error listing git config from file: %s", err))
	}

	output = fileOutput + "\n" + listOutput

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
