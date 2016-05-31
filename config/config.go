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
	Config                 = NewConfig()
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
	CurrentRemote   string
	NtlmSession     ntlm.ClientSession
	envVars         map[string]string
	envVarsMutex    sync.Mutex
	IsTracingHttp   bool
	IsDebuggingHttp bool
	IsLoggingStats  bool

	loading           sync.Mutex // guards initialization of gitConfig and remotes
	gitConfig         map[string]string
	origConfig        map[string]string
	remotes           []string
	extensions        map[string]Extension
	fetchIncludePaths []string
	fetchExcludePaths []string
	fetchPruneConfig  *FetchPruneConfig
	manualEndpoint    *Endpoint
	parsedNetrc       netrcfinder
}

func NewConfig() *Configuration {
	c := &Configuration{
		CurrentRemote: defaultRemote,
		envVars:       make(map[string]string),
	}
	c.IsTracingHttp = c.GetenvBool("GIT_CURL_VERBOSE", false)
	c.IsDebuggingHttp = c.GetenvBool("LFS_DEBUG_HTTP", false)
	c.IsLoggingStats = c.GetenvBool("GIT_LOG_STATS", false)
	return c
}

func (c *Configuration) Getenv(key string) string {
	c.envVarsMutex.Lock()
	defer c.envVarsMutex.Unlock()

	if i, ok := c.envVars[key]; ok {
		return i
	}

	v := os.Getenv(key)
	c.envVars[key] = v
	return v
}

func (c *Configuration) Setenv(key, value string) error {
	c.envVarsMutex.Lock()
	defer c.envVarsMutex.Unlock()

	// Check see if we have this in our cache, if so update it
	if _, ok := c.envVars[key]; ok {
		c.envVars[key] = value
	}

	// Now set in process
	return os.Setenv(key, value)
}

func (c *Configuration) GetAllEnv() map[string]string {
	c.envVarsMutex.Lock()
	defer c.envVarsMutex.Unlock()

	ret := make(map[string]string)
	for k, v := range c.envVars {
		ret[k] = v
	}
	return ret
}

func (c *Configuration) SetAllEnv(env map[string]string) {
	c.envVarsMutex.Lock()
	defer c.envVarsMutex.Unlock()

	c.envVars = make(map[string]string)
	for k, v := range env {
		c.envVars[k] = v
	}
}

// GetenvBool parses a boolean environment variable and returns the result as a bool.
// If the environment variable is unset, empty, or if the parsing fails,
// the value of def (default) is returned instead.
func (c *Configuration) GetenvBool(key string, def bool) bool {
	s := c.Getenv(key)
	if len(s) == 0 {
		return def
	}

	b, err := parseConfigBool(s)
	if err != nil {
		return def
	}
	return b
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

func (c *Configuration) BatchTransfer() bool {
	value, ok := c.GitConfig("lfs.batch")
	if !ok || len(value) == 0 {
		return true
	}

	useBatch, err := parseConfigBool(value)
	if err != nil {
		return false
	}

	return useBatch
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
// true, 1, on, yes
func (c *Configuration) GitConfigBool(key string) bool {
	s, _ := c.GitConfig(key)
	if len(s) == 0 {
		return false
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
	return c.GetenvBool("GIT_LFS_SKIP_DOWNLOAD_ERRORS", false) || c.GitConfigBool("lfs.skipdownloaderrors")
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

	c.gitConfig = make(map[string]string)
	c.extensions = make(map[string]Extension)
	uniqRemotes := make(map[string]bool)

	configFiles := []string{
		filepath.Join(LocalWorkingDir, ".lfsconfig"),

		// TODO: remove .gitconfig support for Git LFS v2.0 https://github.com/github/git-lfs/issues/839
		filepath.Join(LocalWorkingDir, ".gitconfig"),
	}
	c.readGitConfigFromFiles(configFiles, 0, uniqRemotes)

	listOutput, err := git.Config.List()
	if err != nil {
		panic(fmt.Errorf("Error listing git config: %s", err))
	}

	c.readGitConfig(listOutput, uniqRemotes, false)

	c.remotes = make([]string, 0, len(uniqRemotes))
	for remote, isOrigin := range uniqRemotes {
		if isOrigin {
			continue
		}
		c.remotes = append(c.remotes, remote)
	}

	return true
}

func (c *Configuration) readGitConfigFromFiles(filenames []string, filenameIndex int, uniqRemotes map[string]bool) {
	filename := filenames[filenameIndex]
	_, err := os.Stat(filename)
	if err == nil {
		if filenameIndex > 0 && ShowConfigWarnings {
			expected := ".lfsconfig"
			fmt.Fprintf(os.Stderr, "WARNING: Reading LFS config from %q, not %q. Rename to %q before Git LFS v2.0 to remove this warning.\n",
				filepath.Base(filename), expected, expected)
		}

		fileOutput, err := git.Config.ListFromFile(filename)
		if err != nil {
			panic(fmt.Errorf("Error listing git config from %s: %s", filename, err))
		}
		c.readGitConfig(fileOutput, uniqRemotes, true)
		return
	}

	if os.IsNotExist(err) {
		newIndex := filenameIndex + 1
		if len(filenames) > newIndex {
			c.readGitConfigFromFiles(filenames, newIndex, uniqRemotes)
		}
		return
	}

	panic(fmt.Errorf("Error listing git config from %s: %s", filename, err))
}

func (c *Configuration) readGitConfig(output string, uniqRemotes map[string]bool, onlySafe bool) {
	lines := strings.Split(output, "\n")
	uniqKeys := make(map[string]string)

	for _, line := range lines {
		pieces := strings.SplitN(line, "=", 2)
		if len(pieces) < 2 {
			continue
		}

		allowed := !onlySafe
		key := strings.ToLower(pieces[0])
		value := pieces[1]

		if origKey, ok := uniqKeys[key]; ok {
			if ShowConfigWarnings && c.gitConfig[key] != value && strings.HasPrefix(key, gitConfigWarningPrefix) {
				fmt.Fprintf(os.Stderr, "WARNING: These git config values clash:\n")
				fmt.Fprintf(os.Stderr, "  git config %q = %q\n", origKey, c.gitConfig[key])
				fmt.Fprintf(os.Stderr, "  git config %q = %q\n", pieces[0], value)
			}
		} else {
			uniqKeys[key] = pieces[0]
		}

		keyParts := strings.Split(key, ".")
		if len(keyParts) == 4 && keyParts[0] == "lfs" && keyParts[1] == "extension" {
			name := keyParts[2]
			ext := c.extensions[name]
			switch keyParts[3] {
			case "clean":
				if onlySafe {
					continue
				}
				ext.Clean = value
			case "smudge":
				if onlySafe {
					continue
				}
				ext.Smudge = value
			case "priority":
				allowed = true
				p, err := strconv.Atoi(value)
				if err == nil && p >= 0 {
					ext.Priority = p
				}
			}

			ext.Name = name
			c.extensions[name] = ext
		} else if len(keyParts) > 1 && keyParts[0] == "remote" {
			if onlySafe && (len(keyParts) == 3 && keyParts[2] != "lfsurl") {
				continue
			}

			allowed = true
			remote := keyParts[1]
			uniqRemotes[remote] = remote == "origin"
		} else if len(keyParts) > 2 && keyParts[len(keyParts)-1] == "access" {
			allowed = true
		}

		if !allowed && keyIsUnsafe(key) {
			continue
		}

		c.gitConfig[key] = value

		if len(keyParts) == 2 && keyParts[0] == "lfs" && keyParts[1] == "fetchinclude" {
			for _, inc := range strings.Split(value, ",") {
				inc = strings.TrimSpace(inc)
				c.fetchIncludePaths = append(c.fetchIncludePaths, inc)
			}
		} else if len(keyParts) == 2 && keyParts[0] == "lfs" && keyParts[1] == "fetchexclude" {
			for _, ex := range strings.Split(value, ",") {
				ex = strings.TrimSpace(ex)
				c.fetchExcludePaths = append(c.fetchExcludePaths, ex)
			}
		}
	}
}

func keyIsUnsafe(key string) bool {
	for _, safe := range safeKeys {
		if safe == key {
			return false
		}
	}
	return true
}

var safeKeys = []string{
	"lfs.fetchexclude",
	"lfs.fetchinclude",
	"lfs.gitprotocol",
	"lfs.url",
}

// only used for tests
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

func (c *Configuration) ResetConfig() {
	c.loading.Lock()
	c.gitConfig = make(map[string]string)
	for k, v := range c.origConfig {
		c.gitConfig[k] = v
	}
	c.loading.Unlock()
}
