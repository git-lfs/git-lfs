// Package git contains various commands that shell out to git
package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

// CatFile is equivalent to `git cat-file -p <sha1>`
func CatFile(sha1 string) (string, error) {
	return simpleExec(nil, "git", "cat-file", "-p", sha1)
}

// Grep is equivalent to `git grep --full-name --name-only --cached <pattern>`
func Grep(pattern string) (string, error) {
	return simpleExec(nil, "git", "grep", "--full-name", "--name-only", "--cached", pattern)
}

// HashObject is equivalent to `git hash-object --stdin` where data is passed
// to stdin.
func HashObject(data []byte) (string, error) {
	buf := bytes.NewBuffer(data)
	return simpleExec(buf, "git", "hash-object", "--stdin")
}

var z40 = regexp.MustCompile(`\^?0{40}`)

type GitObject struct {
	Sha1 string
	Name string
}

// RevListObjects has two modes:
// When all is true, left and right are ignored, equivalent to `git rev-list --objects --all`
// When all is false, it is equivalent to `git rev-list --objects <left> <right>`
// If right is 40 0s, it is removed from the arguments.
func RevListObjects(left, right string, all bool) ([]*GitObject, error) {
	objects := make([]*GitObject, 0)

	refArgs := []string{"rev-list", "--objects"}
	if all {
		refArgs = append(refArgs, "--all")
	} else {
		if left == "" {
			return nil, errors.New("Left commit required")
		}

		refArgs = append(refArgs, left)

		if right != "" && !z40.MatchString(right) {
			refArgs = append(refArgs, right)
		}
	}

	output, err := simpleExec(nil, "git", refArgs...)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewBufferString(output))
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		objects = append(objects, &GitObject{line[0], strings.Join(line[1:len(line)], " ")})
	}

	return objects, nil
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
