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

	cmd := repo.Command("smudge")
	cmd.Input = bytes.NewBufferString("# git-media\nSOMEOID")
	cmd.Output = "whatever"

	cmd.Before(func() {
		path := filepath.Join(repo.Path, ".git", "media", "SO", "ME")
		file := filepath.Join(path, "SOMEOID")
		assert.Equal(t, nil, os.MkdirAll(path, 0755))
		assert.Equal(t, nil, ioutil.WriteFile(file, []byte("whatever\n"), 0755))
	})

	repo.Test()
}
