package git

import (
	"errors"
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

// FindGlobal returns the git config value global scope for the key
func (c *Configuration) FindGlobal(key string) string {
	output, _ := c.gitConfig("--global", key)
	return output
}

// FindSystem returns the git config value in system scope for the key
func (c *Configuration) FindSystem(key string) string {
	output, _ := c.gitConfig("--system", key)
	return output
}

// Find returns the git config value for the key
func (c *Configuration) FindLocal(key string) string {
	output, _ := c.gitConfig("--local", key)
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

// UnsetGlobalSection removes the entire named section from the global config
func (c *Configuration) UnsetGlobalSection(key string) (string, error) {
	return c.gitConfigWrite("--global", "--remove-section", key)
}

// UnsetSystemSection removes the entire named section from the system config
func (c *Configuration) UnsetSystemSection(key string) (string, error) {
	return c.gitConfigWrite("--system", "--remove-section", key)
}

// UnsetLocalSection removes the entire named section from the system config
func (c *Configuration) UnsetLocalSection(key string) (string, error) {
	return c.gitConfigWrite("--local", "--remove-section", key)
}

// SetLocal sets the git config value for the key in the specified config file
func (c *Configuration) SetLocal(key, val string) (string, error) {
	return c.gitConfigWrite("--replace-all", key, val)
}

// UnsetLocalKey removes the git config value for the key from the specified config file
func (c *Configuration) UnsetLocalKey(key string) (string, error) {
	return c.gitConfigWrite("--unset", key)
}

func (c *Configuration) Sources(optionalFilename string) ([]*ConfigurationSource, error) {
	gitconfig, err := c.Source()
	if err != nil {
		return nil, err
	}

	fileconfig, err := c.FileSource(optionalFilename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
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

func (c *Configuration) Source() (*ConfigurationSource, error) {
	out, err := c.gitConfig("-l")
	if err != nil {
		return nil, err
	}
	return ParseConfigLines(out, false), nil
}

func (c *Configuration) gitConfig(args ...string) (string, error) {
	args = append([]string{"config"}, args...)
	subprocess.Trace("git", args...)
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
