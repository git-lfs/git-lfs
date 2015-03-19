package commands

import (
	"bytes"
	"github.com/bmizerany/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestClean(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	content := "HI\n"
	oid := "f712374589a4f37f0fd6b941a104c7ccf43f68b1fdecb4d5cd88b80acbf98fc2"
	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd := repo.Command("clean", "somefile")
	cmd.Input = bytes.NewBufferString(content)
	cmd.Output = `version https://git-lfs.github.com/spec/v1
oid sha256:` + oid + `
size 3`

	cmd.After(func() {
		// assert hooks
		stat, err := os.Stat(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, stat.IsDir())
	})

	cmd = repo.Command("clean")
	cmd.Input = bytes.NewBufferString(content)
	cmd.Output = `version https://git-lfs.github.com/spec/v1
oid sha256:` + oid + `
size 3`
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
