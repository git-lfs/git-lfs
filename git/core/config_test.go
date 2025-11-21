package core_test // to avoid import cycles

import (
	"testing"

	"github.com/git-lfs/git-lfs/v3/git/core"
	"github.com/stretchr/testify/assert"
)

func TestReadOnlyConfig(t *testing.T) {
	cfg := core.NewReadOnlyConfig("", "")
	_, err := cfg.SetLocal("lfs.this.should", "fail")
	assert.Equal(t, err, core.ErrReadOnly)
}
