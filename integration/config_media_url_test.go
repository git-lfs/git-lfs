package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMediaUrlConfig(t *testing.T) {
	repo := NewRepository(t, "config_media_url")

	cmd := repo.Command("config")
	SetConfigOutput(cmd, map[string]string{
		"Endpoint":        "http://foo/bar",
		"LocalWorkingDir": repo.Path,
		"LocalGitDir":     filepath.Join(repo.Path, ".git"),
		"LocalMediaDir":   filepath.Join(repo.Path, ".git", "media"),
		"TempDir":         filepath.Join(os.TempDir(), "git-media"),
	})

	repo.Test()
}
