package lfs

import (
	"sync"
	"sync/atomic"
)

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
	wg         sync.WaitGroup
	input      chan interface{}
	batchReady chan []interface{}
	flush      chan interface{}
}

// NewBatcher creates a Batcher with the batchSize.
func NewBatcher(batchSize int) *Batcher {
	b := &Batcher{
		batchSize:  batchSize,
		input:      make(chan interface{}, batchSize),
		batchReady: make(chan []interface{}),
		flush:      make(chan interface{}),
	}

	go b.acceptInput()
	return b
}

// Add adds an item (or many items) to the batcher. Add is safe to call from
// multiple goroutines.
func (b *Batcher) Add(ts ...interface{}) {
	if atomic.CompareAndSwapUint32(&b.exited, 1, 0) {
		b.input = make(chan interface{}, b.batchSize)
		b.flush = make(chan interface{})
		go b.acceptInput()
	}

	b.wg.Add(len(ts))
	for _, t := range ts {
		b.input <- t
	}
	b.wg.Wait()
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
// clients. Without truncation, it fills and dispenses batches in a sequential
// order: for a batch size N, N items will be processed before a new batch is
// ready. If a batch is truncated while still filling itself, it will be
// returned immediately, opening up a new batch for all subsequent items.
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
				b.wg.Done()
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
