package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLsFiles(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("ls-files")
	cmd.Output = "somefile"

	cmd.Before(func() {
		// Add a link file
		dir := filepath.Join(repo.Path, ".git", "media", "objects", "48")
		os.MkdirAll(dir, 0755)

		path := filepath.Join(dir, "baff6546c517fcd41b98413bb2b0bcbb8d6505")
		repo.WriteFile(path, "oid f712374589a4f37f0fd6b941a104c7ccf43f68b1fdecb4d5cd88b80acbf98fc2\nname somefile")
	})
}
