package commands

import (
	"github.com/bmizerany/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("init")
	cmd.Output = "git media initialized"

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd.After(func() {
		// assert media filter config
		configs := GlobalGitConfig(t)
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
		stat, err := os.Stat(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, stat.IsDir())
	})

	cmd = repo.Command("init")
	cmd.Output = "Hook already exists: pre-push\ngit media initialized"

	customHook := []byte("echo 'yo'")
	cmd.Before(func() {
		err := ioutil.WriteFile(prePushHookFile, customHook, 0755)
		assert.Equal(t, nil, err)
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, string(customHook), string(by))
	})
}
