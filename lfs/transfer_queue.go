package lfs

import (
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/vendor/_nuts/github.com/cheggaaa/pb"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

type Transferable interface {
	Check() (*objectResource, *WrappedError)
	Transfer(CopyCallback) *WrappedError
	Object() *objectResource
	Oid() string
	Size() int64
	SetObject(*objectResource)
}

// TransferQueue provides a queue that will allow concurrent transfers.
type TransferQueue struct {
	transferc        chan Transferable
	errorc           chan *WrappedError
	watchers         []chan string
	errors           []*WrappedError
	wg               sync.WaitGroup
	workers          int
	files            int
	finished         int64
	size             int64
	authCond         *sync.Cond
	transferables    map[string]Transferable
	bar              *pb.ProgressBar
	clientAuthorized int32
	transferKind     string
}

// newTransferQueue builds a TransferQueue, allowing `workers` concurrent transfers.
func newTransferQueue(workers, files int) *TransferQueue {
	return &TransferQueue{
		transferc:     make(chan Transferable, files),
		errorc:        make(chan *WrappedError),
		watchers:      make([]chan string, 0),
		workers:       workers,
		files:         files,
		authCond:      sync.NewCond(&sync.Mutex{}),
		transferables: make(map[string]Transferable),
	}
}

// Add adds a Transferable to the transfer queue.
func (q *TransferQueue) Add(t Transferable) {
	q.transferables[t.Oid()] = t
}

// Watch returns a channel where the queue will write the OID of each transfer
// as it completes. The channel will be closed when the queue finishes processing.
func (q *TransferQueue) Watch() chan string {
	c := make(chan string, q.files)
	q.watchers = append(q.watchers, c)
	return c
}

// processIndividual processes the queue of transfers one at a time by making
// a POST call for each object, feeding the results to the transfer workers.
// If configured, the object transfers can still happen concurrently, the
// sequential nature here is only for the meta POST calls.
func (q *TransferQueue) processIndividual() {
	apic := make(chan Transferable, q.files)
	workersReady := make(chan int, q.workers)
	var wg sync.WaitGroup

	for i := 0; i < q.workers; i++ {
		go func() {
			workersReady <- 1
			for t := range apic {
				// If an API authorization has not occured, we wait until we're woken up.
				q.authCond.L.Lock()
				if atomic.LoadInt32(&q.clientAuthorized) == 0 {
					q.authCond.Wait()
				}
				q.authCond.L.Unlock()

				obj, err := t.Check()
				if err != nil {
					q.errorc <- err
					wg.Done()
					continue
				}
				if obj != nil {
					q.wg.Add(1)
					t.SetObject(obj)
					q.transferc <- t
				}
				wg.Done()
			}
		}()
	}

	q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", q.finished, len(q.transferables)))
	q.bar.Start()

	for _, t := range q.transferables {
		wg.Add(1)
		apic <- t
	}

	<-workersReady
	q.authCond.Signal() // Signal the first goroutine to run
	close(apic)
	wg.Wait()

	close(q.transferc)
}

// processBatch processes the queue of transfers using the batch endpoint,
// making only one POST call for all objects. The results are then handed
// off to the transfer workers.
func (q *TransferQueue) processBatch() error {
	transfers := make([]*objectResource, 0, len(q.transferables))
	for _, t := range q.transferables {
		transfers = append(transfers, &objectResource{Oid: t.Oid(), Size: t.Size()})
	}

	objects, err := Batch(transfers)
	if err != nil {
		if isNotImplError(err) {
			tracerx.Printf("queue: batch not implemented, disabling")
			configFile := filepath.Join(LocalGitDir, "config")
			git.Config.SetLocal(configFile, "lfs.batch", "false")
		}

		return err
	}

	q.files = 0

	for _, o := range objects {
		if _, ok := o.Links[q.transferKind]; ok {
			// This object needs to be transfered
			if transfer, ok := q.transferables[o.Oid]; ok {
				q.files++
				q.wg.Add(1)
				transfer.SetObject(o)
				q.transferc <- transfer
			}
		}
	}

	close(q.transferc)
	q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", q.finished, q.files))
	q.bar.Start()
	sendApiEvent(apiEventSuccess) // Wake up transfer workers
	return nil
}

// Process starts the transfer queue and displays a progress bar. Process will
// do individual or batch transfers depending on the Config.BatchTransfer() value.
// Process will transfer files sequentially or concurrently depending on the
// Concig.ConcurrentTransfers() value.
func (q *TransferQueue) Process() {
	q.bar = pb.New64(q.size)
	q.bar.SetUnits(pb.U_BYTES)
	q.bar.ShowBar = false

	// This goroutine collects errors returned from transfers
	go func() {
		for err := range q.errorc {
			q.errors = append(q.errors, err)
		}
	}()

	// This goroutine watches for apiEvents. In order to prevent multiple
	// credential requests from happening, the queue is processed sequentially
	// until an API request succeeds (meaning authenication has happened successfully).
	// Once the an API request succeeds, all worker goroutines are woken up and allowed
	// to process transfers. Once a success happens, this goroutine exits.
	go func() {
		for {
			event := <-apiEvent
			switch event {
			case apiEventSuccess:
				atomic.StoreInt32(&q.clientAuthorized, 1)
				q.authCond.Broadcast() // Wake all remaining goroutines
				return
			case apiEventFail:
				q.authCond.Signal() // Wake the next goroutine
			}
		}
	}()

	for i := 0; i < q.workers; i++ {
		// These are the worker goroutines that process transfers
		go func(n int) {

			for transfer := range q.transferc {
				cb := func(total, read int64, current int) error {
					q.bar.Add(current)
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

				f := atomic.AddInt64(&q.finished, 1)
				q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", f, q.files))
				q.wg.Done()
			}
		}(i)
	}

	if Config.BatchTransfer() {
		if err := q.processBatch(); err != nil {
			q.processIndividual()
		}
	} else {
		q.processIndividual()
	}

	q.wg.Wait()
	close(q.errorc)
	for _, watcher := range q.watchers {
		close(watcher)
	}

	q.bar.Finish()
}

// Errors returns any errors encountered during transfer.
func (q *TransferQueue) Errors() []*WrappedError {
	return q.errors
}
