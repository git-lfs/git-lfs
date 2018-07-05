package pack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoundsLeft(t *testing.T) {
	assert.EqualValues(t, 1, newBounds(1, 2).Left())
}

func TestBoundsRight(t *testing.T) {
	assert.EqualValues(t, 2, newBounds(1, 2).Right())
}

func TestBoundsWithLeftReturnsNewBounds(t *testing.T) {
	b1 := newBounds(1, 2)
	b2 := b1.WithLeft(3)

	assert.EqualValues(t, 1, b1.Left())
	assert.EqualValues(t, 2, b1.Right())

	assert.EqualValues(t, 3, b2.Left())
	assert.EqualValues(t, 2, b2.Right())
}

func TestBoundsWithRightReturnsNewBounds(t *testing.T) {
	b1 := newBounds(1, 2)
	b2 := b1.WithRight(3)

	assert.EqualValues(t, 1, b1.Left())
	assert.EqualValues(t, 2, b1.Right())

	assert.EqualValues(t, 1, b2.Left())
	assert.EqualValues(t, 3, b2.Right())
}

func TestBoundsEqualWithIdenticalBounds(t *testing.T) {
	b1 := newBounds(1, 2)
	b2 := newBounds(1, 2)

	assert.True(t, b1.Equal(b2))
}

func TestBoundsEqualWithDifferentBounds(t *testing.T) {
	b1 := newBounds(1, 2)
	b2 := newBounds(3, 4)

	assert.False(t, b1.Equal(b2))
}

func TestBoundsEqualWithNilReceiver(t *testing.T) {
	bnil := (*bounds)(nil)
	b2 := newBounds(1, 2)

	assert.False(t, bnil.Equal(b2))
}

func TestBoundsEqualWithNilArgument(t *testing.T) {
	b1 := newBounds(1, 2)
	bnil := (*bounds)(nil)

	assert.False(t, b1.Equal(bnil))
}

func TestBoundsEqualWithNilArgumentAndReceiver(t *testing.T) {
	b1 := (*bounds)(nil)
	b2 := (*bounds)(nil)

	assert.True(t, b1.Equal(b2))
}

func TestBoundsString(t *testing.T) {
	b1 := newBounds(1, 2)

	assert.Equal(t, "[1,2]", b1.String())
}
