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

func CatFile(sha1 string) (string, error) {
	return simpleExec(nil, "git", "cat-file", "-p", sha1)
}

func Grep(pattern string) (string, error) {
	return simpleExec(nil, "git", "grep", "--full-name", "--name-only", "--cached", pattern)
}

func HashObject(data []byte) (string, error) {
	buf := bytes.NewBuffer(data)
	return simpleExec(buf, "git", "hash-object", "--stdin")
}

var z40 = regexp.MustCompile(`\^?0{40}`)

type GitObject struct {
	Sha1 string
	Name string
}

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

func (c *gitConfig) Find(val string) string {
	output, _ := simpleExec(nil, "git", "config", val)
	return output
}

func (c *gitConfig) SetGlobal(key, val string) {
	simpleExec(nil, "git", "config", "--global", "--add", key, val)
}

func (c *gitConfig) UnsetGlobal(key string) {
	simpleExec(nil, "git", "config", "--global", "--unset", key)
}

func (c *gitConfig) List() (string, error) {
	return simpleExec(nil, "git", "config", "-l")
}

func (c *gitConfig) ListFromFile() (string, error) {
	return simpleExec(nil, "git", "config", "-l", "-f", ".gitconfig")
}

func (c *gitConfig) Version() (string, error) {
	return simpleExec(nil, "git", "version")
}

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
