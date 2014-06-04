package commands

import (
	"bytes"
	"github.com/bmizerany/assert"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestClean(t *testing.T) {
	repo := NewRepository(t, "empty")

	content := "HI\n"
	oid := "f712374589a4f37f0fd6b941a104c7ccf43f68b1fdecb4d5cd88b80acbf98fc2"

	cmd := repo.Command("clean")
	cmd.Input = bytes.NewBufferString(content)
	cmd.Output = "# git-media\n" + oid

	cmd.After(func() {
		queueDir := filepath.Join(repo.Path, ".git", "media", "queue", "upload")
		files, err := ioutil.ReadDir(queueDir)
		assert.Equal(t, nil, err)
		assert.Equal(t, 1, len(files))

		by, err := ioutil.ReadFile(filepath.Join(queueDir, files[0].Name()))
		assert.Equal(t, nil, err)
		assert.Equal(t, oid, string(by))
	})

	repo.Test()
}
