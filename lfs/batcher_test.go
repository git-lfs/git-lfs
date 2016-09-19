package lfs

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatcherOnlyReturnsCompleteBatchesWithoutExiting(t *testing.T) {
	runBatcherTests([]batcherTestCase{
		{2, 4, false},
		{3, 5, false},
		{0, 0, false},
	}, t)
}

func TestBatcherReturnsIncompleteBatchesWhenExiting(t *testing.T) {
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

// Assert asserts that appropriate sized batches were emitted given a batch size
// and item count, (and potentially an instruction to exit).
//
// In the case that the item count is an even divisor of the batch size, we
// expect several full batches where the total amount of items returned in the
// batches is equal to the given item count.
//
// In the case where the item count is _not_ an even divisor of the batch size,
// we expect either: one partially full batch, or at least one completely full
// batch and one partially full batch. If the batcher is instructed to exit, we
// expect to see an item count returned equal to the given item count. If the
// batcher is not instructed to exit, we should see an item count equal to the
// number of full batches received * the batch size.
func (b batcherTestCase) Assert(t *testing.T) {
	batcher := NewBatcher(b.BatchSize)

	var remaining = b.ItemCount
	for remaining > 0 {
		size := int(math.Min(float64(b.BatchSize), float64(remaining)))
		for i := 0; i < size; i++ {
			batcher.Add(struct{}{})
		}

		if remaining < b.BatchSize {
			if b.ShouldExit {
				// If there is an uneven amount of items and we
				// choose to exit, we should receive a partial
				// batch upon calling `batcher.Next()` because
				// of the following call to `batcher.Exit()`.
				batcher.Exit()
			} else {
				// If there was an uneven amount of items and we
				// choose _not_ to exit, then those remaining
				// items are abandoned.
				return
			}
		}

		assert.Len(t, batcher.Next(), size)

		remaining -= size
	}
}

// runBatcherTests processes all test cases, throwing assertion errors if they
// fail.
func runBatcherTests(cases []batcherTestCase, t *testing.T) {
	for _, c := range cases {
		c.Assert(t)
	}
}
