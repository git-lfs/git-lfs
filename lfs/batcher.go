package lfs

import "sync/atomic"

// Batcher provides a way to process a set of items in groups of n. Items can
// be added to the batcher from multiple goroutines and pulled off in groups
// when one of the following conditions occurs:
//   * The batch size is reached
//   * Flush() is called, forcing the batch to be returned immediately, as-is
//   * Exit() is called
// When an Exit() or Flush() occurs, the group may be smaller than the batch
// size.
type Batcher struct {
	exited     uint32
	batchSize  int
	input      chan interface{}
	batchReady chan []interface{}
	flush      chan interface{}
}

// NewBatcher creates a Batcher with the batchSize.
func NewBatcher(batchSize int) *Batcher {
	b := &Batcher{
		batchSize:  batchSize,
		input:      make(chan interface{}),
		batchReady: make(chan []interface{}),
		flush:      make(chan interface{}),
	}

	go b.acceptInput()
	return b
}

// Add adds one or more items to the batcher. Add is safe to call from multiple
// goroutines.
func (b *Batcher) Add(ts ...interface{}) {
	if atomic.CompareAndSwapUint32(&b.exited, 1, 0) {
		b.input = make(chan interface{})
		b.flush = make(chan interface{})
		go b.acceptInput()
	}

	for _, t := range ts {
		b.input <- t
	}
}

// Next will wait for the one of the above batch triggers to occur and return
// the accumulated batch.
func (b *Batcher) Next() []interface{} {
	return <-b.batchReady
}

// Flush causes the current batch to halt accumulation and return
// immediately, even if it is smaller than the given batch size.
func (b *Batcher) Flush() {
	b.flush <- struct{}{}
}

// Exit stops all batching and allows Next() to return. Calling Add() after
// calling Exit() will reset the batcher.
func (b *Batcher) Exit() {
	atomic.StoreUint32(&b.exited, 1)
	close(b.input)
	close(b.flush)
}

// acceptInput runs in its own goroutine and accepts input from external
// clients. Without flushing, the batch is filled completely in a sequential
// order, and then dispensed. If, while filling a batch, it is flushed part-way
// through, the batch will be dispensed with its current contents, and all
// subsequent Add()s will be placed in the next batch.
func (b *Batcher) acceptInput() {
	var exit bool

	for {
		batch := make([]interface{}, 0, b.batchSize)
	Acc:
		for len(batch) < b.batchSize {
			select {
			case t, ok := <-b.input:
				if !ok {
					exit = true // input channel was closed by Exit()
					break Acc
				}

				batch = append(batch, t)
			case <-b.flush:
				break Acc
			}
		}

		b.batchReady <- batch

		if exit {
			return
		}
	}
}
