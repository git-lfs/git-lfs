package git

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rubyist/tracerx"
)

var Config = &Configuration{}

// Configuration can fetch or modify the current Git config and track the Git
// version.
type Configuration struct {
	WorkDir string
	GitDir  string
	version string
	mu      sync.Mutex
}

func NewConfig(workdir, gitdir string) *Configuration {
	if len(gitdir) == 0 && len(workdir) > 0 {
		gitdir = filepath.Join(workdir, ".git")
	}
	return &Configuration{WorkDir: workdir, GitDir: gitdir}
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
	output, _ := c.git("config", val)
	return output
}

// FindGlobal returns the git config value global scope for the key
func (c *Configuration) FindGlobal(key string) string {
	output, _ := c.git("config", "--global", key)
	return output
}

// FindSystem returns the git config value in system scope for the key
func (c *Configuration) FindSystem(key string) string {
	output, _ := c.git("config", "--system", key)
	return output
}

// Find returns the git config value for the key
func (c *Configuration) FindLocal(key string) string {
	output, _ := c.git("config", "--local", key)
	return output
}

// SetGlobal sets the git config value for the key in the global config
func (c *Configuration) SetGlobal(key, val string) (string, error) {
	return c.git("config", "--global", "--replace-all", key, val)
}

// SetSystem sets the git config value for the key in the system config
func (c *Configuration) SetSystem(key, val string) (string, error) {
	return c.git("config", "--system", "--replace-all", key, val)
}

// UnsetGlobalSection removes the entire named section from the global config
func (c *Configuration) UnsetGlobalSection(key string) (string, error) {
	return c.git("config", "--global", "--remove-section", key)
}

// UnsetSystemSection removes the entire named section from the system config
func (c *Configuration) UnsetSystemSection(key string) (string, error) {
	return c.git("config", "--system", "--remove-section", key)
}

// UnsetLocalSection removes the entire named section from the system config
func (c *Configuration) UnsetLocalSection(key string) (string, error) {
	return c.git("config", "--local", "--remove-section", key)
}

// SetLocal sets the git config value for the key in the specified config file
func (c *Configuration) SetLocal(file, key, val string) (string, error) {
	args := make([]string, 0, 5)
	if len(file) > 0 {
		args = append(args, "--file", file)
	}
	return c.git("config", append(args, "--replace-all", key, val)...)
}

// UnsetLocalKey removes the git config value for the key from the specified config file
func (c *Configuration) UnsetLocalKey(file, key string) (string, error) {
	args := make([]string, 0, 4)
	if len(file) > 0 {
		args = append(args, "--file", file)
	}
	return c.git("config", append(args, "--unset", key)...)
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

	out, err := c.git("config", "-l", "-f", filename)
	if err != nil {
		return nil, err
	}
	return ParseConfigLines(out, true), nil
}

func (c *Configuration) Source() (*ConfigurationSource, error) {
	out, err := c.git("config", "-l")
	if err != nil {
		return nil, err
	}
	return ParseConfigLines(out, false), nil
}

// Version returns the git version
func (c *Configuration) Version() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.version) == 0 {
		v, err := gitSimple("version")
		if err != nil {
			return v, err
		}
		c.version = v
	}

	return c.version, nil
}

// IsVersionAtLeast returns whether the git version is the one specified or higher
// argument is plain version string separated by '.' e.g. "2.3.1" but can omit minor/patch
func (c *Configuration) IsGitVersionAtLeast(ver string) bool {
	gitver, err := c.Version()
	if err != nil {
		tracerx.Printf("Error getting git version: %v", err)
		return false
	}
	return IsVersionAtLeast(gitver, ver)
}

func (c *Configuration) git(subcmd string, args ...string) (string, error) {
	cmd := make([]string, 1, len(args)+3)
	cmd[0] = subcmd

	if len(c.GitDir) > 0 {
		cmd = append(cmd, "--git-dir", c.GitDir)
	}
	if len(c.WorkDir) > 0 {
		cmd = append(cmd, "--work-tree", c.WorkDir)
	}

	return gitSimple(append(cmd, args...)...)
}
