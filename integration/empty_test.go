package main

import (
	"fmt"
	"github.com/bmizerany/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmptyRepository(t *testing.T) {
	repo := NewRepository(t, "empty")
	repo.AddPath(filepath.Join(repo.Path, ".git"))
	repo.AddPath(filepath.Join(repo.Path, "subdir"))

	cmd := repo.Command("version")
	cmd.Output = fmt.Sprintf("git-media v%s", Version)

	cmd = repo.Command("version", "-comics")
	cmd.Output = fmt.Sprintf("git-media v%s\nNothing may see Gah Lak Tus and survive.", Version)

	cmd = repo.Command("config")
	SetConfigOutput(cmd, map[string]string{
		"Endpoint":        "https://example.com/git/media.git/info/media",
		"LocalWorkingDir": repo.Path,
		"LocalGitDir":     filepath.Join(repo.Path, ".git"),
		"LocalMediaDir":   filepath.Join(repo.Path, ".git", "media"),
		"TempDir":         filepath.Join(os.TempDir(), "git-media"),
	})

	cmd = repo.Command("init")
	cmd.Output = "Installing clean filter\n" +
		"Installing smudge filter\n" +
		"git media initialized"

	cmd.After(func() {
		// assert media filter config
		configs := GlobalGitConfig(t)
		fmt.Println(configs)
		AssertIncludeString(t, "filter.media.clean=git media clean %f", configs)
		AssertIncludeString(t, "filter.media.smudge=git media smudge %f", configs)
		AssertIncludeString(t, "filter.media.required=true", configs)
		found := 0
		for _, line := range configs {
			if strings.HasPrefix(line, "filter.media") {
				found += 1
			}
		}
		assert.Equal(t, 3, found)

		// assert hooks
		prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")
		stat, err := os.Stat(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, stat.IsDir())
	})

	repo.Test()
}

func TestAttributesOnEmptyRepository(t *testing.T) {
	repo := NewRepository(t, "empty")

	cmd := repo.Command("path", "add", "*.gif")
	cmd.Output = "Adding path *.gif"
	cmd.After(func() {
		assert.Equal(t, "*.gif filter=media -crlf\n", repo.ReadFile(".gitattributes"))

		expected := "Listing paths\n    *.gif (.gitattributes)\n"

		assert.Equal(t, expected, repo.MediaCmd("path"))
	})

	cmd = repo.Command("path")
	cmd.Output = "Listing paths"

	repo.Test()
}
