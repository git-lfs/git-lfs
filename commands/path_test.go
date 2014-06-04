package commands

import (
	"github.com/bmizerany/assert"
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	repo := NewRepository(t, "attributes")
	defer repo.Test()

	cmd := repo.Command("path")
	cmd.Output = "Listing paths\n" +
		"    *.mov (.git/info/attributes)\n" +
		"    *.jpg (.gitattributes)\n" +
		"    *.gif (a/.gitattributes)\n" +
		"    *.png (a/b/.gitattributes)"

	cmd.Before(func() {
		path := filepath.Join(".git", "info", "attributes")
		repo.WriteFile(path, "*.mov filter=media -crlf\n")
	})
}

func TestPathOnEmptyRepository(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("path", "add", "*.gif")
	cmd.Output = "Adding path *.gif"
	cmd.After(func() {
		assert.Equal(t, "*.gif filter=media -crlf\n", repo.ReadFile(".gitattributes"))

		expected := "Listing paths\n    *.gif (.gitattributes)\n"

		assert.Equal(t, expected, repo.MediaCmd("path"))
	})

	cmd = repo.Command("path")
	cmd.Output = "Listing paths"
}
