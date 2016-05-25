package point

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssertEqual(t *testing.T) {
	p1 := Point{1, 1}
	p2 := Point{1, 1}

	assert.Equal(t, p1, p2)
}

func TestAssertNotEqual(t *testing.T) {
	p1 := Point{1, 1}
	p2 := Point{2, 1}

	assert.NotEqual(t, p1, p2)
}
