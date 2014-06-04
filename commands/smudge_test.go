package commands

import (
	"bytes"
	"github.com/bmizerany/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestSmudge(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("smudge")
	cmd.Input = bytes.NewBufferString("# git-media\nSOMEOID")
	cmd.Output = "whatever"

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd.Before(func() {
		path := filepath.Join(repo.Path, ".git", "media", "SO", "ME")
		file := filepath.Join(path, "SOMEOID")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(file, []byte("whatever\n"), 0755))
	})

	cmd.After(func() {
		// assert hooks
		stat, err := os.Stat(prePushHookFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, stat.IsDir())
	})
}
