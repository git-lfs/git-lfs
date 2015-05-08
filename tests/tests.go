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

var (
	rootDir string
	testDir string
	bin     string
)

type runner struct {
	tempDir string
	repoDir string
	*testing.T
}

// Git executes a Git command.
func (r *runner) Git(args ...string) string {
	return r.exec("git", args...)
}

func (r *runner) GitBlob(commitish, path string) string {
	out := r.Git("ls-tree", commitish)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		tabs := strings.Split(line, "\t")
		if len(tabs) < 2 {
			continue
		}

		attrs := strings.Split(tabs[0], " ")
		if len(attrs) < 3 {
			continue
		}

		if tabs[1] == path {
			return attrs[2]
		}
	}

	return ""
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

func (r *runner) exec(name string, args ...string) string {
	if name == "git" && len(args) > 0 && args[0] == "lfs" {
		name = bin
		args = args[1:len(args)]
	}

	r.logCmd(name, args...)

	cmd := exec.Command(name, args...)
	out := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Run(); err != nil {
		r.Fatalf("%s\n\n%s", err, out.String())
	}

	return out.String()
}

func (r *runner) logCmd(name string, args ...string) {
	loggedArgs := make([]string, len(args)+1)
	if strings.HasPrefix(name, "/") {
		rel, err := filepath.Rel(rootDir, name)
		if err == nil {
			loggedArgs[0] = rel
		} else {
			r.Errorf("Cannot make %q relative to %q", name, rootDir)
			loggedArgs[0] = name
		}
	} else {
		loggedArgs[0] = name
	}

	for idx, arg := range args {
		if strings.Contains(arg, " ") {
			arg = fmt.Sprintf(`"%s"`, arg)
		}
		loggedArgs[idx+1] = arg
	}

	r.Logf("$ %s", strings.Join(loggedArgs, " "))
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
		fmt.Println(err)
		os.Exit(1)
	}

	rootDir = filepath.Join(wd, "..")

	testDir = filepath.Join(rootDir, "tmp")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bin = filepath.Join(rootDir, "bin", "git-lfs")
	if _, err := os.Stat(bin); err != nil {
		fmt.Println("git-lfs is not compiled to " + bin)
		os.Exit(1)
	}
}
