package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func MinIntPicksTheSmallerInt(t *testing.T) {
	assert.Equal(t, -1, MinInt(-1, 1))
}

func MaxIntPicksTheBiggertInt(t *testing.T) {
	assert.Equal(t, 1, MaxInt(-1, 1))
}

func ClampDiscardsIntsLowerThanMin(t *testing.T) {
	assert.Equal(t, 0, ClampInt(-1, 0, 1))
}

func ClampDiscardsIntsGreaterThanMax(t *testing.T) {
	assert.Equal(t, 1, ClampInt(2, 0, 1))
}

func ClampAcceptsIntsWithinBounds(t *testing.T) {
	assert.Equal(t, 1, ClampInt(1, 0, 2))
}
