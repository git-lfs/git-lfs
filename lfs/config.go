package lfs

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/github/git-lfs/git"
)

type Configuration struct {
	CurrentRemote         string
	gitConfig             map[string]string
	remotes               []string
	httpClient            *http.Client
	redirectingHttpClient *http.Client
	isTracingHttp         bool
	loading               sync.Mutex
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
		isTracingHttp: len(os.Getenv("GIT_CURL_VERBOSE")) > 0,
	}
	return c
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

func (c *Configuration) Endpoint() Endpoint {
	e := c.endpoint()

	if u, err := url.Parse(e.Url); err == nil && u.User != nil {
		fmt.Fprintln(os.Stderr, "warning: configured LFS endpoint contains credentials")
	}

	return e
}

func (c *Configuration) endpoint() Endpoint {
	if url, ok := c.GitConfig("lfs.url"); ok {
		return Endpoint{Url: url}
	}

	if len(c.CurrentRemote) > 0 && c.CurrentRemote != defaultRemote {
		if e := c.RemoteEndpoint(c.CurrentRemote); len(e.Url) > 0 {
			return e
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
		endpoint := Endpoint{Url: url}

		if !httpPrefixRe.MatchString(url) {
			pieces := strings.SplitN(url, ":", 2)
			hostPieces := strings.SplitN(pieces[0], "@", 2)
			if len(hostPieces) < 2 {
				endpoint.Url = "<unknown>"
				return endpoint
			}

			endpoint.SshUserAndHost = pieces[0]
			endpoint.SshPath = pieces[1]
			endpoint.Url = fmt.Sprintf("https://%s/%s", hostPieces[1], pieces[1])
		}

		if path.Ext(url) == ".git" {
			endpoint.Url += "/info/lfs"
		} else {
			endpoint.Url += ".git/info/lfs"
		}

		return endpoint
	}

	return Endpoint{}
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

type AltConfig struct {
	Remote map[string]*struct {
		Media string
	}

	Media struct {
		Url string
	}
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

	fileOutput, err := git.Config.ListFromFile()
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
