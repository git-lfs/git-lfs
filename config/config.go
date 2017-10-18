// Package config collects together all configuration settings
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package config

import (
	"sync"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tools"
)

var (
	Config                 = New()
	ShowConfigWarnings     = false
	defaultRemote          = "origin"
	gitConfigWarningPrefix = "lfs."
)

type Configuration struct {
	// Os provides a `*Environment` used to access to the system's
	// environment through os.Getenv. It is the point of entry for all
	// system environment configuration.
	Os Environment

	// Git provides a `*Environment` used to access to the various levels of
	// `.gitconfig`'s. It is the point of entry for all Git environment
	// configuration.
	Git Environment

	// gitConfig can fetch or modify the current Git config and track the Git
	// version.
	gitConfig *git.Configuration

	CurrentRemote string

	loading    sync.Mutex // guards initialization of gitConfig and remotes
	remotes    []string
	extensions map[string]Extension
}

func New() *Configuration {
	c := &Configuration{
		CurrentRemote: defaultRemote,
		Os:            EnvironmentOf(NewOsFetcher()),
		gitConfig:     git.Config,
	}
	c.Git = &gitEnvironment{
		config:    c,
		gitConfig: git.Config,
	}
	return c
}

// Values is a convenience type used to call the NewFromValues function. It
// specifies `Git` and `Env` maps to use as mock values, instead of calling out
// to real `.gitconfig`s and the `os.Getenv` function.
type Values struct {
	// Git and Os are the stand-in maps used to provide values for their
	// respective environments.
	Git, Os map[string][]string
}

// NewFrom returns a new `*config.Configuration` that reads both its Git
// and Enviornment-level values from the ones provided instead of the actual
// `.gitconfig` file or `os.Getenv`, respectively.
//
// This method should only be used during testing.
func NewFrom(v Values) *Configuration {
	return &Configuration{
		CurrentRemote: defaultRemote,
		Os:            EnvironmentOf(mapFetcher(v.Os)),
		Git:           EnvironmentOf(mapFetcher(v.Git)),
		gitConfig:     git.Config,
	}
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

func (c *Configuration) FetchIncludePaths() []string {
	patterns, _ := c.Git.Get("lfs.fetchinclude")
	return tools.CleanPaths(patterns, ",")
}

func (c *Configuration) FetchExcludePaths() []string {
	patterns, _ := c.Git.Get("lfs.fetchexclude")
	return tools.CleanPaths(patterns, ",")
}

func (c *Configuration) Remotes() []string {
	c.loadGitConfig()

	return c.remotes
}

func (c *Configuration) Extensions() map[string]Extension {
	c.loadGitConfig()
	return c.extensions
}

// SortedExtensions gets the list of extensions ordered by Priority
func (c *Configuration) SortedExtensions() ([]Extension, error) {
	return SortExtensions(c.Extensions())
}

func (c *Configuration) SkipDownloadErrors() bool {
	return c.Os.Bool("GIT_LFS_SKIP_DOWNLOAD_ERRORS", false) || c.Git.Bool("lfs.skipdownloaderrors", false)
}

func (c *Configuration) SetLockableFilesReadOnly() bool {
	return c.Os.Bool("GIT_LFS_SET_LOCKABLE_READONLY", true) && c.Git.Bool("lfs.setlockablereadonly", true)
}

func (c *Configuration) GitConfig() *git.Configuration {
	return c.gitConfig
}

func (c *Configuration) GitVersion() (string, error) {
	return c.gitConfig.Version()
}

func (c *Configuration) IsGitVersionAtLeast(ver string) bool {
	return c.gitConfig.IsGitVersionAtLeast(ver)
}

func (c *Configuration) FindGitGlobalKey(key string) string {
	return c.gitConfig.FindGlobal(key)
}

func (c *Configuration) FindGitSystemKey(key string) string {
	return c.gitConfig.FindSystem(key)
}

func (c *Configuration) FindGitLocalKey(key string) string {
	return c.gitConfig.FindLocal(key)
}

func (c *Configuration) SetGitGlobalKey(key, val string) (string, error) {
	return c.gitConfig.SetGlobal(key, val)
}

func (c *Configuration) SetGitSystemKey(key, val string) (string, error) {
	return c.gitConfig.SetSystem(key, val)
}

func (c *Configuration) SetGitLocalKey(file, key, val string) (string, error) {
	return c.gitConfig.SetLocal(file, key, val)
}

func (c *Configuration) UnsetGitGlobalSection(key string) (string, error) {
	return c.gitConfig.UnsetGlobalSection(key)
}

func (c *Configuration) UnsetGitSystemSection(key string) (string, error) {
	return c.gitConfig.UnsetSystemSection(key)
}

func (c *Configuration) UnsetGitLocalSection(key string) (string, error) {
	return c.gitConfig.UnsetLocalSection(key)
}

func (c *Configuration) UnsetGitLocalKey(file, key string) (string, error) {
	return c.gitConfig.UnsetLocalKey(file, key)
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
