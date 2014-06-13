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
		"Endpoint":        "unknown",
		"LocalWorkingDir": filepath.Join(repo.Path, "attributes", "a"),
		"LocalGitDir":     "../.git/modules/attributes",
		"LocalMediaDir":   "../.git/modules/attributes/media",
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

func TestEnvWithConfiguredSubmodule(t *testing.T) {
	repo := NewRepository(t, "submodule")
	defer repo.Test()

	repo.AddPath(repo.Path, "config_media_url")
	repo.Paths = repo.Paths[1:]

	cmd := repo.Command("env")
	SetConfigOutput(cmd, map[string]string{
		"Endpoint":        "http://foo/bar",
		"LocalWorkingDir": filepath.Join(repo.Path, "config_media_url"),
		"LocalGitDir":     "../.git/modules/config_media_url",
		"LocalMediaDir":   "../.git/modules/config_media_url/media",
		"TempDir":         filepath.Join(os.TempDir(), "git-media"),
	})

	cmd.Before(func() {
		submodPath := filepath.Join(Root, "commands", "repos", "config_media_url.git")
		exec.Command("git", "submodule", "add", submodPath).Run()
		exec.Command("git", "submodule", "update").Run()
		os.Chdir("config_media_url")
		exec.Command("git", "checkout", "-b", "whatevs").Run()
	})
}
