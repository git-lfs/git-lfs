package tq

import "sync"

// Queue organizes and distributes work on OIDs against a set of available
// workers (represented by a *workerQueue). Any failed items returned back by
// that `*workerQueue` will be prioritized into the next batch.
type Queue struct {
	// incoming serves as the barrier between new calls to `Add()` and
	// accumulating the next batch of items to be sent into the worker
	// queue.
	//
	// incoming has a finite non-zero buffer equal to the size of a batch
	// such that one extra batch of data can be buffered before applying
	// back-pressure to callers.
	incoming chan string
	// workers maintains a handle on the *workerQueue instance used to
	// distribute batched sets of OIDs
	workers *workerQueue

	// wg holds onto the run() goroutine and latches only after the
	// goroutine has terminated, enabling safe shutdowns.
	wg *sync.WaitGroup
}

// New instantiates a new `*tq.Queue` type with a given batch size "size",
// number of workers "workers", and a function to do that work, "fn".
//
// Once returned, the `*Queue` will be active and able to accept new writes.
func New(size, workers int, fn WorkerFn) *Queue {
	q := &Queue{
		incoming: make(chan string, size),
		workers:  newWorkerQueue(workers, fn),

		wg: new(sync.WaitGroup),
	}

	q.wg.Add(1)
	go q.run()

	return q
}

// run accumulates new items into a batch for sending to the pool of available
// workers, prioritizing retried objects over new objects.
//
// The function executes as follows:
//
//   1. Allocate a new batch of items for sending to the worker pool.
//   2. Accept as many writes as we can before overfilling the batch from the
//      `q.incoming` channel. If the channel closes, mark that we are closing
//      and continue.
//   3. Send the batch to the worker pool
//   4. Empty the batch, begin re-filling it with items from the retry channel
//      until the channel is closed.
//   5. If closing, and the batch is empty, return.
//   6. Otherwise, go to step 2.
//
// run runs in its own goroutine.
func (q *Queue) run() {
	defer q.wg.Done()

	var closing bool

	batch := make([]string, 0, cap(q.incoming))

	for {
		for !closing && (len(batch) < cap(q.incoming)) {
			oid, ok := <-q.incoming
			if !ok {
				closing = true
				break
			}

			batch = append(batch, oid)
		}

		retries := q.workers.Add(batch)
		batch = make([]string, 0, cap(q.incoming))

		for retry := range retries {
			batch = append(batch, retry)
		}

		if closing && len(batch) == 0 {
			return
		}
	}
}

// Add enqueues a new item for entry into the next available batch of items. If
// the batch is already full and one more full batch's worth of items has been
// Add()-ed already, this function will block.
//
// Add cannot be called after the `*Queue` has been marked as `Wait()`-ing.
func (q *Queue) Add(oid string) {
	// TODO(taylor): potentially store whether or not we're closing and
	// check that before writing to a closed channel? Not sure what to do
	// with that "oid" though... drop it? Return an error? Panic? All seem
	// uniquely bad...

	q.incoming <- oid
}

// Wait marks the queue for shutting down. After calling this function, the
// queue will process all queued new items after processing all of the retried
// items.
//
// Once the queue is marked as closing and a single empty batch is detected, the
// queue will be shut down completely, and this function will return.
func (q *Queue) Wait() {
	close(q.incoming)

	q.wg.Wait()
}
