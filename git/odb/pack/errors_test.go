package pack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsupportedVersionErr(t *testing.T) {
	u := &UnsupportedVersionErr{Got: 3}

	assert.Error(t, u, "git/odb/pack: unsupported version: 3")
}
