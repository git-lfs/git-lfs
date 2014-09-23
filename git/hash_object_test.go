package git

import (
	"bytes"
	"github.com/bmizerany/assert"
	"io"
	"testing"
)

func TestHashObject(t *testing.T) {
	reader := bytes.NewBufferString("hello, world\n")

	gitHash, _ := NewHashObject()
	io.Copy(gitHash, reader)
	gitHash.Close()

	hash := gitHash.Hash()
	assert.Equal(t, "4b5fa63702dd96796042e92787f464e28f09f17d", hash)
}
