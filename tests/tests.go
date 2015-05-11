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
	dir      string
	repos    map[string]*repo
	repoName string
	*testing.T
}

func (r *runner) execCmd(cmd *exec.Cmd) string {
	r.logCmd(cmd.Args[0], cmd.Args[1:len(cmd.Args)]...)

	out := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Run(); err != nil {
		r.Fatalf("%s\n\n%s", err, out.String())
	}

	return strings.TrimSpace(out.String())
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
	repo := r.repo()
	if strings.HasPrefix(cleaned, "/") {
		r.Fatalf("%q is not relative to %q", path, repo.dir)
	}

	return filepath.Join(repo.dir, cleaned)
}

func Setup(t *testing.T) *runner {
	t.Logf("working directory: %s", testDir)
	dir, err := ioutil.TempDir(testDir, "integration-test-")
	if err != nil {
		t.Fatal(err)
	}

	r := &runner{
		dir:   dir,
		repos: make(map[string]*repo),
		T:     t,
	}

	r.Logf("temp: %s", dir)
	r.InitRepo("repo")

	return r
}

func (r *runner) Teardown() {
	for _, repo := range r.repos {
		repo.Teardown()
	}

	if !r.T.Failed() {
		os.RemoveAll(r.dir)
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
