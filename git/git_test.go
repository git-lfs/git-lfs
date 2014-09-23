package git

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestHashObject(t *testing.T) {
	hash, _ := NewHashObject([]byte("hello, world\n"))
	assert.Equal(t, "4b5fa63702dd96796042e92787f464e28f09f17d", hash)
}
