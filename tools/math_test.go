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
