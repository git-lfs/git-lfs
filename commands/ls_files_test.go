package commands

import (
	"path/filepath"
	"testing"
)

func TestLsFiles(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("ls-files")
	cmd.Output = "a.dat"

	cmd.Before(func() {
		path := filepath.Join(".git", "info", "attributes")
		repo.WriteFile(path, "*.dat filter=lfs -crlf\n")

		// Add a Git LFS object
		repo.WriteFile(filepath.Join(repo.Path, "a.dat"), "some data")
		repo.GitCmd("add", "a.dat")
		repo.GitCmd("commit", "-m", "a")

		// Add a regular file
		repo.WriteFile(filepath.Join(repo.Path, "hi.txt"), "some text")
		repo.GitCmd("add", "hi.txt")
		repo.GitCmd("commit", "-m", "hi")
	})
}
