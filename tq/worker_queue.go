package tq

import "sync"

// WorkerFn is a function type, held by a `*workerQueue`, which is called
// anytime there is work to be done. The parameter of the function "oid" is
// given as the OID of the object to transfer. The return value dictates whether
// or not the object was transferred successfully, true if so, false if
// otherwise.
type WorkerFn func(oid string) bool

type workerQueue struct {
	// tasks holds `*task`s to be worked upon by `WorkerFn`s. It is written
	// to from a single source (see: `Add(batch []string)` below), and
	// consumed from many workers.
	tasks chan *task
	// fn is the WorkerFn that executes when an OID is available to be
	// processed. If the function returns true, then the OID is marked as
	// successfully processed, and no retries are enqueued. Otherwise, the
	// items is marked as having failed, and a retry is enqueued by writing
	// the OID to the returned `retries` channel.
	fn WorkerFn

	// wg waits for all items given in an `Add()` call to be processed
	// _completely_ by a worker. After latching, the given retries channel
	// will be closed.
	wg *sync.WaitGroup
	// wwg waits for all workers to terminate after the `tasks` channel
	// closes, and is a terminating condition of the Wait() function below.
	wwg *sync.WaitGroup
}

// task encapsulates the necessary information to perform work on an object.
type task struct {
	// Oid is the object ID to be worked upon.
	Oid string

	// retry is a channel written to when the object is marked as having failed.
	retry chan<- string
	// wg is a *sync.WaitGroup that is incremented when new tasks are
	// created, and decremented when tasks are completed (whether or not
	// they are successful is a different matter).
	wg *sync.WaitGroup
}

// MarkForRetry places the task's object ID on the retry channel.
func (t *task) MarkForRetry() {
	t.retry <- t.Oid
}

// Done signals that work on this object has completed in either a successful or
// failed state.
func (t *task) Done() {
	t.wg.Done()
}

// newWorkerQueue creates a new `*workerQueue` with `count` workers, each
// executing the `WorkerFn` "fn".
func newWorkerQueue(count int, fn WorkerFn) *workerQueue {
	q := &workerQueue{
		tasks: make(chan *task),
		fn:    fn,

		wg:  new(sync.WaitGroup),
		wwg: new(sync.WaitGroup),
	}

	for i := 0; i < count; i++ {
		q.wwg.Add(1)

		go func() {
			defer q.wwg.Done()

			for task := range q.tasks {
				if ok := q.fn(task.Oid); !ok {
					task.MarkForRetry()
				}

				task.Done()
			}
		}()
	}

	return q
}

// Add adds a batch of items to a queue of items to be processed, and then
// distributes those items across the available workers. It immediately returns
// a channel that object IDs are written to when those IDs need to be retried.
// That channel is closed when work has completed on the entire batch.
func (q *workerQueue) Add(batch []string) <-chan string {
	q.wg.Add(len(batch))
	retries := make(chan string, len(batch))

	go func() {
		defer close(retries)

		for _, oid := range batch {
			q.tasks <- &task{oid, retries, q.wg}
		}

		q.wg.Wait()
	}()

	return retries
}

// Wait waits for all workers to finish processing active OIDs, then closes the
// incoming items channel and therefore all living workers. After waiting for
// the transitioning workers to completely transition from an alive to dead
// state, this function returns.
func (q *workerQueue) Wait() {
	q.wg.Wait()
	close(q.tasks)
	q.wwg.Wait()
}
