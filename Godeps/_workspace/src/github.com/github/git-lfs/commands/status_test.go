package commands

import (
	"path/filepath"
	"testing"
)

func TestStatus(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("status", "--porcelain")
	cmd.Output = " M file1.dat 9\nA  file2.dat 10\nA  file3.dat 10"

	cmd.Before(func() {
		path := filepath.Join(".git", "info", "attributes")
		repo.WriteFile(path, "*.dat filter=lfs -crlf\n")

		// Add a Git LFS file
		repo.WriteFile(filepath.Join(repo.Path, "file1.dat"), "some data")
		repo.GitCmd("add", "file1.dat")
		repo.GitCmd("commit", "-m", "a")
		repo.WriteFile(filepath.Join(repo.Path, "file1.dat"), "other data")

		repo.WriteFile(filepath.Join(repo.Path, "file2.dat"), "file2 data")
		repo.GitCmd("add", "file2.dat")

		repo.WriteFile(filepath.Join(repo.Path, "file3.dat"), "file3 data")
		repo.GitCmd("add", "file3.dat")
		repo.WriteFile(filepath.Join(repo.Path, "file3.dat"), "file3 other data")
	})
}
