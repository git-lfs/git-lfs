// Package config collects together all configuration settings
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package config

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"path/filepath"

	"github.com/git-lfs/git-lfs/errors"
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

// Storage configuration
type StorageConfig struct {
	LfsStorageDir string `git:"lfs.storage"`
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

	loading    sync.Mutex // guards initialization of gitConfig and remotes
	remotes    []string
	extensions map[string]Extension
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
	Git, Os map[string][]string
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
func Unmarshal(git, os Enviornment, v interface{}) error {
	into := reflect.ValueOf(v)
	if into.Kind() != reflect.Ptr {
		return fmt.Errorf("lfs/config: unable to parse non-pointer type of %T", v)
	}
	into = into.Elem()

	for i := 0; i < into.Type().NumField(); i++ {
		field := into.Field(i)
		sfield := into.Type().Field(i)

		lookups, err := c.parseTag(sfield.Tag)
		if err != nil {
			return err
		}

		var val interface{}
		for _, lookup := range lookups {
			if _, ok := lookup.Get(); !ok {
				continue
			}

			switch sfield.Type.Kind() {
			case reflect.String:
				val, _ = lookup.Get()
			case reflect.Int:
				val = lookup.Int(int(field.Int()))
			case reflect.Bool:
				val = lookup.Bool(field.Bool())
			default:
				return fmt.Errorf("lfs/config: unsupported target type for field %q: %v",
					sfield.Name, sfield.Type.String())
			}

			if val != nil {
				break
			}
		}

		if val != nil {
			into.Field(i).Set(reflect.ValueOf(val))
		}
	}

	return nil
}

func (c *Configuration) Unmarshal(v interface{}) error {
	return Unmarshal(c.Git, c.Os, v)
}

var (
	tagRe    = regexp.MustCompile("((\\w+:\"[^\"]*\")\\b?)+")
	emptyEnv = EnvironmentOf(MapFetcher(nil))
)

type lookup struct {
	key string
	env Environment
}

func (l *lookup) Get() (interface{}, bool) { return l.env.Get(l.key) }
func (l *lookup) Int(or int) int           { return l.env.Int(l.key, or) }
func (l *lookup) Bool(or bool) bool        { return l.env.Bool(l.key, or) }

// parseTag returns the key, environment, and optional error assosciated with a
// given tag. It will return the XOR of either the `git` or `os` tag. That is to
// say, a field tagged with EITHER `git` OR `os` is valid, but pone tagged with
// both is not.
//
// If neither field was found, then a nil environment will be returned.
func (c *Configuration) parseTag(tag reflect.StructTag) ([]*lookup, error) {
	var lookups []*lookup

	parts := tagRe.FindAllString(string(tag), -1)
	for _, part := range parts {
		sep := strings.SplitN(part, ":", 2)
		if len(sep) != 2 {
			return nil, errors.Errorf("config: invalid struct tag %q", tag)
		}

		var env Environment
		switch strings.ToLower(sep[0]) {
		case "git":
			env = c.Git
		case "os":
			env = c.Os
		default:
			// ignore other struct tags, like `json:""`, etc.
			env = emptyEnv
		}

		uq, err := strconv.Unquote(sep[1])
		if err != nil {
			return nil, err
		}

		lookups = append(lookups, &lookup{
			key: uq,
			env: env,
		})
	}

	return lookups, nil
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

func (c *Configuration) StorageConfig() StorageConfig {
	s := &StorageConfig{
		LfsStorageDir: "lfs",
	}

	if err := c.Unmarshal(s); err != nil {
		panic(err.Error())
	}
	if !filepath.IsAbs(s.LfsStorageDir) {
		s.LfsStorageDir = filepath.Join(LocalGitStorageDir, s.LfsStorageDir)
	}
	return *s
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
