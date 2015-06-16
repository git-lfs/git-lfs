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
	gitConfig             map[string]string
	remotes               []string
	httpClient            *HttpClient
	redirectingHttpClient *http.Client
	isTracingHttp         bool
	isLoggingStats        bool
	loading               sync.Mutex
}

type Endpoint struct {
	Url            string
	SshUserAndHost string
	SshPath        string
	SshPort        string
}

var (
	Config        = NewConfig()
	httpPrefixRe  = regexp.MustCompile("\\Ahttps?://")
	defaultRemote = "origin"
)

func NewConfig() *Configuration {
	c := &Configuration{
		CurrentRemote:  defaultRemote,
		isTracingHttp:  len(os.Getenv("GIT_CURL_VERBOSE")) > 0,
		isLoggingStats: len(os.Getenv("GIT_LOG_STATS")) > 0,
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

func NewEndpoint(urlstr string) Endpoint {
	endpoint := Endpoint{Url: urlstr}
	// Check for SSH URLs
	// Support ssh://user@host.com/path/to/repo and user@host.com:path/to/repo
	u, err := url.Parse(urlstr)
	if err == nil {
		if u.Scheme == "" && u.Path != "" {
			// This might be a bare SSH URL
			// Turn it into ssh:// for simplicity of extraction later
			parts := strings.Split(u.Path, ":")
			if len(parts) > 1 {
				// Treat presence of ':' as a bare URL
				var newPath string
				if len(parts) > 2 { // port included; really should only ever be 3 parts
					newPath = fmt.Sprintf("%v:%v", parts[0], strings.Join(parts[1:], "/"))
				} else {
					newPath = strings.Join(parts, "/")
				}
				newUrlStr := fmt.Sprintf("ssh://%v", newPath)
				newu, err := url.Parse(newUrlStr)
				if err == nil {
					u = newu
				}
			}
		}
		// Now extract the SSH parts from sanitised u
		if u.Scheme == "ssh" {
			var host string
			// Pull out port now, we need it separately for SSH
			regex := regexp.MustCompile(`^([^\:]+)(?:\:(\d+))?$`)
			if match := regex.FindStringSubmatch(u.Host); match != nil {
				host = match[1]
				if u.User != nil && u.User.Username() != "" {
					endpoint.SshUserAndHost = fmt.Sprintf("%s@%s", u.User.Username(), host)
				} else {
					endpoint.SshUserAndHost = host
				}
				if len(match) > 2 {
					endpoint.SshPort = match[2]
				}
			} else {
				endpoint.Url = "<unknown>"
				return endpoint
			}

			// u.Path includes a preceding '/', strip off manually
			// rooted paths in the URL will be '//path/to/blah'
			// this is just how Go's URL parsing works
			if strings.HasPrefix(u.Path, "/") {
				endpoint.SshPath = u.Path[1:]
			} else {
				endpoint.SshPath = u.Path
			}
			// Fallback URL for using HTTPS while still using SSH for git
			// u.Host includes host & port so can't use SSH port
			endpoint.Url = fmt.Sprintf("https://%s%s", host, u.Path)
		}
	} else {
		endpoint.Url = "<unknown>"
	}
	return endpoint

}
func (c *Configuration) Endpoint() Endpoint {
	if url, ok := c.GitConfig("lfs.url"); ok {
		return NewEndpoint(url)
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
		return NewEndpoint(url)
	}

	if url, ok := c.GitConfig("remote." + remote + ".url"); ok {
		endpoint := NewEndpoint(url)

		// When using main remote URL for HTTP, append info/lfs
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
