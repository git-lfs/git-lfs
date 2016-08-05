// Package config collects together all configuration settings
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/github/git-lfs/git"
	"github.com/rubyist/tracerx"
)

var (
	Config                 = New()
	ShowConfigWarnings     = false
	defaultRemote          = "origin"
	gitConfigWarningPrefix = "lfs."
)

// FetchPruneConfig collects together the config options that control fetching and pruning
type FetchPruneConfig struct {
	// The number of days prior to current date for which (local) refs other than HEAD
	// will be fetched with --recent (default 7, 0 = only fetch HEAD)
	FetchRecentRefsDays int
	// Makes the FetchRecentRefsDays option apply to remote refs from fetch source as well (default true)
	FetchRecentRefsIncludeRemotes bool
	// number of days prior to latest commit on a ref that we'll fetch previous
	// LFS changes too (default 0 = only fetch at ref)
	FetchRecentCommitsDays int
	// Whether to always fetch recent even without --recent
	FetchRecentAlways bool
	// Number of days added to FetchRecent*; data outside combined window will be
	// deleted when prune is run. (default 3)
	PruneOffsetDays int
	// Always verify with remote before pruning
	PruneVerifyRemoteAlways bool
	// Name of remote to check for unpushed and verify checks
	PruneRemoteName string
}

type Configuration struct {
	// Os provides a `*Environment` used to access to the system's
	// environment through os.Getenv. It is the point of entry for all
	// system environment configuration.
	Os *Environment

	// Git provides a `*Environment` used to access to the various levels of
	// `.gitconfig`'s. It is the point of entry for all Git environment
	// configuration.
	Git       *Environment
	gitConfig map[string]string

	CurrentRemote   string
	NtlmSession     ntlm.ClientSession
	envVars         map[string]string
	envVarsMutex    sync.Mutex
	IsTracingHttp   bool
	IsDebuggingHttp bool
	IsLoggingStats  bool

	loading           sync.Mutex // guards initialization of gitConfig and remotes
	origConfig        map[string]string
	remotes           []string
	extensions        map[string]Extension
	fetchIncludePaths []string
	fetchExcludePaths []string
	fetchPruneConfig  *FetchPruneConfig
	manualEndpoint    *Endpoint
	parsedNetrc       netrcfinder
}

func New() *Configuration {
	c := &Configuration{
		Os: EnvironmentOf(NewOsFetcher()),

		CurrentRemote: defaultRemote,
		envVars:       make(map[string]string),
	}
	c.IsTracingHttp = c.GetenvBool("GIT_CURL_VERBOSE", false)
	c.IsDebuggingHttp = c.GetenvBool("LFS_DEBUG_HTTP", false)
	c.IsLoggingStats = c.GetenvBool("GIT_LOG_STATS", false)
	return c
}

// Values is a convenience type used to call the NewFromValues function. It
// specifies `Git` and `Env` maps to use as mock values, instead of calling out
// to real `.gitconfig`s and the `os.Getenv` function.
type Values struct {
	// Git and Os are the stand-in maps used to provide values for their
	// respective environments.
	Git, Os map[string]string
}

// NewFrom returns a new `*config.Configuration` that reads both its Git
// and Enviornment-level values from the ones provided instead of the actual
// `.gitconfig` file or `os.Getenv`, respectively.
//
// This method should only be used during testing.
func NewFrom(v Values) *Configuration {
	return &Configuration{
		Os:  EnvironmentOf(mapFetcher(v.Os)),
		Git: EnvironmentOf(mapFetcher(v.Git)),

		envVars: make(map[string]string, 0),
	}
}

// Getenv is shorthand for `c.Os.Get(key)`.
func (c *Configuration) Getenv(key string) string {
	return c.Os.Get(key)
}

// GetenvBool is shorthand for `c.Os.Bool(key, def)`.
func (c *Configuration) GetenvBool(key string, def bool) bool {
	return c.Os.Bool(key, def)
}

// GitRemoteUrl returns the git clone/push url for a given remote (blank if not found)
// the forpush argument is to cater for separate remote.name.pushurl settings
func (c *Configuration) GitRemoteUrl(remote string, forpush bool) string {
	if forpush {
		if u, ok := c.GitConfig("remote." + remote + ".pushurl"); ok {
			return u
		}
	}

	if u, ok := c.GitConfig("remote." + remote + ".url"); ok {
		return u
	}

	return ""

}

// Manually set an Endpoint to use instead of deriving from Git config
func (c *Configuration) SetManualEndpoint(e Endpoint) {
	c.manualEndpoint = &e
}

func (c *Configuration) Endpoint(operation string) Endpoint {
	if c.manualEndpoint != nil {
		return *c.manualEndpoint
	}

	if operation == "upload" {
		if url, ok := c.GitConfig("lfs.pushurl"); ok {
			return NewEndpointWithConfig(url, c)
		}
	}

	if url, ok := c.GitConfig("lfs.url"); ok {
		return NewEndpointWithConfig(url, c)
	}

	if len(c.CurrentRemote) > 0 && c.CurrentRemote != defaultRemote {
		if endpoint := c.RemoteEndpoint(c.CurrentRemote, operation); len(endpoint.Url) > 0 {
			return endpoint
		}
	}

	return c.RemoteEndpoint(defaultRemote, operation)
}

func (c *Configuration) ConcurrentTransfers() int {
	if c.NtlmAccess("download") {
		return 1
	}

	uploads := 3

	if v, ok := c.GitConfig("lfs.concurrenttransfers"); ok {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			uploads = n
		}
	}

	return uploads
}

// BasicTransfersOnly returns whether to only allow "basic" HTTP transfers.
// Default is false, including if the lfs.basictransfersonly is invalid
func (c *Configuration) BasicTransfersOnly() bool {
	return c.GitConfigBool("lfs.basictransfersonly", false)
}

// TusTransfersAllowed returns whether to only use "tus.io" HTTP transfers.
// Default is false, including if the lfs.tustransfers is invalid
func (c *Configuration) TusTransfersAllowed() bool {
	return c.GitConfigBool("lfs.tustransfers", false)
}

func (c *Configuration) BatchTransfer() bool {
	return c.GitConfigBool("lfs.batch", true)
}

func (c *Configuration) NtlmAccess(operation string) bool {
	return c.Access(operation) == "ntlm"
}

// PrivateAccess will retrieve the access value and return true if
// the value is set to private. When a repo is marked as having private
// access, the http requests for the batch api will fetch the credentials
// before running, otherwise the request will run without credentials.
func (c *Configuration) PrivateAccess(operation string) bool {
	return c.Access(operation) != "none"
}

// Access returns the access auth type.
func (c *Configuration) Access(operation string) string {
	return c.EndpointAccess(c.Endpoint(operation))
}

// SetAccess will set the private access flag in .git/config.
func (c *Configuration) SetAccess(operation string, authType string) {
	c.SetEndpointAccess(c.Endpoint(operation), authType)
}

func (c *Configuration) FindNetrcHost(host string) (*netrc.Machine, error) {
	c.loading.Lock()
	defer c.loading.Unlock()
	if c.parsedNetrc == nil {
		n, err := c.parseNetrc()
		if err != nil {
			return nil, err
		}
		c.parsedNetrc = n
	}

	return c.parsedNetrc.FindMachine(host), nil
}

// Manually override the netrc config
func (c *Configuration) SetNetrc(n netrcfinder) {
	c.parsedNetrc = n
}

func (c *Configuration) EndpointAccess(e Endpoint) string {
	key := fmt.Sprintf("lfs.%s.access", e.Url)
	if v, ok := c.GitConfig(key); ok && len(v) > 0 {
		lower := strings.ToLower(v)
		if lower == "private" {
			return "basic"
		}
		return lower
	}
	return "none"
}

func (c *Configuration) SetEndpointAccess(e Endpoint, authType string) {
	tracerx.Printf("setting repository access to %s", authType)
	key := fmt.Sprintf("lfs.%s.access", e.Url)

	// Modify the config cache because it's checked again in this process
	// without being reloaded.
	switch authType {
	case "", "none":
		git.Config.UnsetLocalKey("", key)

		c.loading.Lock()
		delete(c.gitConfig, strings.ToLower(key))
		c.loading.Unlock()
	default:
		git.Config.SetLocal("", key, authType)

		c.loading.Lock()
		c.gitConfig[strings.ToLower(key)] = authType
		c.loading.Unlock()
	}
}

func (c *Configuration) FetchIncludePaths() []string {
	c.loadGitConfig()
	return c.fetchIncludePaths
}

func (c *Configuration) FetchExcludePaths() []string {
	c.loadGitConfig()
	return c.fetchExcludePaths
}

func (c *Configuration) RemoteEndpoint(remote, operation string) Endpoint {
	if len(remote) == 0 {
		remote = defaultRemote
	}

	// Support separate push URL if specified and pushing
	if operation == "upload" {
		if url, ok := c.GitConfig("remote." + remote + ".lfspushurl"); ok {
			return NewEndpointWithConfig(url, c)
		}
	}
	if url, ok := c.GitConfig("remote." + remote + ".lfsurl"); ok {
		return NewEndpointWithConfig(url, c)
	}

	// finally fall back on git remote url (also supports pushurl)
	if url := c.GitRemoteUrl(remote, operation == "upload"); url != "" {
		return NewEndpointFromCloneURLWithConfig(url, c)
	}

	return Endpoint{}
}

func (c *Configuration) Remotes() []string {
	c.loadGitConfig()
	return c.remotes
}

// GitProtocol returns the protocol for the LFS API when converting from a
// git:// remote url.
func (c *Configuration) GitProtocol() string {
	if value, ok := c.GitConfig("lfs.gitprotocol"); ok {
		return value
	}
	return "https"
}

func (c *Configuration) Extensions() map[string]Extension {
	c.loadGitConfig()
	return c.extensions
}

// SortedExtensions gets the list of extensions ordered by Priority
func (c *Configuration) SortedExtensions() ([]Extension, error) {
	return SortExtensions(c.Extensions())
}

// GitConfigInt parses a git config value and returns it as an integer.
func (c *Configuration) GitConfigInt(key string, def int) int {
	s, _ := c.GitConfig(key)
	if len(s) == 0 {
		return def
	}

	i, _ := strconv.Atoi(s)
	if i < 1 {
		return def
	}

	return i
}

// GitConfigBool parses a git config value and returns true if defined as
// true, 1, on, yes, or def if not defined
func (c *Configuration) GitConfigBool(key string, def bool) bool {
	s, _ := c.GitConfig(key)
	if len(s) == 0 {
		return def
	}

	ret, err := parseConfigBool(s)
	if err != nil {
		return false
	}
	return ret
}

func (c *Configuration) GitConfig(key string) (string, bool) {
	c.loadGitConfig()
	value, ok := c.gitConfig[strings.ToLower(key)]
	return value, ok
}

func (c *Configuration) AllGitConfig() map[string]string {
	c.loadGitConfig()
	return c.gitConfig
}

func (c *Configuration) FetchPruneConfig() *FetchPruneConfig {
	if c.fetchPruneConfig == nil {
		c.fetchPruneConfig = &FetchPruneConfig{
			FetchRecentRefsDays:           7,
			FetchRecentRefsIncludeRemotes: true,
			FetchRecentCommitsDays:        0,
			PruneOffsetDays:               3,
			PruneVerifyRemoteAlways:       false,
			PruneRemoteName:               "origin",
		}
		if v, ok := c.GitConfig("lfs.fetchrecentrefsdays"); ok {
			n, err := strconv.Atoi(v)
			if err == nil && n >= 0 {
				c.fetchPruneConfig.FetchRecentRefsDays = n
			}
		}
		if v, ok := c.GitConfig("lfs.fetchrecentremoterefs"); ok {
			if b, err := parseConfigBool(v); err == nil {
				c.fetchPruneConfig.FetchRecentRefsIncludeRemotes = b
			}
		}
		if v, ok := c.GitConfig("lfs.fetchrecentcommitsdays"); ok {
			n, err := strconv.Atoi(v)
			if err == nil && n >= 0 {
				c.fetchPruneConfig.FetchRecentCommitsDays = n
			}
		}
		if v, ok := c.GitConfig("lfs.fetchrecentalways"); ok {
			if b, err := parseConfigBool(v); err == nil {
				c.fetchPruneConfig.FetchRecentAlways = b
			}
		}
		if v, ok := c.GitConfig("lfs.pruneoffsetdays"); ok {
			n, err := strconv.Atoi(v)
			if err == nil && n >= 0 {
				c.fetchPruneConfig.PruneOffsetDays = n
			}
		}
		if v, ok := c.GitConfig("lfs.pruneverifyremotealways"); ok {
			if b, err := parseConfigBool(v); err == nil {
				c.fetchPruneConfig.PruneVerifyRemoteAlways = b
			}
		}
		if v, ok := c.GitConfig("lfs.pruneremotetocheck"); ok {
			c.fetchPruneConfig.PruneRemoteName = v
		}

	}
	return c.fetchPruneConfig
}

func (c *Configuration) SkipDownloadErrors() bool {
	return c.GetenvBool("GIT_LFS_SKIP_DOWNLOAD_ERRORS", false) || c.GitConfigBool("lfs.skipdownloaderrors", false)
}

func parseConfigBool(str string) (bool, error) {
	switch strings.ToLower(str) {
	case "true", "1", "on", "yes", "t":
		return true, nil
	case "false", "0", "off", "no", "f":
		return false, nil
	}
	return false, fmt.Errorf("Unable to parse %q as a boolean", str)
}

func (c *Configuration) loadGitConfig() bool {
	c.loading.Lock()
	defer c.loading.Unlock()

	if c.gitConfig != nil {
		return false
	}

	var sources []*GitConfig

	lfsconfig := filepath.Join(LocalWorkingDir, ".lfsconfig")
	if _, err := os.Stat(lfsconfig); !os.IsNotExist(err) {
		lines, err := git.Config.ListFromFile(lfsconfig)
		if err != nil {
			sources = append(sources, &GitConfig{
				Lines:        strings.Split(lines, "\n"),
				OnlySafeKeys: true,
			})
		} else {
			gitconfig := filepath.Join(LocalWorkingDir, ".gitconfig")
			if _, err := os.Stat(lfsconfig); !os.IsNotExist(err) {
				if ShowConfigWarnings {
					expected := ".lfsconfig"
					fmt.Fprintf(os.Stderr, "WARNING: Reading LFS config from %q, not %q. Rename to %q before Git LFS v2.0 to remove this warning.\n",
						filepath.Base(gitconfig), expected, expected)
				}

				lines, err := git.Config.ListFromFile(lfsconfig)
				if err != nil {
					sources = append(sources, &GitConfig{
						Lines:        strings.Split(lines, "\n"),
						OnlySafeKeys: true,
					})
				}
			}
		}
	}

	globalList, err := git.Config.List()
	if err != nil {
		sources = append(sources, &GitConfig{
			Lines:        strings.Split(globalList, "\n"),
			OnlySafeKeys: false,
		})
	}

	gf, extensions, uniqRemotes, include, exclude := ReadGitConfig(sources...)

	c.fetchIncludePaths = include
	c.fetchExcludePaths = exclude

	c.Git = EnvironmentOf(gf)
	c.gitConfig = gf.vals // XXX TERRIBLE
	c.extensions = extensions

	c.remotes = make([]string, 0, len(uniqRemotes))
	for remote, isOrigin := range uniqRemotes {
		if isOrigin {
			continue
		}
		c.remotes = append(c.remotes, remote)
	}

	return true
}

// XXX(taylor): remove mutability
func (c *Configuration) SetConfig(key, value string) {
	if c.loadGitConfig() {
		c.loading.Lock()
		c.origConfig = make(map[string]string)
		for k, v := range c.gitConfig {
			c.origConfig[k] = v
		}
		c.loading.Unlock()
	}

	c.gitConfig[key] = value
}

// XXX(taylor): remove mutability
func (c *Configuration) ClearConfig() {
	if c.loadGitConfig() {
		c.loading.Lock()
		c.origConfig = make(map[string]string)
		for k, v := range c.gitConfig {
			c.origConfig[k] = v
		}
		c.loading.Unlock()
	}

	c.gitConfig = make(map[string]string)
}

// XXX(taylor): remove mutability
func (c *Configuration) ResetConfig() {
	c.loading.Lock()
	c.gitConfig = make(map[string]string)
	for k, v := range c.origConfig {
		c.gitConfig[k] = v
	}
	c.loading.Unlock()
}
