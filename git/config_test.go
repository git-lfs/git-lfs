package git_test // to avoid import cycles

import (
	"testing"

	. "github.com/git-lfs/git-lfs/v3/git"
	"github.com/stretchr/testify/assert"
)

func TestReadOnlyConfig(t *testing.T) {
	cfg := NewReadOnlyConfig("", "")
	_, err := cfg.SetLocal("lfs.this.should", "fail")
	assert.Equal(t, err, ErrReadOnly)
}
