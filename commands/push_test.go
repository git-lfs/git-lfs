package commands

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestPushWithEmptyQueue(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("push", "origin", "master")
	cmd.Output = ""
}

func TestPushToMaster(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("push", "--dry-run", "origin", "master")
	cmd.Output = "push a.dat"

	cmd.Before(func() {
		repo.GitCmd("remote", "remove", "origin")

		originPath := filepath.Join(Root, "commands", "repos", "empty.git")
		repo.GitCmd("remote", "add", "origin", originPath)

		repo.GitCmd("fetch")

		repo.WriteFile(filepath.Join(repo.Path, ".gitattributes"), "*.dat filter=lfs -crlf\n")

		// Add a Git LFS file
		repo.WriteFile(filepath.Join(repo.Path, "a.dat"), "some data")
		repo.GitCmd("add", "a.dat")
		repo.GitCmd("commit", "-m", "a")
	})
}

func TestPushToNewBranch(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("push", "--dry-run", "origin", "newbranch")
	cmd.Output = "push a.dat\npush b.dat"

	cmd.Before(func() {
		repo.GitCmd("remote", "remove", "origin")

		originPath := filepath.Join(Root, "commands", "repos", "empty.git")
		repo.GitCmd("remote", "add", "origin", originPath)

		repo.GitCmd("fetch")

		repo.WriteFile(filepath.Join(repo.Path, ".gitattributes"), "*.dat filter=lfs -crlf\n")
		repo.GitCmd("add", ".gitattributes")
		repo.GitCmd("commit", "-m", "attributes")

		// Add a Git LFS file
		repo.WriteFile(filepath.Join(repo.Path, "a.dat"), "some data")
		repo.GitCmd("add", "a.dat")
		repo.GitCmd("commit", "-m", "a")

		// Branch off
		repo.GitCmd("checkout", "-b", "newbranch")

		repo.WriteFile(filepath.Join(repo.Path, "b.dat"), "some more data")
		repo.GitCmd("add", "b.dat")
		repo.GitCmd("commit", "-m", "b")
	})

}

func TestPushStdin(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("push", "--stdin", "--dry-run", "origin", "https://git-remote.com")
	cmd.Input = strings.NewReader("refs/heads/master master refs/heads/master 2206c37dddba83f58b1ada72709a6b60cf8b058e")
	cmd.Output = ""

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")
	cmd.Before(func() {
		err := ioutil.WriteFile(prePushHookFile, []byte("#!/bin/sh\ngit lfs push --stdin \"$@\"\n"), 0755)
		if err != nil {
			t.Fatalf("Error writing pre-push in Before(): %s", err)
		}
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != "#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 0; }\ngit lfs pre-push \"$@\"\n" {
			t.Errorf("Unexpected pre-push hook:\n%s", string(by))
		}
	})
}

func TestPushStdinWithUnexpectedHook(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("push", "--stdin", "--dry-run", "origin", "https://git-remote.com")
	cmd.Input = strings.NewReader("refs/heads/master master refs/heads/master 2206c37dddba83f58b1ada72709a6b60cf8b058e")
	cmd.Output = ""

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")
	cmd.Before(func() {
		err := ioutil.WriteFile(prePushHookFile, []byte("sup\n"), 0755)
		if err != nil {
			t.Fatalf("Error writing pre-push in Before(): %s", err)
		}
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != "sup\n" {
			t.Errorf("Unexpected pre-push hook:\n%s", string(by))
		}
	})
}
