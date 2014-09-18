package gitmedia

import (
	"bytes"
	"github.com/bmizerany/assert"
	"io"
	"testing"
)

func TestGitHash(t *testing.T) {
	reader := bytes.NewBufferString("hello, world\n")

	gitHash, _ := NewGitHash(reader)
	io.Copy(gitHash, reader)
	gitHash.Close()

	hash := gitHash.Hash()
	assert.Equal(t, "4b5fa63702dd96796042e92787f464e28f09f17d", hash)
}
