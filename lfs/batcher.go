package lfs

import "sync/atomic"

// Batcher provides a way to process a set of items in groups of n. Items can
// be added to the batcher from multiple goroutines and pulled off in groups
// when one of the following conditions occurs:
//   * The batch size is reached
//   * Exit() is called
// When an Exit() occurs, the group may be smaller than the batch size.
type Batcher struct {
	exited     uint32
	batchSize  int
	input      chan interface{}
	batchReady chan []interface{}
	truncate   chan interface{}
}

// NewBatcher creates a Batcher with the batchSize.
func NewBatcher(batchSize int) *Batcher {
	b := &Batcher{
		batchSize:  batchSize,
		input:      make(chan interface{}),
		batchReady: make(chan []interface{}),
		truncate:   make(chan interface{}),
	}

	go b.acceptInput()
	return b
}

// Add adds an item to the batcher. Add is safe to call from multiple
// goroutines.
func (b *Batcher) Add(t interface{}) {
	if atomic.CompareAndSwapUint32(&b.exited, 1, 0) {
		b.input = make(chan interface{})
		b.truncate = make(chan interface{})
		go b.acceptInput()
	}

	b.input <- t
}

// Next will wait for the one of the above batch triggers to occur and return
// the accumulated batch.
func (b *Batcher) Next() []interface{} {
	return <-b.batchReady
}

// Truncate causes the current batch to halt accumulation and return
// immediately, even if it is smaller than the given batch size.
func (b *Batcher) Truncate() {
	b.truncate <- struct{}{}
}

// Exit stops all batching and allows Next() to return. Calling Add() after
// calling Exit() will reset the batcher.
func (b *Batcher) Exit() {
	atomic.StoreUint32(&b.exited, 1)
	close(b.input)
	close(b.truncate)
}

// acceptInput runs in its own goroutine and accepts input from external
// clients. Without truncation, it fills and dispenses batches in a sequential
// order: for a batch size N, N items will be processed before a new batch is
// ready. If a batch is truncated while still filling itself, it will be
// returned immediately, opening up a new batch for all subsequent items.
func (b *Batcher) acceptInput() {
	exit := false

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
			case <-b.truncate:
				break Acc
			}
		}

		b.batchReady <- batch

		if exit {
			return
		}
	}
}
