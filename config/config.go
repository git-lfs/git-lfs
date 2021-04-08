// Package config collects together all configuration settings
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

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
	mask       int
	maskOnce   sync.Once
	timestamp  time.Time
}

func New() *Configuration {
	return NewIn("", "")
}

func NewIn(workdir, gitdir string) *Configuration {
	gitConf := git.NewConfig(workdir, gitdir)
	c := &Configuration{
		Os:        EnvironmentOf(NewOsFetcher()),
		gitConfig: gitConf,
		timestamp: time.Now(),
	}

	if len(gitConf.WorkDir) > 0 {
		c.gitDir = &gitConf.GitDir
		c.workDir = gitConf.WorkDir
	}

	c.Git = &delayedEnvironment{
		callback: func() Environment {
			sources, err := gitConf.Sources(c.LocalWorkingDir(), ".lfsconfig")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading git config: %s\n", err)
			}
			return c.readGitConfig(sources...)
		},
	}
	return c
}

func (c *Configuration) getMask() int {
	// This logic is necessarily complex because Git's logic is complex.
	c.maskOnce.Do(func() {
		val, ok := c.Git.Get("core.sharedrepository")
		if !ok {
			val = "umask"
		} else if Bool(val, false) {
			val = "group"
		}

		switch strings.ToLower(val) {
		case "group", "true", "1":
			c.mask = 007
		case "all", "world", "everybody", "2":
			c.mask = 002
		case "umask", "false", "0":
			c.mask = umask()
		default:
			if mode, err := strconv.ParseInt(val, 8, 16); err != nil {
				// If this doesn't look like an octal number, then it
				// could be a falsy value, in which case we should use
				// the umask, or it's just invalid, in which case the
				// umask is a safe bet.
				c.mask = umask()
			} else {
				c.mask = 0666 & ^int(mode)
			}
		}
	})
	return c.mask
}

func (c *Configuration) readGitConfig(gitconfigs ...*git.ConfigurationSource) Environment {
	gf, extensions, uniqRemotes := readGitConfig(gitconfigs...)
	c.extensions = extensions
	c.remotes = make([]string, 0, len(uniqRemotes))
	for remote := range uniqRemotes {
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
// and Environment-level values from the ones provided instead of the actual
// `.gitconfig` file or `os.Getenv`, respectively.
//
// This method should only be used during testing.
func NewFrom(v Values) *Configuration {
	c := &Configuration{
		Os:        EnvironmentOf(mapFetcher(v.Os)),
		gitConfig: git.NewConfig("", ""),
		timestamp: time.Now(),
	}
	c.Git = &delayedEnvironment{
		callback: func() Environment {
			source := &git.ConfigurationSource{
				Lines: make([]string, 0, len(v.Git)),
			}

			for key, values := range v.Git {
				parts := strings.Split(key, ".")
				isCaseSensitive := len(parts) >= 3
				hasUpper := strings.IndexFunc(key, unicode.IsUpper) > -1

				// This branch should only ever trigger in
				// tests, and only if they'd be broken.
				if !isCaseSensitive && hasUpper {
					panic(fmt.Sprintf("key %q has uppercase, shouldn't", key))
				}
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
// 2. The value of remote.lfsdefault.
// 3. Any other SINGLE remote defined in .git/config
// 4. Use "origin" as a fallback.
// Results are cached after the first hit.
func (c *Configuration) Remote() string {
	ref := c.CurrentRef()

	c.loading.Lock()
	defer c.loading.Unlock()

	if c.currentRemote == nil {
		if remote, ok := c.Git.Get(fmt.Sprintf("branch.%s.remote", ref.Name)); len(ref.Name) != 0 && ok {
			// try tracking remote
			c.currentRemote = &remote
		} else if remote, ok := c.Git.Get("remote.lfsdefault"); ok {
			// try default remote
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
		} else if remote, ok := c.Git.Get("remote.lfspushdefault"); ok {
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
		name := git.RewriteLocalPathAsURL(name)
		if err := git.ValidateRemote(name); err != nil {
			return err
		}
	}
	c.SetRemote(name)
	return nil
}

func (c *Configuration) SetValidPushRemote(name string) error {
	if err := git.ValidateRemote(name); err != nil {
		name := git.RewriteLocalPathAsURL(name)
		if err := git.ValidateRemote(name); err != nil {
			return err
		}
	}
	c.SetPushRemote(name)
	return nil
}

func (c *Configuration) SetRemote(name string) {
	c.currentRemote = &name
}

func (c *Configuration) SetPushRemote(name string) {
	c.pushRemote = &name
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

func (c *Configuration) ForceProgress() bool {
	return c.Os.Bool("GIT_LFS_FORCE_PROGRESS", false) || c.Git.Bool("lfs.forceprogress", false)
}

// HookDir returns the location of the hooks owned by this repository. If the
// core.hooksPath configuration variable is supported, we prefer that and expand
// paths appropriately.
func (c *Configuration) HookDir() (string, error) {
	if git.IsGitVersionAtLeast("2.9.0") {
		hp, ok := c.Git.Get("core.hooksPath")
		if ok {
			path, err := tools.ExpandPath(hp, false)
			if err != nil {
				return "", err
			}
			if filepath.IsAbs(path) {
				return path, nil
			}
			return filepath.Join(c.LocalWorkingDir(), path), nil
		}
	}
	return filepath.Join(c.LocalGitStorageDir(), "hooks"), nil
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
		if !strings.Contains(strings.ToLower(errMsg),
			"not a git repository") {
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

func (c *Configuration) LocalReferenceDirs() []string {
	return c.Filesystem().ReferenceDirs
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
		c.fs = fs.New(
			c.Os,
			c.LocalGitDir(),
			c.LocalWorkingDir(),
			lfsdir,
			c.RepositoryPermissions(false),
		)
	}

	return c.fs
}

func (c *Configuration) Cleanup() error {
	if c == nil {
		return nil
	}
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

func (c *Configuration) FindGitWorktreeKey(key string) string {
	return c.gitConfig.FindWorktree(key)
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

func (c *Configuration) SetGitWorktreeKey(key, val string) (string, error) {
	return c.gitConfig.SetWorktree(key, val)
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

func (c *Configuration) UnsetGitWorktreeSection(key string) (string, error) {
	return c.gitConfig.UnsetWorktreeSection(key)
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

var (
	// dateFormats is a list of all the date formats that Git accepts,
	// except for the built-in one, which is handled below.
	dateFormats = []string{
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"2006-01-02T15:04:05-0700",
		"2006-01-02 15:04:05-0700",
		"2006.01.02T15:04:05-0700",
		"2006.01.02 15:04:05-0700",
		"01/02/2006T15:04:05-0700",
		"01/02/2006 15:04:05-0700",
		"02.01.2006T15:04:05-0700",
		"02.01.2006 15:04:05-0700",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05Z",
		"2006.01.02T15:04:05Z",
		"2006.01.02 15:04:05Z",
		"01/02/2006T15:04:05Z",
		"01/02/2006 15:04:05Z",
		"02.01.2006T15:04:05Z",
		"02.01.2006 15:04:05Z",
	}

	// defaultDatePattern is the regexp for Git's native date format.
	defaultDatePattern = regexp.MustCompile(`\A(\d+) ([+-])(\d{2})(\d{2})\z`)
)

// findUserData returns the name/email that should be used in the commit header.
// We use the same technique as Git for finding this information, except that we
// don't fall back to querying the system for defaults if no values are found in
// the Git configuration or environment.
//
// envType should be "author" or "committer".
func (c *Configuration) findUserData(envType string) (name, email string) {
	var filter = func(r rune) rune {
		switch r {
		case '<', '>', '\n':
			return -1
		default:
			return r
		}
	}

	envType = strings.ToUpper(envType)

	name, ok := c.Os.Get("GIT_" + envType + "_NAME")
	if !ok {
		name, _ = c.Git.Get("user.name")
	}

	email, ok = c.Os.Get("GIT_" + envType + "_EMAIL")
	if !ok {
		email, ok = c.Git.Get("user.email")
	}
	if !ok {
		email, _ = c.Os.Get("EMAIL")
	}

	// Git filters certain characters out of the name and email fields.
	name = strings.Map(filter, name)
	email = strings.Map(filter, email)
	return
}

func (c *Configuration) findUserTimestamp(envType string) time.Time {
	date, ok := c.Os.Get(fmt.Sprintf("GIT_%s_DATE", strings.ToUpper(envType)))
	if !ok {
		return c.timestamp
	}

	// time.Parse doesn't parse seconds from the Epoch, like we use in the
	// Git native format, so we have to do it ourselves.
	strs := defaultDatePattern.FindStringSubmatch(date)
	if strs != nil {
		unixSecs, _ := strconv.ParseInt(strs[1], 10, 64)
		hours, _ := strconv.Atoi(strs[3])
		offset, _ := strconv.Atoi(strs[4])
		offset = (offset + hours*60) * 60
		if strs[2] == "-" {
			offset = -offset
		}

		return time.Unix(unixSecs, 0).In(time.FixedZone("", offset))
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, date); err == nil {
			return t
		}
	}

	// The user provided an invalid value, so default to the current time.
	return c.timestamp
}

// CurrentCommitter returns the name/email that would be used to commit a change
// with this configuration. In particular, the "user.name" and "user.email"
// configuration values are used
func (c *Configuration) CurrentCommitter() (name, email string) {
	return c.findUserData("committer")
}

// CurrentCommitterTimestamp returns the timestamp that would be used to commit
// a change with this configuration.
func (c *Configuration) CurrentCommitterTimestamp() time.Time {
	return c.findUserTimestamp("committer")
}

// CurrentAuthor returns the name/email that would be used to author a change
// with this configuration. In particular, the "user.name" and "user.email"
// configuration values are used
func (c *Configuration) CurrentAuthor() (name, email string) {
	return c.findUserData("author")
}

// CurrentCommitterTimestamp returns the timestamp that would be used to commit
// a change with this configuration.
func (c *Configuration) CurrentAuthorTimestamp() time.Time {
	return c.findUserTimestamp("author")
}

// RepositoryPermissions returns the permissions that should be used to write
// files in the repository.
func (c *Configuration) RepositoryPermissions(executable bool) os.FileMode {
	perms := os.FileMode(0666 & ^c.getMask())
	if executable {
		return tools.ExecutablePermissions(perms)
	}
	return perms
}
