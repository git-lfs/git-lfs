package commands

import (
	"github.com/bmizerany/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	repo := NewRepository(t, "attributes")
	defer repo.Test()

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")
	customHook := []byte("echo 'yo'")

	cmd := repo.Command("path")
	cmd.Output = "Listing paths\n" +
		"    *.mov (.git/info/attributes)\n" +
		"    *.jpg (.gitattributes)\n" +
		"    *.gif (a/.gitattributes)\n" +
		"    *.png (a/b/.gitattributes)"

	cmd.Before(func() {
		// write attributes file in .git
		path := filepath.Join(".git", "info", "attributes")
		repo.WriteFile(path, "*.mov filter=hawser -crlf\n")

		// add hook
		err := ioutil.WriteFile(prePushHookFile, customHook, 0755)
		assert.Equal(t, nil, err)
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, string(customHook), string(by))
	})
}

func TestPathOnEmptyRepository(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd := repo.Command("path", "add", "*.gif")
	cmd.Output = "Adding path *.gif"
	cmd.After(func() {
		// assert path was added
		assert.Equal(t, "*.gif filter=hawser -crlf\n", repo.ReadFile(".gitattributes"))

		expected := "Listing paths\n    *.gif (.gitattributes)\n"

		assert.Equal(t, expected, repo.MediaCmd("path"))

		// assert hook was created
		stat, err := os.Stat(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, stat.IsDir())
	})

	cmd = repo.Command("path")
	cmd.Output = "Listing paths"
}
