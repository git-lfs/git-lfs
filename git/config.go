package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/subprocess"
)

var (
	ErrReadOnly = errors.New("configuration is read-only")
)

// Environment is a restricted version of config.Environment that only provides
// a single method.
type Environment interface {
	// Get is shorthand for calling `e.Fetcher.Get(key)`.
	Get(key string) (val string, ok bool)
}

// Configuration can fetch or modify the current Git config and track the Git
// version.
type Configuration struct {
	WorkDir  string
	GitDir   string
	version  *string
	readOnly bool
	mu       sync.Mutex
}

func NewConfig(workdir, gitdir string) *Configuration {
	if len(gitdir) == 0 && len(workdir) > 0 {
		gitdir = filepath.Join(workdir, ".git")
	}
	return &Configuration{WorkDir: workdir, GitDir: gitdir}
}

// NewReadOnlyConfig creates a new confguration that returns an error if an
// attempt to write to the configuration is made.
func NewReadOnlyConfig(workdir, gitdir string) *Configuration {
	cfg := NewConfig(workdir, gitdir)
	cfg.readOnly = true
	return cfg

}

func ParseConfigLines(lines string, onlySafeKeys bool) *ConfigurationSource {
	return &ConfigurationSource{
		Lines:        strings.Split(lines, "\n"),
		OnlySafeKeys: onlySafeKeys,
	}
}

type ConfigurationSource struct {
	Lines        []string
	OnlySafeKeys bool
}

// Find returns the git config value for the key
func (c *Configuration) Find(val string) string {
	output, _ := c.gitConfig(val)
	return output
}

// FindGlobal returns the git config value in global scope for the key
func (c *Configuration) FindGlobal(key string) string {
	output, _ := c.gitConfig("--global", key)
	return output
}

// FindSystem returns the git config value in system scope for the key
func (c *Configuration) FindSystem(key string) string {
	output, _ := c.gitConfig("--system", key)
	return output
}

// FindLocal returns the git config value in local scope for the key
func (c *Configuration) FindLocal(key string) string {
	output, _ := c.gitConfig("--local", key)
	return output
}

// FindWorktree returns the git config value in worktree or local scope for the key, depending on whether multiple worktrees are in use
func (c *Configuration) FindWorktree(key string) string {
	output, _ := c.gitConfig("--worktree", key)
	return output
}

// SetGlobal sets the git config value for the key in the global config
func (c *Configuration) SetGlobal(key, val string) (string, error) {
	return c.gitConfigWrite("--global", "--replace-all", key, val)
}

// SetSystem sets the git config value for the key in the system config
func (c *Configuration) SetSystem(key, val string) (string, error) {
	return c.gitConfigWrite("--system", "--replace-all", key, val)
}

// SetLocal sets the git config value for the key in the specified config file
func (c *Configuration) SetLocal(key, val string) (string, error) {
	return c.gitConfigWrite("--replace-all", key, val)
}

// SetWorktree sets the git config value for the key in the worktree or local config, depending on whether multiple worktrees are in use
func (c *Configuration) SetWorktree(key, val string) (string, error) {
	return c.gitConfigWrite("--worktree", "--replace-all", key, val)
}

// UnsetGlobalSection removes the entire named section from the global config
func (c *Configuration) UnsetGlobalSection(key string) (string, error) {
	return c.gitConfigWrite("--global", "--remove-section", key)
}

// UnsetSystemSection removes the entire named section from the system config
func (c *Configuration) UnsetSystemSection(key string) (string, error) {
	return c.gitConfigWrite("--system", "--remove-section", key)
}

// UnsetLocalSection removes the entire named section from the local config
func (c *Configuration) UnsetLocalSection(key string) (string, error) {
	return c.gitConfigWrite("--local", "--remove-section", key)
}

// UnsetWorktreeSection removes the entire named section from the worktree or local config, depending on whether multiple worktrees are in use
func (c *Configuration) UnsetWorktreeSection(key string) (string, error) {
	return c.gitConfigWrite("--worktree", "--remove-section", key)
}

// UnsetLocalKey removes the git config value for the key from the specified config file
func (c *Configuration) UnsetLocalKey(key string) (string, error) {
	return c.gitConfigWrite("--unset", key)
}

func (c *Configuration) Sources(dir string, optionalFilename string) ([]*ConfigurationSource, error) {
	gitconfig, err := c.Source()
	if err != nil {
		return nil, err
	}

	bare, err := IsBare()
	if err != nil {
		return nil, err
	}

	// First try to read from the working directory and then the index if
	// the file is missing from the working directory.
	var fileconfig *ConfigurationSource
	if !bare {
		fileconfig, err = c.FileSource(filepath.Join(dir, optionalFilename))
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			fileconfig, _ = c.RevisionSource(fmt.Sprintf(":%s", optionalFilename))
		}
	}
	if fileconfig == nil {
		fileconfig, _ = c.RevisionSource(fmt.Sprintf("HEAD:%s", optionalFilename))
	}

	configs := make([]*ConfigurationSource, 0, 2)
	if fileconfig != nil {
		configs = append(configs, fileconfig)
	}

	return append(configs, gitconfig), nil
}

func (c *Configuration) FileSource(filename string) (*ConfigurationSource, error) {
	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	out, err := c.gitConfig("-l", "-f", filename)
	if err != nil {
		return nil, err
	}
	return ParseConfigLines(out, true), nil
}

func (c *Configuration) RevisionSource(revision string) (*ConfigurationSource, error) {
	out, err := c.gitConfig("-l", "--blob", revision)
	if err != nil {
		return nil, err
	}
	return ParseConfigLines(out, true), nil
}

func (c *Configuration) Source() (*ConfigurationSource, error) {
	out, err := c.gitConfig("-l")
	if err != nil {
		return nil, err
	}
	return ParseConfigLines(out, false), nil
}

func (c *Configuration) gitConfig(args ...string) (string, error) {
	args = append([]string{"config", "--includes"}, args...)
	cmd := subprocess.ExecCommand("git", args...)
	if len(c.GitDir) > 0 {
		cmd.Dir = c.GitDir
	}
	return subprocess.Output(cmd)
}

func (c *Configuration) gitConfigWrite(args ...string) (string, error) {
	if c.readOnly {
		return "", ErrReadOnly
	}
	return c.gitConfig(args...)
}
