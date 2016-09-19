package lfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatcherSizeMet(t *testing.T) {
	runBatcherTests([]batcherTestCase{
		{2, 4, false},
		{3, 5, false},
		{0, 0, false},
	}, t)
}

func TestBatcherExit(t *testing.T) {
	runBatcherTests([]batcherTestCase{
		{2, 4, true},
		{3, 5, true},
		{0, 0, true},
	}, t)
}

func TestBatcherTruncatesPartialBatches(t *testing.T) {
	first, second := "first", "second"

	b := NewBatcher(3)
	b.Add(first)
	b.Add(second)
	b.Truncate()

	batch := b.Next()

	require.Len(t, batch, 2)
	assert.Equal(t, first, batch[0])
	assert.Equal(t, second, batch[1])
}

// batcherTestCase specifies information about how to run a particular test
// around the type lfs.Batcher.
type batcherTestCase struct {
	BatchSize  int
	ItemCount  int
	ShouldExit bool
}

// Batcher makes and returns a lfs.Batcher according to the specification given
// in this instance of batcherTestCase. When returned, it is filled with the
// given amount of items, and has exited if it was told to.
func (b batcherTestCase) Batcher() *Batcher {
	batcher := NewBatcher(b.BatchSize)
	for i := 0; i < b.ItemCount; i++ {
		batcher.Add(&Downloadable{})
	}

	if b.ShouldExit {
		batcher.Exit()
	}

	return batcher
}

// Batches returns the number of individual batches expected to be processed by
// the batcher under test.
func (b batcherTestCase) Batches() int {
	if b.BatchSize == 0 {
		return b.ItemCount
	}

	return b.ItemCount / b.BatchSize
}

// runBatcherTests processes all test cases, throwing assertion errors if they
// fail.
func runBatcherTests(cases []batcherTestCase, t *testing.T) {
	for _, c := range cases {
		b := c.Batcher()

		items := c.ItemCount
		for i := 0; i < c.Batches(); i++ {
			group := b.Next()
			assert.Len(t, group, c.BatchSize)
			items -= c.BatchSize
		}
	}
}
