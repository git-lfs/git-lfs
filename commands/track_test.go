package commands

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmizerany/assert"
)

func TestTrack(t *testing.T) {
	repo := NewRepository(t, "attributes")
	defer repo.Test()

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")
	customHook := "echo 'yo'"

	cmd := repo.Command("track")
	cmd.Output = "Listing tracked paths\n" +
		"    *.mov (.git/info/attributes)\n" +
		"    *.jpg (.gitattributes)\n" +
		"    *.gif (a/.gitattributes)\n" +
		"    *.png (a/b/.gitattributes)"

	cmd.Before(func() {
		// write attributes file in .git
		path := filepath.Join(repo.Path, ".git", "info", "attributes")
		repo.WriteFile(path, "*.mov filter=lfs -crlf\n")

		// add hook
		repo.WriteFile(prePushHookFile, customHook)
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, string(customHook), string(by))
	})
}

func TestTrackOnEmptyRepository(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd := repo.Command("track", "*.gif")
	cmd.Output = "Tracking *.gif"

	cmd.Before(func() {
		// write attributes file in .git
		path := filepath.Join(repo.Path, ".gitattributes")
		repo.WriteFile(path, "*.mov filter=lfs diff=lfs merge=lfs -crlf\n")
	})

	cmd.After(func() {
		// assert path was added
		assert.Equal(t, "*.mov filter=lfs diff=lfs merge=lfs -crlf\n*.gif filter=lfs diff=lfs merge=lfs -crlf\n", repo.ReadFile(".gitattributes"))

		expected := "Listing tracked paths\n    *.mov (.gitattributes)\n    *.gif (.gitattributes)\n"

		assert.Equal(t, expected, repo.MediaCmd("track"))

		// assert hook was created
		stat, err := os.Stat(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, stat.IsDir())
	})

	cmd = repo.Command("track")
	cmd.Output = "Listing tracked paths"
}

func TestTrackWithoutTrailingLinebreak(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd := repo.Command("track", "*.gif")
	cmd.Output = "Tracking *.gif"

	cmd.Before(func() {
		repo.WriteFile(filepath.Join(repo.Path, ".gitattributes"), "*.mov filter=lfs -crlf")
	})

	cmd.After(func() {
		// assert path was added
		assert.Equal(t, "*.mov filter=lfs -crlf\n*.gif filter=lfs diff=lfs merge=lfs -crlf\n", repo.ReadFile(".gitattributes"))

		expected := "Listing tracked paths\n    *.mov (.gitattributes)\n    *.gif (.gitattributes)\n"

		assert.Equal(t, expected, repo.MediaCmd("track"))

		// assert hook was created
		stat, err := os.Stat(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, stat.IsDir())
	})

	cmd = repo.Command("track")
	cmd.Output = "Listing tracked paths"
}
