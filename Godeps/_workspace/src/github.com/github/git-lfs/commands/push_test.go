package commands

import (
	"path/filepath"
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
