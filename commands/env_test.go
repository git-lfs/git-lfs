package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEnv(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("env")
	SetConfigOutput(cmd, map[string]string{
		"Endpoint":        "https://example.com/git/media.git/info/media",
		"LocalWorkingDir": repo.Path,
		"LocalGitDir":     filepath.Join(repo.Path, ".git"),
		"LocalMediaDir":   filepath.Join(repo.Path, ".git", "media"),
		"TempDir":         filepath.Join(os.TempDir(), "git-media"),
	})
}

func TestEnvWithMediaUrl(t *testing.T) {
	repo := NewRepository(t, "config_media_url")
	defer repo.Test()

	cmd := repo.Command("env")
	SetConfigOutput(cmd, map[string]string{
		"Endpoint":        "http://foo/bar",
		"LocalWorkingDir": repo.Path,
		"LocalGitDir":     filepath.Join(repo.Path, ".git"),
		"LocalMediaDir":   filepath.Join(repo.Path, ".git", "media"),
		"TempDir":         filepath.Join(os.TempDir(), "git-media"),
	})
}

func TestEnvWithSubmoduleFromRepository(t *testing.T) {
	repo := NewRepository(t, "submodule")
	defer repo.Test()

	cmd := repo.Command("env")
	SetConfigOutput(cmd, map[string]string{
		"Endpoint":        "https://example.com/git/media.git/info/media",
		"LocalWorkingDir": repo.Path,
		"LocalGitDir":     filepath.Join(repo.Path, ".git"),
		"LocalMediaDir":   filepath.Join(repo.Path, ".git", "media"),
		"TempDir":         filepath.Join(os.TempDir(), "git-media"),
	})

	cmd.Before(func() {
		submodPath := filepath.Join(Root, "commands", "repos", "attributes.git")
		exec.Command("git", "config", "--add", "submodule.attributes.url", submodPath).Run()
		exec.Command("git", "submodule", "update").Run()
	})
}

func TestEnvWithSubmoduleFromSubmodule(t *testing.T) {
	repo := NewRepository(t, "submodule")
	defer repo.Test()

	repo.AddPath(repo.Path, "attributes", "a")
	repo.Paths = repo.Paths[1:]

	cmd := repo.Command("env")
	SetConfigOutput(cmd, map[string]string{
		// having trouble guessing the media endpoint from the submodule's origin
		// on the local filesystem.
		"Endpoint":        "unknown",
		"LocalWorkingDir": repo.Path,
		"LocalGitDir":     filepath.Join(repo.Path, ".git"),
		"LocalMediaDir":   filepath.Join(repo.Path, ".git", "media"),
		"TempDir":         filepath.Join(os.TempDir(), "git-media"),
	})

	cmd.Before(func() {
		submodPath := filepath.Join(Root, "commands", "repos", "attributes.git")
		exec.Command("git", "config", "--add", "submodule.attributes.url", submodPath).Run()
		exec.Command("git", "submodule", "update").Run()
		os.Chdir(filepath.Join("attributes", "a", "b"))
		exec.Command("git", "checkout", "-b", "whatevs").Run()
	})
}
