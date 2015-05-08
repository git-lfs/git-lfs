package tests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var testDir string

type runner struct {
	tempDir string
	repoDir string
	*testing.T
}

// Git executes a Git command.
func (r *runner) Git(args ...string) {
	r.exec("git", args...)
}

// WriteFile writes the byte contents to the given path, which should be
// relative to this test's repository directory.
func (r *runner) WriteFile(path string, contents []byte) {
	err := ioutil.WriteFile(r.repoPath(path), contents, 0755)
	if err != nil {
		r.Fatal(err)
	}
}

// ReadFile reads the byte contents of the given path, which should be relative
// to this test's repository directory.
func (r *runner) ReadFile(path string) []byte {
	by, err := ioutil.ReadFile(r.repoPath(path))
	if err != nil {
		r.Fatal(err)
	}
	return by
}

func (r *runner) exec(name string, args ...string) {
	loggedArgs := make([]string, len(args)+1)
	loggedArgs[0] = name
	for idx, arg := range args {
		if strings.Contains(arg, " ") {
			arg = fmt.Sprintf(`"%s"`, arg)
		}
		loggedArgs[idx+1] = arg
	}

	r.Logf("$ %s", strings.Join(loggedArgs, " "))

	cmd := exec.Command(name, args...)
	out := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Run(); err != nil {
		r.Fatalf("%s\n\n%s", err, out.String())
	}
}

func (r *runner) repoPath(path string) string {
	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, ".") || strings.HasPrefix(cleaned, "/") {
		r.Fatalf("%q is not relative to %q", path, r.repoDir)
	}

	return filepath.Join(r.repoDir, cleaned)
}

func Setup(t *testing.T) *runner {
	t.Parallel()
	t.Logf("working directory: %s", testDir)
	dir, err := ioutil.TempDir(testDir, "integration-test-")
	if err != nil {
		t.Fatal(err)
	}

	r := &runner{
		tempDir: dir,
		repoDir: filepath.Join(dir, "repo"),
		T:       t,
	}

	if err := os.MkdirAll(r.repoDir, 0777); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(r.repoDir); err != nil {
		t.Fatal(err)
	}

	t.Logf("temp: %s", r.tempDir)
	r.Git("init")

	return r
}

func (r *runner) Teardown() {
	if !r.T.Failed() {
		os.RemoveAll(r.tempDir)
	}
}

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	testDir = filepath.Join(wd, "..", "tmp")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		panic(err)
	}
}
