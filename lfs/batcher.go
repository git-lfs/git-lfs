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
}

// NewBatcher creates a Batcher with the batchSize.
func NewBatcher(batchSize int) *Batcher {
	b := &Batcher{
		batchSize:  batchSize,
		input:      make(chan interface{}, batchSize),
		batchReady: make(chan []interface{}),
	}

	go b.acceptInput()
	return b
}

// Add adds an item to the batcher. Add is safe to call from multiple
// goroutines.
func (b *Batcher) Add(t interface{}) {
	if atomic.CompareAndSwapUint32(&b.exited, 1, 0) {
		b.input = make(chan interface{}, b.batchSize)
		go b.acceptInput()
	}

	b.input <- t
}

// Next will wait for the one of the above batch triggers to occur and return
// the accumulated batch.
func (b *Batcher) Next() []interface{} {
	return <-b.batchReady
}

// Exit stops all batching and allows Next() to return. Calling Add() after
// calling Exit() will reset the batcher.
func (b *Batcher) Exit() {
	atomic.StoreUint32(&b.exited, 1)
	close(b.input)
}

// acceptInput runs in its own goroutine and accepts input from external
// clients. It fills and dispenses batches in a sequential order: for a batch
// size N, N items will be processed before a new batch is ready.
func (b *Batcher) acceptInput() {
	exit := false

	for {
		batch := make([]interface{}, 0, b.batchSize)
	Loop:
		for len(batch) < b.batchSize {
			t, ok := <-b.input
			if !ok {
				exit = true // input channel was closed by Exit()
				break Loop
			}
			batch = append(batch, t)
		}

		b.batchReady <- batch

		if exit {
			return
		}
	}
}
