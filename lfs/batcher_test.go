package lfs

import (
	"testing"

	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

func TestBatcherSizeMet(t *testing.T) {
	assertAll([]batcherTestCase{
		{2, 4, false},
		{3, 5, false},
		{0, 0, false},
	}, t)
}

func TestBatcherExit(t *testing.T) {
	assertAll([]batcherTestCase{
		{2, 4, true},
		{3, 5, true},
		{0, 0, true},
	}, t)
}

// Type batcherTestCase specifies information about how to run a particular test
// around the type lfs.Batcher.
type batcherTestCase struct {
	BatchSize  int
	ItemCount  int
	ShouldExit bool
}

// Func Batcher makes and retunrs a lfs.Batcher according to the specification
// given in this instance of batcherTestCase. When returned, it is filled with
// the given amount of items, and has exited if it was told to.
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

// Func Batches returns the number of individual batches expected to be
// processed by the batcher under test.
func (b batcherTestCase) Batches() int {
	if b.BatchSize == 0 {
		return 0
	}

	return b.ItemCount / b.BatchSize
}

// Func assertAll processes all test cases, throwing assertion errors if they
// fail.
func assertAll(cases []batcherTestCase, t *testing.T) {
	for _, c := range cases {
		b := c.Batcher()

		items := c.ItemCount
		for i := 0; i < c.Batches(); i++ {
			group := b.Next()
			assert.Equal(t, c.BatchSize, len(group))
			items -= c.BatchSize
		}
	}
}
