package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

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

func (r *runner) Cd(path string) {
	fullPath := r.repoPath(path)
	r.Logf("$ cd %s", fullPath)
	if err := os.Chdir(fullPath); err != nil {
		r.Fatal(err)
	}
}

func (r *runner) MkdirP(path string) {
	fullPath := r.repoPath(path)
	r.Logf("$ mkdir -p %s", fullPath)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		r.Fatal(err)
	}
}

// Pointer builds a Git LFS pointer.
func Pointer(oid string, size int64) string {
	return fmt.Sprintf("version https://git-lfs.github.com/spec/v1\noid sha256:%s\nsize %d", oid, size)
}

// AssertPointerBlob ensures that the pointer of the given oid and size is
// committed to Git at the given commitish and path.
func AssertPointerBlob(run *runner, oid string, size int64, commitish, path string) {
	blob := run.GitBlob(commitish, path)
	AssertCommand(run.T,
		run.Git("cat-file", "-p", blob),
		Pointer(oid, size),
	)
}

// AssertString ensures that the given strings match.
func AssertString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("Expected %s, got %s", expected, actual)
	}
}

// AssertCommand checks the output of a command.  The expected value comes
// after the command, and does not need a trailing linebreak.
func AssertCommand(t *testing.T, output, expected string) {
	if expected != strings.TrimSpace(output) {
		t.Fatalf("Expected:\n%s\n\nGot:\n%s", expected, output)
	}
}

// AssertCommandContains checks that each given part is in the command's output.
func AssertCommandContains(t *testing.T, output string, parts ...string) {
	trimmed := strings.TrimSpace(output)
	failed := false

	for _, part := range parts {
		if !strings.Contains(trimmed, part) {
			failed = true
			t.Errorf("Expected %q", part)
		}
	}

	if failed {
		t.Fatalf("Got:\n%s", output)
	}
}

// RefuteServerObject ensures that the given object has not been uploaded to
// the current Git LFS server.
func RefuteServerObject(r *runner, oid string) {
	_, ok := r.repo().largeObjects[oid]
	if ok {
		r.Fatalf("object found: %s", oid)
	}
}

// AssertServerObject ensures that the given object has been sucessfully
// uploaded to the current Git LFS server.
func AssertServerObject(r *runner, oid string, contents []byte) {
	repo := r.repo()
	by, ok := repo.largeObjects[oid]
	if !ok {
		r.Fatalf("object not found: %s", oid)
	}

	AssertString(r.T, string(contents), string(by))
}
