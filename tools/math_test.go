package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinIntPicksTheSmallerInt(t *testing.T) {
	assert.Equal(t, -1, MinInt(-1, 1))
}

func TestMaxIntPicksTheBiggertInt(t *testing.T) {
	assert.Equal(t, 1, MaxInt(-1, 1))
}

func TestClampDiscardsIntsLowerThanMin(t *testing.T) {
	assert.Equal(t, 0, ClampInt(-1, 0, 1))
}

func TestClampDiscardsIntsGreaterThanMax(t *testing.T) {
	assert.Equal(t, 1, ClampInt(2, 0, 1))
}

func TestClampAcceptsIntsWithinBounds(t *testing.T) {
	assert.Equal(t, 1, ClampInt(1, 0, 2))
}
