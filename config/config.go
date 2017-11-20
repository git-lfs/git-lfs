// Package config collects together all configuration settings
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/fs"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

var (
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

	currentRemote *string
	pushRemote    *string

	// gitConfig can fetch or modify the current Git config and track the Git
	// version.
	gitConfig *git.Configuration

	ref        *git.Ref
	remoteRef  *git.Ref
	fs         *fs.Filesystem
	gitDir     *string
	workDir    string
	loading    sync.Mutex // guards initialization of gitConfig and remotes
	loadingGit sync.Mutex // guards initialization of local git and working dirs
	remotes    []string
	extensions map[string]Extension
}

func New() *Configuration {
	return NewIn("", "")
}

func NewIn(workdir, gitdir string) *Configuration {
	gitConf := git.NewConfig(workdir, gitdir)
	c := &Configuration{
		Os:        EnvironmentOf(NewOsFetcher()),
		gitConfig: gitConf,
	}

	if len(gitConf.WorkDir) > 0 {
		c.gitDir = &gitConf.GitDir
		c.workDir = gitConf.WorkDir
	}

	c.Git = &delayedEnvironment{
		callback: func() Environment {
			sources, err := gitConf.Sources(filepath.Join(c.LocalWorkingDir(), ".lfsconfig"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading git config: %s\n", err)
			}
			return c.readGitConfig(sources...)
		},
	}
	return c
}

func (c *Configuration) readGitConfig(gitconfigs ...*git.ConfigurationSource) Environment {
	gf, extensions, uniqRemotes := readGitConfig(gitconfigs...)
	c.extensions = extensions
	c.remotes = make([]string, 0, len(uniqRemotes))
	for remote, isOrigin := range uniqRemotes {
		if isOrigin {
			continue
		}
		c.remotes = append(c.remotes, remote)
	}

	return EnvironmentOf(gf)
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
	c := &Configuration{
		Os:        EnvironmentOf(mapFetcher(v.Os)),
		gitConfig: git.NewConfig("", ""),
	}
	c.Git = &delayedEnvironment{
		callback: func() Environment {
			source := &git.ConfigurationSource{
				Lines: make([]string, 0, len(v.Git)),
			}

			for key, values := range v.Git {
				for _, value := range values {
					fmt.Printf("Config: %s=%s\n", key, value)
					source.Lines = append(source.Lines, fmt.Sprintf("%s=%s", key, value))
				}
			}

			return c.readGitConfig(source)
		},
	}
	return c
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

func (c *Configuration) CurrentRef() *git.Ref {
	c.loading.Lock()
	defer c.loading.Unlock()
	if c.ref == nil {
		r, err := git.CurrentRef()
		if err != nil {
			tracerx.Printf("Error loading current ref: %s", err)
			c.ref = &git.Ref{}
		} else {
			c.ref = r
		}
	}
	return c.ref
}

func (c *Configuration) IsDefaultRemote() bool {
	return c.Remote() == defaultRemote
}

// Remote returns the default remote based on:
// 1. The currently tracked remote branch, if present
// 2. Any other SINGLE remote defined in .git/config
// 3. Use "origin" as a fallback.
// Results are cached after the first hit.
func (c *Configuration) Remote() string {
	ref := c.CurrentRef()

	c.loading.Lock()
	defer c.loading.Unlock()

	if c.currentRemote == nil {
		if len(ref.Name) == 0 {
			c.currentRemote = &defaultRemote
			return defaultRemote
		}

		if remote, ok := c.Git.Get(fmt.Sprintf("branch.%s.remote", ref.Name)); ok {
			// try tracking remote
			c.currentRemote = &remote
		} else if remotes := c.Remotes(); len(remotes) == 1 {
			// use only remote if there is only 1
			c.currentRemote = &remotes[0]
		} else {
			// fall back to default :(
			c.currentRemote = &defaultRemote
		}
	}
	return *c.currentRemote
}

func (c *Configuration) PushRemote() string {
	ref := c.CurrentRef()
	c.loading.Lock()
	defer c.loading.Unlock()

	if c.pushRemote == nil {
		if remote, ok := c.Git.Get(fmt.Sprintf("branch.%s.pushRemote", ref.Name)); ok {
			c.pushRemote = &remote
		} else if remote, ok := c.Git.Get("remote.pushDefault"); ok {
			c.pushRemote = &remote
		} else {
			c.loading.Unlock()
			remote := c.Remote()
			c.loading.Lock()

			c.pushRemote = &remote
		}
	}

	return *c.pushRemote
}

func (c *Configuration) SetValidRemote(name string) error {
	if err := git.ValidateRemote(name); err != nil {
		return err
	}
	c.SetRemote(name)
	return nil
}

func (c *Configuration) SetRemote(name string) {
	c.currentRemote = &name
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

func (c *Configuration) HookDir() string {
	if git.IsGitVersionAtLeast("2.9.0") {
		hp, ok := c.Git.Get("core.hooksPath")
		if ok {
			return hp
		}
	}
	return filepath.Join(c.LocalGitDir(), "hooks")
}

func (c *Configuration) InRepo() bool {
	return len(c.LocalGitDir()) > 0
}

func (c *Configuration) LocalWorkingDir() string {
	c.loadGitDirs()
	return c.workDir
}

func (c *Configuration) LocalGitDir() string {
	c.loadGitDirs()
	return *c.gitDir
}

func (c *Configuration) loadGitDirs() {
	c.loadingGit.Lock()
	defer c.loadingGit.Unlock()

	if c.gitDir != nil {
		return
	}

	gitdir, workdir, err := git.GitAndRootDirs()
	if err != nil {
		errMsg := err.Error()
		tracerx.Printf("Error running 'git rev-parse': %s", errMsg)
		if !strings.Contains(errMsg, "Not a git repository") {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
		}
		c.gitDir = &gitdir
	}

	gitdir = tools.ResolveSymlinks(gitdir)
	c.gitDir = &gitdir
	c.workDir = tools.ResolveSymlinks(workdir)
}

func (c *Configuration) LocalGitStorageDir() string {
	return c.Filesystem().GitStorageDir
}

func (c *Configuration) LocalReferenceDir() string {
	return c.Filesystem().ReferenceDir
}

func (c *Configuration) LFSStorageDir() string {
	return c.Filesystem().LFSStorageDir
}

func (c *Configuration) LFSObjectDir() string {
	return c.Filesystem().LFSObjectDir()
}

func (c *Configuration) LFSObjectExists(oid string, size int64) bool {
	return c.Filesystem().ObjectExists(oid, size)
}

func (c *Configuration) EachLFSObject(fn func(fs.Object) error) error {
	return c.Filesystem().EachObject(fn)
}

func (c *Configuration) LocalLogDir() string {
	return c.Filesystem().LogDir()
}

func (c *Configuration) TempDir() string {
	return c.Filesystem().TempDir()
}

func (c *Configuration) Filesystem() *fs.Filesystem {
	c.loadGitDirs()
	c.loading.Lock()
	defer c.loading.Unlock()

	if c.fs == nil {
		lfsdir, _ := c.Git.Get("lfs.storage")
		c.fs = fs.New(c.LocalGitDir(), c.LocalWorkingDir(), lfsdir)
	}

	return c.fs
}

func (c *Configuration) Cleanup() error {
	c.loading.Lock()
	defer c.loading.Unlock()
	return c.fs.Cleanup()
}

func (c *Configuration) OSEnv() Environment {
	return c.Os
}

func (c *Configuration) GitEnv() Environment {
	return c.Git
}

func (c *Configuration) GitConfig() *git.Configuration {
	return c.gitConfig
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

func (c *Configuration) SetGitLocalKey(key, val string) (string, error) {
	return c.gitConfig.SetLocal(key, val)
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

func (c *Configuration) UnsetGitLocalKey(key string) (string, error) {
	return c.gitConfig.UnsetLocalKey(key)
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
func (c *Configuration) loadGitConfig() {
	if g, ok := c.Git.(*delayedEnvironment); ok {
		g.Load()
	}
}

// CurrentCommitter returns the name/email that would be used to author a commit
// with this configuration. In particular, the "user.name" and "user.email"
// configuration values are used
func (c *Configuration) CurrentCommitter() (name, email string) {
	name, _ = c.Git.Get("user.name")
	email, _ = c.Git.Get("user.email")
	return
}
