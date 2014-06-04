package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	repo := NewRepository(t, "empty")
	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("config")
	SetConfigOutput(cmd, map[string]string{
		"Endpoint":        "https://example.com/git/media.git/info/media",
		"LocalWorkingDir": repo.Path,
		"LocalGitDir":     filepath.Join(repo.Path, ".git"),
		"LocalMediaDir":   filepath.Join(repo.Path, ".git", "media"),
		"TempDir":         filepath.Join(os.TempDir(), "git-media"),
	})

	repo.Test()
}

func TestConfigWithMediaUrl(t *testing.T) {
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
