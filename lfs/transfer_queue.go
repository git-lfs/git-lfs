package lfs

import (
	"sync"
	"sync/atomic"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errors"
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/progress"
	"github.com/github/git-lfs/transfer"
	"github.com/rubyist/tracerx"
)

const (
	batchSize = 100
)

type Transferable interface {
	Oid() string
	Size() int64
	Name() string
	Path() string
	Object() *api.ObjectResource
	SetObject(*api.ObjectResource)
	// Legacy API check - TODO remove this and only support batch
	LegacyCheck() (*api.ObjectResource, error)
}

// TransferQueue organises the wider process of uploading and downloading,
// including calling the API, passing the actual transfer request to transfer
// adapters, and dealing with progress, errors and retries
type TransferQueue struct {
	direction         transfer.Direction
	adapter           transfer.TransferAdapter
	adapterInProgress bool
	adapterResultChan chan transfer.TransferResult
	adapterInitMutex  sync.Mutex
	dryRun            bool
	retrying          uint32
	meter             *progress.ProgressMeter
	errors            []error
	transferables     map[string]Transferable
	retries           []Transferable
	batcher           *Batcher
	apic              chan Transferable // Channel for processing individual API requests
	retriesc          chan Transferable // Channel for processing retries
	errorc            chan error        // Channel for processing errors
	watchers          []chan string
	trMutex           *sync.Mutex
	errorwait         sync.WaitGroup
	retrywait         sync.WaitGroup
	wait              sync.WaitGroup // Incremented on Add(), decremented on transfer complete or skip
	oldApiWorkers     int            // Number of non-batch API workers to spawn (deprecated)
	manifest          *transfer.Manifest
}

// newTransferQueue builds a TransferQueue, direction and underlying mechanism determined by adapter
func newTransferQueue(files int, size int64, dryRun bool, dir transfer.Direction) *TransferQueue {
	logPath, _ := config.Config.Os.Get("GIT_LFS_PROGRESS")

	q := &TransferQueue{
		direction:     dir,
		dryRun:        dryRun,
		meter:         progress.NewProgressMeter(files, size, dryRun, logPath),
		apic:          make(chan Transferable, batchSize),
		retriesc:      make(chan Transferable, batchSize),
		errorc:        make(chan error),
		oldApiWorkers: config.Config.ConcurrentTransfers(),
		transferables: make(map[string]Transferable),
		trMutex:       &sync.Mutex{},
		manifest:      transfer.ConfigureManifest(transfer.NewManifest(), config.Config),
	}

	q.errorwait.Add(1)
	q.retrywait.Add(1)

	q.run()

	return q
}

// Add adds a Transferable to the transfer queue.
func (q *TransferQueue) Add(t Transferable) {
	q.wait.Add(1)
	q.trMutex.Lock()
	q.transferables[t.Oid()] = t
	q.trMutex.Unlock()

	if q.batcher != nil {
		q.batcher.Add(t)
		return
	}

	q.apic <- t
}

func (q *TransferQueue) useAdapter(name string) {
	q.adapterInitMutex.Lock()
	defer q.adapterInitMutex.Unlock()

	if q.adapter != nil {
		if q.adapter.Name() == name {
			// re-use, this is the normal path
			return
		}
		// If the adapter we're using isn't the same as the one we've been
		// told to use now, must wait for the current one to finish then switch
		// This will probably never happen but is just in case server starts
		// changing adapter support in between batches
		q.finishAdapter()
	}
	q.adapter = q.manifest.NewAdapterOrDefault(name, q.direction)
}

func (q *TransferQueue) finishAdapter() {
	if q.adapterInProgress {
		q.adapter.End()
		q.adapterInProgress = false
		q.adapter = nil
	}
}

func (q *TransferQueue) addToAdapter(t Transferable) {
	tr := transfer.NewTransfer(t.Name(), t.Object(), t.Path())

	if q.dryRun {
		// Don't actually transfer
		res := transfer.TransferResult{tr, nil}
		q.handleTransferResult(res)
		return
	}
	err := q.ensureAdapterBegun()
	if err != nil {
		q.errorc <- err
		q.Skip(t.Size())
		q.wait.Done()
		return
	}
	q.adapter.Add(tr)
}

func (q *TransferQueue) Skip(size int64) {
	q.meter.Skip(size)
}

func (q *TransferQueue) transferKind() string {
	if q.direction == transfer.Download {
		return "download"
	} else {
		return "upload"
	}
}

func (q *TransferQueue) ensureAdapterBegun() error {
	q.adapterInitMutex.Lock()
	defer q.adapterInitMutex.Unlock()

	if q.adapterInProgress {
		return nil
	}

	adapterResultChan := make(chan transfer.TransferResult, 20)

	// Progress callback - receives byte updates
	cb := func(name string, total, read int64, current int) error {
		q.meter.TransferBytes(q.transferKind(), name, read, total, current)
		return nil
	}

	tracerx.Printf("tq: starting transfer adapter %q", q.adapter.Name())
	err := q.adapter.Begin(config.Config.ConcurrentTransfers(), cb, adapterResultChan)
	if err != nil {
		return err
	}
	q.adapterInProgress = true

	// Collector for completed transfers
	// q.wait.Done() in handleTransferResult is enough to know when this is complete for all transfers
	go func() {
		for res := range adapterResultChan {
			q.handleTransferResult(res)
		}
	}()

	return nil
}

func (q *TransferQueue) handleTransferResult(res transfer.TransferResult) {
	if res.Error != nil {
		if q.canRetry(res.Error) {
			tracerx.Printf("tq: retrying object %s", res.Transfer.Object.Oid)
			q.trMutex.Lock()
			t, ok := q.transferables[res.Transfer.Object.Oid]
			q.trMutex.Unlock()
			if ok {
				q.retry(t)
			} else {
				q.errorc <- res.Error
			}
		} else {
			q.errorc <- res.Error
		}
	} else {
		oid := res.Transfer.Object.Oid
		for _, c := range q.watchers {
			c <- oid
		}

		q.meter.FinishTransfer(res.Transfer.Name)
	}

	q.wait.Done()

}

// Wait waits for the queue to finish processing all transfers. Once Wait is
// called, Add will no longer add transferables to the queue. Any failed
// transfers will be automatically retried once.
func (q *TransferQueue) Wait() {
	if q.batcher != nil {
		q.batcher.Exit()
	}

	q.wait.Wait()

	// Handle any retries
	close(q.retriesc)
	q.retrywait.Wait()
	atomic.StoreUint32(&q.retrying, 1)

	if len(q.retries) > 0 {
		tracerx.Printf("tq: retrying %d failed transfers", len(q.retries))
		for _, t := range q.retries {
			q.Add(t)
		}
		if q.batcher != nil {
			q.batcher.Exit()
		}
		q.wait.Wait()
	}

	atomic.StoreUint32(&q.retrying, 0)

	close(q.apic)
	q.finishAdapter()
	close(q.errorc)

	for _, watcher := range q.watchers {
		close(watcher)
	}

	q.meter.Finish()
	q.errorwait.Wait()
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
// TODO LEGACY API: remove when legacy API removed
func (q *TransferQueue) individualApiRoutine(apiWaiter chan interface{}) {
	for t := range q.apic {
		obj, err := t.LegacyCheck()
		if err != nil {
			if q.canRetry(err) {
				q.retry(t)
			} else {
				q.errorc <- err
			}
			q.wait.Done()
			continue
		}

		if apiWaiter != nil { // Signal to launch more individual api workers
			q.meter.Start()
			select {
			case apiWaiter <- 1:
			default:
			}
		}

		// Legacy API has no support for anything but basic transfer adapter
		q.useAdapter(transfer.BasicAdapterName)
		if obj != nil {
			t.SetObject(obj)
			q.meter.Add(t.Name())
			q.addToAdapter(t)
		} else {
			q.Skip(t.Size())
			q.wait.Done()
		}
	}
}

// legacyFallback is used when a batch request is made to a server that does
// not support the batch endpoint. When this happens, the Transferables are
// fed from the batcher into apic to be processed individually.
// TODO LEGACY API: remove when legacy API removed
func (q *TransferQueue) legacyFallback(failedBatch []interface{}) {
	tracerx.Printf("tq: batch api not implemented, falling back to individual")

	q.launchIndividualApiRoutines()

	for _, t := range failedBatch {
		q.apic <- t.(Transferable)
	}

	for {
		batch := q.batcher.Next()
		if batch == nil {
			break
		}

		for _, t := range batch {
			q.apic <- t.(Transferable)
		}
	}
}

// batchApiRoutine processes the queue of transfers using the batch endpoint,
// making only one POST call for all objects. The results are then handed
// off to the transfer workers.
func (q *TransferQueue) batchApiRoutine() {
	var startProgress sync.Once

	transferAdapterNames := q.manifest.GetAdapterNames(q.direction)

	for {
		batch := q.batcher.Next()
		if batch == nil {
			break
		}

		tracerx.Printf("tq: sending batch of size %d", len(batch))

		transfers := make([]*api.ObjectResource, 0, len(batch))
		for _, i := range batch {
			t := i.(Transferable)
			transfers = append(transfers, &api.ObjectResource{Oid: t.Oid(), Size: t.Size()})
		}

		if len(transfers) == 0 {
			continue
		}

		objs, adapterName, err := api.Batch(config.Config, transfers, q.transferKind(), transferAdapterNames)
		if err != nil {
			if errors.IsNotImplementedError(err) {
				git.Config.SetLocal("", "lfs.batch", "false")

				go q.legacyFallback(batch)
				return
			}

			if q.canRetry(err) {
				for _, t := range batch {
					q.retry(t.(Transferable))
				}
			} else {
				q.errorc <- err
			}

			q.wait.Add(-len(transfers))
			continue
		}

		q.useAdapter(adapterName)
		startProgress.Do(q.meter.Start)

		for _, o := range objs {
			if o.Error != nil {
				q.errorc <- errors.Wrapf(o.Error, "[%v] %v", o.Oid, o.Error.Message)
				q.Skip(o.Size)
				q.wait.Done()
				continue
			}

			if _, ok := o.Rel(q.transferKind()); ok {
				// This object needs to be transferred
				q.trMutex.Lock()
				transfer, ok := q.transferables[o.Oid]
				q.trMutex.Unlock()

				if ok {
					transfer.SetObject(o)
					q.meter.Add(transfer.Name())
					q.addToAdapter(transfer)
				} else {
					q.Skip(transfer.Size())
					q.wait.Done()
				}
			} else {
				q.Skip(o.Size)
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
	q.errorwait.Done()
}

func (q *TransferQueue) retryCollector() {
	for t := range q.retriesc {
		q.retries = append(q.retries, t)
	}
	q.retrywait.Done()
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

		for i := 0; i < q.oldApiWorkers-1; i++ {
			go q.individualApiRoutine(nil)
		}
	}()
}

// run starts the transfer queue, doing individual or batch transfers depending
// on the Config.BatchTransfer() value. run will transfer files sequentially or
// concurrently depending on the Config.ConcurrentTransfers() value.
func (q *TransferQueue) run() {
	go q.errorCollector()
	go q.retryCollector()

	if config.Config.BatchTransfer() {
		tracerx.Printf("tq: running as batched queue, batch size of %d", batchSize)
		q.batcher = NewBatcher(batchSize)
		go q.batchApiRoutine()
	} else {
		tracerx.Printf("tq: running as individual queue")
		q.launchIndividualApiRoutines()
	}
}

func (q *TransferQueue) retry(t Transferable) {
	q.retriesc <- t
}

func (q *TransferQueue) canRetry(err error) bool {
	if !errors.IsRetriableError(err) || atomic.LoadUint32(&q.retrying) == 1 {
		return false
	}

	return true
}

// Errors returns any errors encountered during transfer.
func (q *TransferQueue) Errors() []error {
	return q.errors
}
