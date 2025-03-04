package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClampDiscardsIntsLowerThanMin(t *testing.T) {
	assert.Equal(t, 0, ClampInt(-1, 0, 1))
}

func TestClampDiscardsIntsGreaterThanMax(t *testing.T) {
	assert.Equal(t, 1, ClampInt(2, 0, 1))
}

func TestClampAcceptsIntsWithinBounds(t *testing.T) {
	assert.Equal(t, 1, ClampInt(1, 0, 2))
}
