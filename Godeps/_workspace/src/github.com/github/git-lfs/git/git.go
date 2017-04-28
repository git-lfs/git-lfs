// Package git contains various commands that shell out to git
package git

import (
	"errors"
	"fmt"
	"github.com/rubyist/tracerx"
	"io"
	"os/exec"
	"strings"
)

func LsRemote(repo, refspec string) (string, error) {
	if repo == "" {
		return "", errors.New("repo required")
	}
	if refspec == "" {
		return simpleExec(nil, "git", "ls-remote", repo)

	}
	return simpleExec(nil, "git", "ls-remote", repo, refspec)
}

func CurrentRef() (string, error) {
	return simpleExec(nil, "git", "rev-parse", "HEAD")
}

func CurrentBranch() (string, error) {
	return simpleExec(nil, "git", "rev-parse", "--abbrev-ref", "HEAD")
}

func CurrentRemoteRef() (string, error) {
	remote, err := CurrentRemote()
	if err != nil {
		return "", err
	}

	return simpleExec(nil, "git", "rev-parse", remote)
}

func CurrentRemote() (string, error) {
	branch, err := CurrentBranch()
	if err != nil {
		return "", err
	}

	if branch == "HEAD" {
		return "", errors.New("not on a branch")
	}

	remote := Config.Find(fmt.Sprintf("branch.%s.remote", branch))
	if remote == "" {
		return "", errors.New("remote not found")
	}

	return remote + "/" + branch, nil
}

type gitConfig struct {
}

var Config = &gitConfig{}

// Find returns the git config value for the key
func (c *gitConfig) Find(val string) string {
	output, _ := simpleExec(nil, "git", "config", val)
	return output
}

// SetGlobal sets the git config value for the key in the global config
func (c *gitConfig) SetGlobal(key, val string) {
	simpleExec(nil, "git", "config", "--global", "--add", key, val)
}

// SetGlobal removes the git config value for the key from the global config
func (c *gitConfig) UnsetGlobal(key string) {
	simpleExec(nil, "git", "config", "--global", "--unset", key)
}

// List lists all of the git config values
func (c *gitConfig) List() (string, error) {
	return simpleExec(nil, "git", "config", "-l")
}

// ListFromFile lists all of the git config values in the given config file
func (c *gitConfig) ListFromFile() (string, error) {
	return simpleExec(nil, "git", "config", "-l", "-f", ".gitconfig")
}

// Version returns the git version
func (c *gitConfig) Version() (string, error) {
	return simpleExec(nil, "git", "version")
}

// simpleExec is a small wrapper around os/exec.Command. If the passed stdin
// is not nil it will be hooked up to the subprocess stdin.
func simpleExec(stdin io.Reader, name string, arg ...string) (string, error) {
	tracerx.Printf("run_command: '%s' %s", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)
	if stdin != nil {
		cmd.Stdin = stdin
	}

	output, err := cmd.Output()
	if _, ok := err.(*exec.ExitError); ok {
		return "", nil
	} else if err != nil {
		return fmt.Sprintf("Error running %s %s", name, arg), err
	}

	return strings.Trim(string(output), " \n"), nil
}
