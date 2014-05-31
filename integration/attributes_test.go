package main

import (
	"path/filepath"
	"testing"
)

func TestAttributes(t *testing.T) {
	repo := NewRepository(t, "attributes")

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

	repo.Test()
}
