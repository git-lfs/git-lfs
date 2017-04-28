package commands

import (
	"fmt"
	"github.com/bmizerany/assert"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	Root         string
	Bin          string
	TempDir      string
	GitEnv       []string
	JoinedGitEnv string
	configKeys   = []string{"Endpoint", "LocalWorkingDir", "LocalGitDir", "LocalMediaDir", "TempDir"}
)

func NewRepository(t *testing.T, name string) *Repository {
	path := filepath.Join(TempDir, name)
	r := &Repository{
		T:        t,
		Name:     name,
		Path:     path,
		Paths:    []string{path},
		Commands: make([]*TestCommand, 0),
	}
	r.clone()
	r.Path = expand(path)
	return r
}

func AssertIncludeString(t *testing.T, expected string, actual []string) {
	found := false
	for _, line := range actual {
		if line == expected {
			found = true
		}
	}
	assert.Tf(t, found, "%s not included.", expected)
}

func GlobalGitConfig(t *testing.T) []string {
	o := cmd(t, "git", "config", "-l", "--global")
	return strings.Split(o, "\n")
}

func SetConfigOutput(c *TestCommand, keys map[string]string) {
	pieces := make([]string, 0, len(keys))

	for _, key := range configKeys {
		if v, ok := keys[key]; ok {
			pieces = append(pieces, key+"="+v)
		}
	}

	c.Output = strings.Join(pieces, "\n")

	if len(JoinedGitEnv) > 0 {
		c.Output += "\n" + JoinedGitEnv
	}
}

type Repository struct {
	T                *testing.T
	Name             string
	Path             string
	Paths            []string
	Commands         []*TestCommand
	expandedTempPath bool
}

func (r *Repository) AddPath(paths ...string) {
	r.Paths = append(r.Paths, filepath.Join(paths...))
}

func (r *Repository) Command(args ...string) *TestCommand {
	cmd := &TestCommand{
		T:               r.T,
		Args:            args,
		BeforeCallbacks: make([]func(), 0),
		AfterCallbacks:  make([]func(), 0),
		Env:             make([]string, 0),
	}
	r.Commands = append(r.Commands, cmd)
	return cmd
}

func (r *Repository) ReadFile(paths ...string) string {
	args := make([]string, 1, len(paths)+1)
	args[0] = r.Path
	args = append(args, paths...)
	by, err := ioutil.ReadFile(filepath.Join(args...))
	assert.Equal(r.T, nil, err)
	return string(by)
}

func (r *Repository) WriteFile(filename, output string) {
	r.e(ioutil.WriteFile(filename, []byte(output), 0755))
}

func (r *Repository) MediaCmd(args ...string) string {
	return r.cmd(Bin, args...)
}

func (r *Repository) GitCmd(args ...string) string {
	return r.cmd("git", args...)
}

func (r *Repository) Test() {
	for _, path := range r.Paths {
		r.test(path)
	}
}

func (r *Repository) test(path string) {
	fmt.Println("Command tests for\n", path)
	for _, cmd := range r.Commands {
		r.clone()
		cmd.Run(path)
	}
}

func (r *Repository) clone() {
	clone(r.T, r.Name, r.Path)
}

func (r *Repository) e(err error) {
	e(r.T, err)
}

func (r *Repository) cmd(name string, args ...string) string {
	return cmd(r.T, name, args...)
}

type TestCommand struct {
	T               *testing.T
	Args            []string
	Env             []string
	Input           io.Reader
	Output          string
	BeforeCallbacks []func()
	AfterCallbacks  []func()
}

func (c *TestCommand) Run(path string) {
	fmt.Println("$ git lfs", strings.Join(c.Args, " "))

	for _, cb := range c.BeforeCallbacks {
		cb()
	}

	c.e(os.Chdir(path))

	cmd := exec.Command(Bin, c.Args...)
	cmd.Stdin = c.Input
	if c.Env != nil && len(c.Env) > 0 {
		cmd.Env = append(os.Environ(), c.Env...)
	}
	outputBytes, err := cmd.CombinedOutput()
	c.e(err)

	if len(c.Output) > 0 {
		assert.Equal(c.T, c.Output+"\n", string(outputBytes))
	}

	for _, cb := range c.AfterCallbacks {
		cb()
	}
}

func (c *TestCommand) Before(f func()) {
	c.BeforeCallbacks = append(c.BeforeCallbacks, f)
}

func (c *TestCommand) After(f func()) {
	c.AfterCallbacks = append(c.AfterCallbacks, f)
}

func (c *TestCommand) e(err error) {
	e(c.T, err)
}

func cmd(t *testing.T, name string, args ...string) string {
	cmd := exec.Command(name, args...)
	o, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf(
			"Error running command:\n$ %s\n\n%s",
			strings.Join(cmd.Args, " "),
			string(o),
		)
	}
	return string(o)
}

func e(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err.Error())
	}
}

func expand(path string) string {
	p, err := filepath.EvalSymlinks(path)
	if err != nil {
		panic(err)
	}
	return p
}

func clone(t *testing.T, name, path string) {
	e(t, os.RemoveAll(path))

	reposPath := filepath.Join(Root, "commands", "repos")
	e(t, os.Chdir(reposPath))
	cmd(t, "git", "clone", name, path)
	e(t, os.Chdir(path))
	cmd(t, "git", "remote", "remove", "origin")
	cmd(t, "git", "remote", "add", "origin", "https://example.com/git/lfs")
}

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	Root = filepath.Join(wd, "..")
	Bin = filepath.Join(Root, "bin", "git-lfs")
	TempDir = filepath.Join(os.TempDir(), "git-lfs-tests")

	env := os.Environ()
	GitEnv = make([]string, 0, len(env))
	for _, e := range env {
		if !strings.Contains(e, "GIT_") {
			continue
		}
		GitEnv = append(GitEnv, e)
	}
	JoinedGitEnv = strings.Join(GitEnv, "\n")
}
