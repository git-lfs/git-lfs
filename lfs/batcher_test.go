package lfs

import (
	"testing"
	"time"

	"github.com/bmizerany/assert"
)

func TestBatcherSizeMet(t *testing.T) {
	b := NewBatcher(2)

	for i := 0; i < 4; i++ {
		b.Add(&Downloadable{})
	}

	group := b.Next()
	assert.Equal(t, 2, len(group))

	group = b.Next()
	assert.Equal(t, 2, len(group))
}

func TestBatcherTimeLimit(t *testing.T) {
	b := NewBatcher(4)

	for i := 0; i < 2; i++ {
		b.Add(&Downloadable{})
	}
	time.Sleep(time.Millisecond * 251)

	group := b.Next()
	assert.Equal(t, 2, len(group))
}

func TestBatcherExit(t *testing.T) {
	b := NewBatcher(4)

	for i := 0; i < 2; i++ {
		b.Add(&Downloadable{})
	}

	b.Exit()

	group := b.Next()
	assert.Equal(t, 2, len(group))
}
