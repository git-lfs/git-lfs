// Package config collects together all configuration settings
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/tools"
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
	FetchRecentRefsDays int `git:"lfs.fetchrecentrefsdays"`
	// Makes the FetchRecentRefsDays option apply to remote refs from fetch source as well (default true)
	FetchRecentRefsIncludeRemotes bool `git:"lfs.fetchrecentremoterefs"`
	// number of days prior to latest commit on a ref that we'll fetch previous
	// LFS changes too (default 0 = only fetch at ref)
	FetchRecentCommitsDays int `git:"lfs.fetchrecentcommitsdays"`
	// Whether to always fetch recent even without --recent
	FetchRecentAlways bool `git:"lfs.fetchrecentalways"`
	// Number of days added to FetchRecent*; data outside combined window will be
	// deleted when prune is run. (default 3)
	PruneOffsetDays int `git:"lfs.pruneoffsetdays"`
	// Always verify with remote before pruning
	PruneVerifyRemoteAlways bool `git:"lfs.pruneverifyremotealways"`
	// Name of remote to check for unpushed and verify checks
	PruneRemoteName string `git:"lfs.pruneremotetocheck"`
}

type Configuration struct {
	// Os provides a `*Environment` used to access to the system's
	// environment through os.Getenv. It is the point of entry for all
	// system environment configuration.
	Os Environment

	// Git provides a `*Environment` used to access to the various levels of
	// `.gitconfig`'s. It is the point of entry for all Git environment
	// configuration.
	Git Environment

	CurrentRemote string

	loading        sync.Mutex // guards initialization of gitConfig and remotes
	remotes        []string
	extensions     map[string]Extension
	manualEndpoint *lfsapi.Endpoint
	endpointFinder lfsapi.EndpointFinder
	endpointMu     sync.Mutex
}

func New() *Configuration {
	c := &Configuration{Os: EnvironmentOf(NewOsFetcher())}
	c.Git = &gitEnvironment{config: c}
	initConfig(c)
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
	c := &Configuration{
		Os:  EnvironmentOf(mapFetcher(v.Os)),
		Git: EnvironmentOf(mapFetcher(v.Git)),
	}
	initConfig(c)
	return c
}

func initConfig(c *Configuration) {
	c.CurrentRemote = defaultRemote
}

// Unmarshal unmarshals the *Configuration in context into all of `v`'s fields,
// according to the following rules:
//
// Values are marshaled according to the given key and environment, as follows:
//	type T struct {
//		Field string `git:"key"`
//		Other string `os:"key"`
//	}
//
// If an unknown environment is given, an error will be returned. If there is no
// method supporting conversion into a field's type, an error will be returned.
// If no value is associated with the given key and environment, the field will
// // only be modified if there is a config value present matching the given
// key. If the field is already set to a non-zero value of that field's type,
// then it will be left alone.
//
// Otherwise, the field will be set to the value of calling the
// appropriately-typed method on the specified environment.
func (c *Configuration) Unmarshal(v interface{}) error {
	into := reflect.ValueOf(v)
	if into.Kind() != reflect.Ptr {
		return fmt.Errorf("lfs/config: unable to parse non-pointer type of %T", v)
	}
	into = into.Elem()

	for i := 0; i < into.Type().NumField(); i++ {
		field := into.Field(i)
		sfield := into.Type().Field(i)

		key, env, err := c.parseTag(sfield.Tag)
		if err != nil {
			return err
		}

		if env == nil {
			continue
		}

		var val interface{}
		switch sfield.Type.Kind() {
		case reflect.String:
			var ok bool

			val, ok = env.Get(key)
			if !ok {
				val = field.String()
			}
		case reflect.Int:
			val = env.Int(key, int(field.Int()))
		case reflect.Bool:
			val = env.Bool(key, field.Bool())
		default:
			return fmt.Errorf(
				"lfs/config: unsupported target type for field %q: %v",
				sfield.Name, sfield.Type.String())
		}

		if val != nil {
			into.Field(i).Set(reflect.ValueOf(val))
		}
	}

	return nil
}

// parseTag returns the key, environment, and optional error assosciated with a
// given tag. It will return the XOR of either the `git` or `os` tag. That is to
// say, a field tagged with EITHER `git` OR `os` is valid, but pone tagged with
// both is not.
//
// If neither field was found, then a nil environment will be returned.
func (c *Configuration) parseTag(tag reflect.StructTag) (key string, env Environment, err error) {
	git, os := tag.Get("git"), tag.Get("os")

	if len(git) != 0 && len(os) != 0 {
		return "", nil, errors.New("lfs/config: ambiguous tags")
	}

	if len(git) != 0 {
		return git, c.Git, nil
	}
	if len(os) != 0 {
		return os, c.Os, nil
	}

	return
}

// GitRemoteUrl returns the git clone/push url for a given remote (blank if not found)
// the forpush argument is to cater for separate remote.name.pushurl settings
func (c *Configuration) GitRemoteUrl(remote string, forpush bool) string {
	return c.endpointConfig().GitRemoteURL(remote, forpush)
}

// Manually set an Endpoint to use instead of deriving from Git config
func (c *Configuration) SetManualEndpoint(e lfsapi.Endpoint) {
	c.manualEndpoint = &e
}

func (c *Configuration) Endpoint(operation string) lfsapi.Endpoint {
	if c.manualEndpoint != nil {
		return *c.manualEndpoint
	}
	return c.endpointConfig().Endpoint(operation, c.CurrentRemote)
}

func (c *Configuration) ConcurrentTransfers() int {
	if c.NtlmAccess("download") {
		return 1
	}

	uploads := 3

	if v, ok := c.Git.Get("lfs.concurrenttransfers"); ok {
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
	return c.Git.Bool("lfs.basictransfersonly", false)
}

// TusTransfersAllowed returns whether to only use "tus.io" HTTP transfers.
// Default is false, including if the lfs.tustransfers is invalid
func (c *Configuration) TusTransfersAllowed() bool {
	return c.Git.Bool("lfs.tustransfers", false)
}

func (c *Configuration) BatchTransfer() bool {
	return c.Git.Bool("lfs.batch", true)
}

func (c *Configuration) NtlmAccess(operation string) bool {
	return c.Access(operation) == lfsapi.NTLMAccess
}

// Access returns the access auth type.
func (c *Configuration) Access(operation string) lfsapi.Access {
	return c.EndpointAccess(c.Endpoint(operation))
}

func (c *Configuration) EndpointAccess(e lfsapi.Endpoint) lfsapi.Access {
	return c.endpointConfig().AccessFor(e.Url)
}

func (c *Configuration) FetchIncludePaths() []string {
	patterns, _ := c.Git.Get("lfs.fetchinclude")
	return tools.CleanPaths(patterns, ",")
}

func (c *Configuration) FetchExcludePaths() []string {
	patterns, _ := c.Git.Get("lfs.fetchexclude")
	return tools.CleanPaths(patterns, ",")
}

func (c *Configuration) RemoteEndpoint(remote, operation string) lfsapi.Endpoint {
	return c.endpointConfig().RemoteEndpoint(operation, remote)
}

func (c *Configuration) Remotes() []string {
	c.loadGitConfig()

	return c.remotes
}

func (c *Configuration) GitProtocol() string {
	return c.endpointConfig().GitProtocol()
}

func (c *Configuration) endpointConfig() lfsapi.EndpointFinder {
	c.endpointMu.Lock()
	defer c.endpointMu.Unlock()

	if c.endpointFinder == nil {
		c.endpointFinder = lfsapi.NewEndpointFinder(c.Git)
	}

	return c.endpointFinder
}

func (c *Configuration) Extensions() map[string]Extension {
	c.loadGitConfig()
	return c.extensions
}

// SortedExtensions gets the list of extensions ordered by Priority
func (c *Configuration) SortedExtensions() ([]Extension, error) {
	return SortExtensions(c.Extensions())
}

func (c *Configuration) FetchPruneConfig() FetchPruneConfig {
	f := &FetchPruneConfig{
		FetchRecentRefsDays:           7,
		FetchRecentRefsIncludeRemotes: true,
		PruneOffsetDays:               3,
		PruneRemoteName:               "origin",
	}

	if err := c.Unmarshal(f); err != nil {
		panic(err.Error())
	}
	return *f
}

func (c *Configuration) SkipDownloadErrors() bool {
	return c.Os.Bool("GIT_LFS_SKIP_DOWNLOAD_ERRORS", false) || c.Git.Bool("lfs.skipdownloaderrors", false)
}

func (c *Configuration) SetLockableFilesReadOnly() bool {
	return c.Os.Bool("GIT_LFS_SET_LOCKABLE_READONLY", true) && c.Git.Bool("lfs.setlockablereadonly", true)
}

// loadGitConfig is a temporary measure to support legacy behavior dependent on
// accessing properties set by ReadGitConfig, namely:
//  - `c.extensions`
//  - `c.uniqRemotes`
//  - `c.gitConfig`
//
// Since the *gitEnvironment is responsible for setting these values on the
// (*config.Configuration) instance, we must call that method, if it exists.
//
// loadGitConfig returns a bool returning whether or not `loadGitConfig` was
// called AND the method did not return early.
func (c *Configuration) loadGitConfig() bool {
	if g, ok := c.Git.(*gitEnvironment); ok {
		return g.loadGitConfig()
	}

	return false
}

// CurrentCommitter returns the name/email that would be used to author a commit
// with this configuration. In particular, the "user.name" and "user.email"
// configuration values are used
func (c *Configuration) CurrentCommitter() (name, email string) {
	name, _ = c.Git.Get("user.name")
	email, _ = c.Git.Get("user.email")
	return
}
