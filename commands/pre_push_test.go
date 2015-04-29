package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPrePushWithEmptyQueue(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("pre-push", "--dry-run", "origin", "https://git-remote.com")
	cmd.Input = strings.NewReader("refs/heads/master master refs/heads/master 2206c37dddba83f58b1ada72709a6b60cf8b058e")
	cmd.Output = ""
}

func TestPrePushToMaster(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("pre-push", "--dry-run", "origin", "https://git-remote.com")
	cmd.Input = strings.NewReader("refs/heads/master master refs/heads/master 2206c37dddba83f58b1ada72709a6b60cf8b058e")
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
