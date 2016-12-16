package lfs

import (
	"sync"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/git-lfs/git-lfs/transfer"
	"github.com/rubyist/tracerx"
)

const (
	batchSize         = 100
	defaultMaxRetries = 1
)

type Transferable interface {
	Oid() string
	Size() int64
	Name() string
	Path() string
	Object() *api.ObjectResource
	SetObject(*api.ObjectResource)
}

type retryCounter struct {
	// MaxRetries is the maximum number of retries a single object can
	// attempt to make before it will be dropped.
	MaxRetries int `git:"lfs.transfer.maxretries"`

	// cmu guards count
	cmu sync.Mutex
	// count maps OIDs to number of retry attempts
	count map[string]int
}

// newRetryCounter instantiates a new *retryCounter. It parses the gitconfig
// value: `lfs.transfer.maxretries`, and falls back to defaultMaxRetries if none
// was provided.
//
// If it encountered an error in Unmarshaling the *config.Configuration, it will
// be returned, otherwise nil.
func newRetryCounter(cfg *config.Configuration) *retryCounter {
	rc := &retryCounter{
		MaxRetries: defaultMaxRetries,

		count: make(map[string]int),
	}

	if err := cfg.Unmarshal(rc); err != nil {
		tracerx.Printf("rc: error parsing config, falling back to default values...: %v", err)
		rc.MaxRetries = 1
	}

	if rc.MaxRetries < 1 {
		tracerx.Printf("rc: invalid retry count: %d, defaulting to %d", rc.MaxRetries, 1)
		rc.MaxRetries = 1
	}

	return rc
}

// Increment increments the number of retries for a given OID. It is safe to
// call across multiple goroutines.
func (r *retryCounter) Increment(oid string) {
	r.cmu.Lock()
	defer r.cmu.Unlock()

	r.count[oid]++
}

// CountFor returns the current number of retries for a given OID. It is safe to
// call across multiple goroutines.
func (r *retryCounter) CountFor(oid string) int {
	r.cmu.Lock()
	defer r.cmu.Unlock()

	return r.count[oid]
}

// CanRetry returns the current number of retries, and whether or not it exceeds
// the maximum number of retries (see: retryCounter.MaxRetries).
func (r *retryCounter) CanRetry(oid string) (int, bool) {
	count := r.CountFor(oid)
	return count, count < r.MaxRetries
}

// TransferQueue organises the wider process of uploading and downloading,
// including calling the API, passing the actual transfer request to transfer
// adapters, and dealing with progress, errors and retries.
type TransferQueue struct {
	direction         transfer.Direction
	adapter           transfer.TransferAdapter
	adapterInProgress bool
	adapterResultChan chan transfer.TransferResult
	adapterInitMutex  sync.Mutex
	dryRun            bool
	meter             progress.Meter
	errors            []error
	transferables     map[string]Transferable
	batcher           *Batcher
	retriesc          chan Transferable // Channel for processing retries
	errorc            chan error        // Channel for processing errors
	watchers          []chan string
	trMutex           *sync.Mutex
	errorwait         sync.WaitGroup
	retrywait         sync.WaitGroup
	// wait is used to keep track of pending transfers. It is incremented
	// once per unique OID on Add(), and is decremented when that transfer
	// is marked as completed or failed, but not retried.
	wait     sync.WaitGroup
	manifest *transfer.Manifest
	rc       *retryCounter
}

type transferQueueOption func(*TransferQueue)

func DryRun(dryRun bool) transferQueueOption {
	return func(tq *TransferQueue) {
		tq.dryRun = dryRun
	}
}

func WithProgress(m progress.Meter) transferQueueOption {
	return func(tq *TransferQueue) {
		tq.meter = m
	}
}

// newTransferQueue builds a TransferQueue, direction and underlying mechanism determined by adapter
func newTransferQueue(dir transfer.Direction, options ...transferQueueOption) *TransferQueue {
	q := &TransferQueue{
		direction:     dir,
		retriesc:      make(chan Transferable, batchSize),
		errorc:        make(chan error),
		transferables: make(map[string]Transferable),
		trMutex:       &sync.Mutex{},
		manifest:      transfer.ConfigureManifest(transfer.NewManifest(), config.Config),
		rc:            newRetryCounter(config.Config),
	}

	for _, opt := range options {
		opt(q)
	}

	if q.meter == nil {
		q.meter = progress.Noop()
	}

	q.errorwait.Add(1)
	q.retrywait.Add(1)
	q.run()

	return q
}

// Add adds a Transferable to the transfer queue. It only increments the amount
// of waiting the TransferQueue has to do if the Transferable "t" is new.
func (q *TransferQueue) Add(t Transferable) {
	q.trMutex.Lock()
	if _, ok := q.transferables[t.Oid()]; !ok {
		q.wait.Add(1)
		q.transferables[t.Oid()] = t
		q.trMutex.Unlock()
	} else {
		tracerx.Printf("already transferring %q, skipping duplicate", t)
		q.trMutex.Unlock()
		return
	}

	q.batcher.Add(t)
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

// handleTransferResult is responsible for dealing with the result of a
// successful or failed transfer.
//
// If there was an error assosicated with the given transfer, "res.Error", and
// it is retriable (see: `q.canRetryObject`), it will be placed in the next
// batch and be retried. If that error is not retriable for any reason, the
// transfer will be marked as having failed, and the error will be reported.
//
// If the transfer was successful, the watchers of this transfer queue will be
// notified, and the transfer will be marked as having been completed.
func (q *TransferQueue) handleTransferResult(res transfer.TransferResult) {
	oid := res.Transfer.Object.Oid

	if res.Error != nil {
		if q.canRetryObject(oid, res.Error) {
			tracerx.Printf("tq: retrying object %s", oid)
			q.trMutex.Lock()
			t, ok := q.transferables[oid]
			q.trMutex.Unlock()
			if ok {
				q.retry(t)
			} else {
				q.errorc <- res.Error
			}
		} else {
			q.errorc <- res.Error
			q.wait.Done()
		}
	} else {
		for _, c := range q.watchers {
			c <- oid
		}

		q.meter.FinishTransfer(res.Transfer.Name)
		q.wait.Done()
	}
}

// Wait waits for the queue to finish processing all transfers. Once Wait is
// called, Add will no longer add transferables to the queue. Any failed
// transfers will be automatically retried once.
func (q *TransferQueue) Wait() {
	q.batcher.Exit()
	q.wait.Wait()

	// Handle any retries
	close(q.retriesc)
	q.retrywait.Wait()

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
			var errOnce sync.Once
			for _, o := range batch {
				t := o.(Transferable)

				if q.canRetryObject(t.Oid(), err) {
					q.retry(t)
				} else {
					errOnce.Do(func() { q.errorc <- err })
					q.wait.Done()
				}
			}

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
					q.meter.StartTransfer(transfer.Name())
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

// retryCollector collects objects to retry, increments the number of times that
// they have been retried, and then enqueues them in the next batch.  If the
// transfer queue is using a batcher, the batch will be flushed immediately.
//
// retryCollector runs in its own goroutine.
func (q *TransferQueue) retryCollector() {
	for t := range q.retriesc {
		q.rc.Increment(t.Oid())
		count := q.rc.CountFor(t.Oid())

		tracerx.Printf("tq: enqueue retry #%d for %q (size: %d)", count, t.Oid(), t.Size())

		// XXX(taylor): reuse some of the logic in
		// `*TransferQueue.Add(t)` here to circumvent banned duplicate
		// OIDs
		tracerx.Printf("tq: flushing batch in response to retry #%d for %q (size: %d)", count, t.Oid(), t.Size())

		q.batcher.Add(t)
		q.batcher.Flush()
	}
	q.retrywait.Done()
}

// run starts the transfer queue, doing individual or batch transfers depending
// on the Config.BatchTransfer() value. run will transfer files sequentially or
// concurrently depending on the Config.ConcurrentTransfers() value.
func (q *TransferQueue) run() {
	go q.errorCollector()
	go q.retryCollector()

	tracerx.Printf("tq: running as batched queue, batch size of %d", batchSize)
	q.batcher = NewBatcher(batchSize)
	go q.batchApiRoutine()
}

func (q *TransferQueue) retry(t Transferable) {
	q.retriesc <- t
}

// canRetry returns whether or not the given error "err" is retriable.
func (q *TransferQueue) canRetry(err error) bool {
	return errors.IsRetriableError(err)
}

// canRetryObject returns whether the given error is retriable for the object
// given by "oid". If the an OID has met its retry limit, then it will not be
// able to be retried again. If so, canRetryObject returns whether or not that
// given error "err" is retriable.
func (q *TransferQueue) canRetryObject(oid string, err error) bool {
	if count, ok := q.rc.CanRetry(oid); !ok {
		tracerx.Printf("tq: refusing to retry %q, too many retries (%d)", oid, count)
		return false
	}

	return q.canRetry(err)
}

// Errors returns any errors encountered during transfer.
func (q *TransferQueue) Errors() []error {
	return q.errors
}
