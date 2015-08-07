package lfs

import (
	"path/filepath"
	"sync"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

const (
	batchSize = 100
)

type Transferable interface {
	Check() (*objectResource, *WrappedError)
	Transfer(CopyCallback) *WrappedError
	Object() *objectResource
	Oid() string
	Size() int64
	Name() string
	SetObject(*objectResource)
}

// TransferQueue provides a queue that will allow concurrent transfers.
type TransferQueue struct {
	meter         *ProgressMeter
	workers       int // Number of transfer workers to spawn
	transferKind  string
	errors        []*WrappedError
	transferables map[string]Transferable
	batcher       *Batcher
	apic          chan Transferable  // Channel for processing individual API requests
	transferc     chan Transferable  // Channel for processing transfers
	errorc        chan *WrappedError // Channel for processing errors
	watchers      []chan string
	wait          sync.WaitGroup
}

// newTransferQueue builds a TransferQueue, allowing `workers` concurrent transfers.
func newTransferQueue(files int, size int64, dryRun bool) *TransferQueue {
	q := &TransferQueue{
		meter:         NewProgressMeter(files, size, dryRun),
		apic:          make(chan Transferable, batchSize),
		transferc:     make(chan Transferable, batchSize),
		errorc:        make(chan *WrappedError),
		workers:       Config.ConcurrentTransfers(),
		transferables: make(map[string]Transferable),
	}

	q.run()

	return q
}

// Add adds a Transferable to the transfer queue.
func (q *TransferQueue) Add(t Transferable) {
	q.wait.Add(1)
	q.transferables[t.Oid()] = t

	if q.batcher != nil {
		q.batcher.Add(t)
		return
	}

	q.apic <- t
}

// Wait waits for the queue to finish processing all transfers
func (q *TransferQueue) Wait() {
	if q.batcher != nil {
		q.batcher.Exit()
	}

	q.wait.Wait()
	close(q.apic)
	close(q.transferc)
	close(q.errorc)

	for _, watcher := range q.watchers {
		close(watcher)
	}

	q.meter.Finish()
}

// Watch returns a channel where the queue will write the OID of each transfer
// as it completes. The channel will be closed when the queue finishes processing.
func (q *TransferQueue) Watch() chan string {
	c := make(chan string, batchSize)
	q.watchers = append(q.watchers, c)
	return c
}

// individualApiRoutine processes the queue of transfers one at a time by making
// a POST call for each object, feeding the results to the transfer workers.
// If configured, the object transfers can still happen concurrently, the
// sequential nature here is only for the meta POST calls.
func (q *TransferQueue) individualApiRoutine(apiWaiter chan interface{}) {
	for t := range q.apic {
		obj, err := t.Check()
		if err != nil {
			q.wait.Done()
			q.errorc <- err
			continue
		}

		if apiWaiter != nil { // Signal to launch more individual api workers
			q.meter.Start()
			select {
			case apiWaiter <- 1:
			default:
			}
		}

		if obj != nil {
			t.SetObject(obj)
			q.meter.Add(t.Name())
			q.transferc <- t
		}
	}
}

// legacyFallback is used when a batch request is made to a server that does
// not support the batch endpoint. When this happens, the Transferables are
// fed from the batcher into apic to be processed individually.
func (q *TransferQueue) legacyFallback(failedBatch []Transferable) {
	tracerx.Printf("tq: batch api not implemented, falling back to individual")

	q.launchIndividualApiRoutines()

	for _, t := range failedBatch {
		q.apic <- t
	}

	for {
		batch := q.batcher.Next()
		if batch == nil {
			break
		}

		for _, t := range batch {
			q.apic <- t
		}
	}
}

// batchApiRoutine processes the queue of transfers using the batch endpoint,
// making only one POST call for all objects. The results are then handed
// off to the transfer workers.
func (q *TransferQueue) batchApiRoutine() {
	var startProgress sync.Once

	for {
		batch := q.batcher.Next()
		if batch == nil {
			break
		}

		tracerx.Printf("tq: sending batch of size %d", len(batch))

		transfers := make([]*objectResource, 0, len(batch))
		for _, t := range batch {
			transfers = append(transfers, &objectResource{Oid: t.Oid(), Size: t.Size()})
		}

		objects, err := Batch(transfers, q.transferKind)
		if err != nil {
			if isNotImplError(err) {
				configFile := filepath.Join(LocalGitDir, "config")
				git.Config.SetLocal(configFile, "lfs.batch", "false")

				go q.legacyFallback(batch)
				return
			}
			q.errorc <- err
			continue
		}

		startProgress.Do(q.meter.Start)

		for _, o := range objects {
			if _, ok := o.Rel(q.transferKind); ok {
				// This object has an error
				if o.Error != nil {
					q.errorc <- Error(o.Error)
					q.meter.Skip(o.Size)
					q.wait.Done()
					continue
				}

				// This object needs to be transferred
				if transfer, ok := q.transferables[o.Oid]; ok {
					transfer.SetObject(o)
					q.meter.Add(transfer.Name())
					q.transferc <- transfer
				} else {
					q.meter.Skip(transfer.Size())
					q.wait.Done()
				}
			} else {
				q.meter.Skip(o.Size)
				q.wait.Done()
			}
		}
	}
}

// This goroutine collects errors returned from transfers
func (q *TransferQueue) errorCollector() {
	for err := range q.errorc {
		q.errors = append(q.errors, err)
	}
}

func (q *TransferQueue) transferWorker() {
	for transfer := range q.transferc {
		cb := func(total, read int64, current int) error {
			q.meter.TransferBytes(q.transferKind, transfer.Name(), read, total, current)
			return nil
		}

		if err := transfer.Transfer(cb); err != nil {
			q.errorc <- err
		} else {
			oid := transfer.Oid()
			for _, c := range q.watchers {
				c <- oid
			}
		}

		q.meter.FinishTransfer(transfer.Name())

		q.wait.Done()
	}
}

// launchIndividualApiRoutines first launches a single api worker. When it
// receives the first successful api request it launches workers - 1 more
// workers. This prevents being prompted for credentials multiple times at once
// when they're needed.
func (q *TransferQueue) launchIndividualApiRoutines() {
	go func() {
		apiWaiter := make(chan interface{})
		go q.individualApiRoutine(apiWaiter)

		<-apiWaiter

		for i := 0; i < q.workers-1; i++ {
			go q.individualApiRoutine(nil)
		}
	}()
}

// run starts the transfer queue, doing individual or batch transfers depending
// on the Config.BatchTransfer() value. run will transfer files sequentially or
// concurrently depending on the Config.ConcurrentTransfers() value.
func (q *TransferQueue) run() {
	go q.errorCollector()

	tracerx.Printf("tq: starting %d transfer workers", q.workers)
	for i := 0; i < q.workers; i++ {
		go q.transferWorker()
	}

	if Config.BatchTransfer() {
		tracerx.Printf("tq: running as batched queue, batch size of %d", batchSize)
		q.batcher = NewBatcher(batchSize)
		go q.batchApiRoutine()
	} else {
		tracerx.Printf("tq: running as individual queue")
		q.launchIndividualApiRoutines()
	}
}

// Errors returns any errors encountered during transfer.
func (q *TransferQueue) Errors() []*WrappedError {
	return q.errors
}
