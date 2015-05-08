package tests

import (
	"fmt"
	"strings"
	"testing"
)

func Pointer(oid string, size int64) string {
	return fmt.Sprintf("version https://git-lfs.github.com/spec/v1\noid sha256:%s\nsize %d", oid, size)
}

func AssertPointerBlob(run *runner, oid string, size int64, commitish, path string) {
	blob := run.GitBlob(commitish, path)
	AssertCommand(run.T,
		run.Git("cat-file", "-p", blob),
		Pointer(oid, size),
	)
}

func AssertString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("Expected %s, got %s", expected, actual)
	}
}

func AssertCommand(t *testing.T, output, expected string) {
	if expected != strings.TrimSpace(output) {
		t.Fatalf("Expected:\n%s\n\nGot:\n%s", expected, output)
	}
}

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
