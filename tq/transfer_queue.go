package tq

import (
	"sort"
	"sync"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/rubyist/tracerx"
)

const (
	defaultBatchSize  = 100
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

func newRetryCounter() *retryCounter {
	return &retryCounter{
		MaxRetries: defaultMaxRetries,
		count:      make(map[string]int),
	}
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

// Batch implements the sort.Interface interface and enables sorting on a slice
// of `Transferable`s by object size.
//
// This interface is implemented here so that the largest objects can be
// processed first. Since adding a new batch is unable to occur until the
// current batch has finished processing, this enables us to reduce the risk of
// a single worker getting tied up on a large item at the end of a batch while
// all other workers are sitting idle.
type Batch []Transferable

func (b Batch) ApiObjects() []*api.ObjectResource {
	transfers := make([]*api.ObjectResource, 0, len(b))
	for _, t := range b {
		transfers = append(transfers, &api.ObjectResource{
			Oid:  t.Oid(),
			Size: t.Size(),
		})
	}

	return transfers
}

func (b Batch) Len() int           { return len(b) }
func (b Batch) Less(i, j int) bool { return b[i].Size() < b[j].Size() }
func (b Batch) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// TransferQueue organises the wider process of uploading and downloading,
// including calling the API, passing the actual transfer request to transfer
// adapters, and dealing with progress, errors and retries.
type TransferQueue struct {
	direction         Direction
	adapter           Adapter
	adapterInProgress bool
	adapterInitMutex  sync.Mutex
	dryRun            bool
	meter             progress.Meter
	errors            []error
	transferables     map[string]Transferable
	batchSize         int
	bufferDepth       int
	// Channel for processing (and buffering) incoming items
	incoming      chan Transferable
	errorc        chan error // Channel for processing errors
	watchers      []chan string
	trMutex       *sync.Mutex
	startProgress sync.Once
	collectorWait sync.WaitGroup
	errorwait     sync.WaitGroup
	// wait is used to keep track of pending transfers. It is incremented
	// once per unique OID on Add(), and is decremented when that transfer
	// is marked as completed or failed, but not retried.
	wait     sync.WaitGroup
	manifest *Manifest
	rc       *retryCounter
}

type Option func(*TransferQueue)

func DryRun(dryRun bool) Option {
	return func(tq *TransferQueue) {
		tq.dryRun = dryRun
	}
}

func WithProgress(m progress.Meter) Option {
	return func(tq *TransferQueue) {
		tq.meter = m
	}
}

func WithBatchSize(size int) Option {
	return func(tq *TransferQueue) { tq.batchSize = size }
}

func WithBufferDepth(depth int) Option {
	return func(tq *TransferQueue) { tq.bufferDepth = depth }
}

func WithGitEnv(gitEnv Environment) Option {
	return func(tq *TransferQueue) {
		ConfigureManifest(tq.manifest, gitEnv)

		if mr := gitEnv.Int("lfs.transfer.maxretries", 0); mr > 0 {
			tq.rc.MaxRetries = mr
		}
	}
}

// NewTransferQueue builds a TransferQueue, direction and underlying mechanism determined by adapter
func NewTransferQueue(dir Direction, options ...Option) *TransferQueue {
	q := &TransferQueue{
		direction:     dir,
		errorc:        make(chan error),
		transferables: make(map[string]Transferable),
		trMutex:       &sync.Mutex{},
		manifest:      NewManifest(),
		rc: &retryCounter{
			MaxRetries: defaultMaxRetries,
			count:      make(map[string]int),
		},
	}

	for _, opt := range options {
		opt(q)
	}

	if q.batchSize <= 0 {
		q.batchSize = defaultBatchSize
	}
	if q.bufferDepth <= 0 {
		q.bufferDepth = q.batchSize
	}

	q.incoming = make(chan Transferable, q.bufferDepth)

	if q.meter == nil {
		q.meter = progress.Noop()
	}

	q.collectorWait.Add(1)
	q.errorwait.Add(1)
	q.run()

	return q
}

// Add adds a Transferable to the transfer queue. It only increments the amount
// of waiting the TransferQueue has to do if the Transferable "t" is new.
func (q *TransferQueue) Add(t Transferable) {
	if isNew := q.remember(t); !isNew {
		tracerx.Printf("already transferring %q, skipping duplicate", t.Oid())
		return
	}

	q.incoming <- t
}

// remember remembers the Transferable "t" if the *TransferQueue doesn't already
// know about a Transferable with the same OID.
//
// It returns if the value is new or not.
func (q *TransferQueue) remember(t Transferable) bool {
	q.trMutex.Lock()
	defer q.trMutex.Unlock()

	if _, ok := q.transferables[t.Oid()]; !ok {
		q.wait.Add(1)
		q.transferables[t.Oid()] = t

		return true
	}
	return false
}

// collectBatches collects batches in a loop, prioritizing failed items from the
// previous before adding new items. The process works as follows:
//
//   1. Create a new batch, of size `q.batchSize`, and containing no items
//   2. While the batch contains less items than `q.batchSize` AND the channel
//      is open, read one item from the `q.incoming` channel.
//      a. If the read was a channel close, go to step 4.
//      b. If the read was a Transferable item, go to step 3.
//   3. Append the item to the batch.
//   4. Sort the batch by descending object size, make a batch API call, send
//      the items to the `*adapterBase`.
//   5. Process the worker results, incrementing and appending retries if
//      possible.
//   6. If the `q.incoming` channel is open, go to step 2.
//   7. If the next batch is empty AND the `q.incoming` channel is closed,
//      terminate immediately.
//
// collectBatches runs in its own goroutine.
func (q *TransferQueue) collectBatches() {
	defer q.collectorWait.Done()

	var closing bool
	batch := q.makeBatch()

	for {
		for !closing && (len(batch) < q.batchSize) {
			t, ok := <-q.incoming
			if !ok {
				closing = true
				break
			}

			batch = append(batch, t)
		}

		// Before enqueuing the next batch, sort by descending object
		// size.
		sort.Sort(sort.Reverse(batch))

		retries, err := q.enqueueAndCollectRetriesFor(batch)
		if err != nil {
			q.errorc <- err
		}

		if closing && len(retries) == 0 {
			break
		}

		batch = retries
	}
}

// enqueueAndCollectRetriesFor makes a Batch API call and returns a "next" batch
// containing all of the objects that failed from the previous batch and had
// retries availale to them.
//
// If an error was encountered while making the API request, _all_ of the items
// from the previous batch (that have retries available to them) will be
// returned immediately, along with the error that was encountered.
//
// enqueueAndCollectRetriesFor blocks until the entire Batch "batch" has been
// processed.
func (q *TransferQueue) enqueueAndCollectRetriesFor(batch Batch) (Batch, error) {
	next := q.makeBatch()
	transferAdapterNames := q.manifest.GetAdapterNames(q.direction)

	tracerx.Printf("tq: sending batch of size %d", len(batch))

	objs, adapterName, err := api.Batch(
		config.Config, batch.ApiObjects(), q.transferKind(), transferAdapterNames,
	)
	if err != nil {
		// If there was an error making the batch API call, mark all of
		// the objects for retry, and return them along with the error
		// that was encountered. If any of the objects couldn't be
		// retried, they will be marked as failed.
		for _, t := range batch {
			if q.canRetryObject(t.Oid(), err) {
				q.rc.Increment(t.Oid())

				next = append(next, t)
			} else {
				q.wait.Done()
			}
		}

		return next, err
	}

	q.useAdapter(adapterName)
	q.startProgress.Do(q.meter.Start)

	toTransfer := make([]*Transfer, 0, len(objs))

	for _, o := range objs {
		if o.Error != nil {
			q.errorc <- errors.Wrapf(o.Error, "[%v] %v", o.Oid, o.Error.Message)
			q.Skip(o.Size)
			q.wait.Done()

			continue
		}

		if _, needsTransfer := o.Rel(q.transferKind()); needsTransfer {
			// If the object has a link relation for the kind of
			// transfer that we want to perform, grab a Transferable
			// that matches the object's OID.
			q.trMutex.Lock()
			t, ok := q.transferables[o.Oid]
			q.trMutex.Unlock()

			if ok {
				// If we knew about an associated Transferable,
				// begin the transfer.
				t.SetObject(o)
				q.meter.StartTransfer(t.Name())

				toTransfer = append(toTransfer, NewTransfer(
					t.Name(), t.Object(), t.Path(),
				))
			} else {
				// If we couldn't find any associated
				// Transferable object, then we give up on the
				// transfer by telling the progress meter to
				// skip the number of bytes in "o".
				q.errorc <- errors.Errorf("[%v] The server returned an unknown OID.", o.Oid)

				q.Skip(o.Size)
				q.wait.Done()
			}
		} else {
			// Otherwise, if the object didn't need a transfer at
			// all, skip and decrement it.
			q.Skip(o.Size)
			q.wait.Done()
		}
	}

	retries := q.addToAdapter(toTransfer)
	for t := range retries {
		q.rc.Increment(t.Oid())
		count := q.rc.CountFor(t.Oid())

		tracerx.Printf("tq: enqueue retry #%d for %q (size: %d)", count, t.Oid(), t.Size())

		next = append(next, t)
	}

	return next, nil
}

// makeBatch returns a new, empty batch, with a capacity equal to the maximum
// batch size designated by the `*TransferQueue`.
func (q *TransferQueue) makeBatch() Batch { return make(Batch, 0, q.batchSize) }

// addToAdapter adds the given "pending" transfers to the transfer adapters and
// returns a channel of Transferables that are to be retried in the next batch.
// After all of the items in the batch have been processed, the channel is
// closed.
//
// addToAdapter returns immediately, and does not block.
func (q *TransferQueue) addToAdapter(pending []*Transfer) <-chan Transferable {
	retries := make(chan Transferable, len(pending))

	if err := q.ensureAdapterBegun(); err != nil {
		close(retries)

		q.errorc <- err
		for _, t := range pending {
			q.Skip(t.Object.Size)
			q.wait.Done()
		}

		return retries
	}

	go func() {
		defer close(retries)

		var results <-chan TransferResult
		if q.dryRun {
			results = q.makeDryRunResults(pending)
		} else {
			results = q.adapter.Add(pending...)
		}

		for res := range results {
			q.handleTransferResult(res, retries)
		}
	}()

	return retries
}

// makeDryRunResults returns a channel populated immediately with "successful"
// results for all of the given transfers in "ts".
func (q *TransferQueue) makeDryRunResults(ts []*Transfer) <-chan TransferResult {
	results := make(chan TransferResult, len(ts))
	for _, t := range ts {
		results <- TransferResult{t, nil}
	}

	close(results)

	return results
}

// handleTransferResult observes the transfer result, sending it on the retries
// channel if it was able to be retried.
func (q *TransferQueue) handleTransferResult(
	res TransferResult, retries chan<- Transferable,
) {
	oid := res.Transfer.Object.Oid

	if res.Error != nil {
		// If there was an error encountered when processing the
		// transfer (res.Transfer), handle the error as is appropriate:

		if q.canRetryObject(oid, res.Error) {
			// If the object can be retried, send it on the retries
			// channel, where it will be read at the call-site and
			// its retry count will be incremented.
			tracerx.Printf("tq: retrying object %s", oid)

			q.trMutex.Lock()
			t, ok := q.transferables[oid]
			q.trMutex.Unlock()

			if ok {
				retries <- t
			} else {
				q.errorc <- res.Error
			}
		} else {
			// If the error wasn't retriable, OR the object has
			// exceeded its retry budget, it will be NOT be sent to
			// the retry channel, and the error will be reported
			// immediately.
			q.errorc <- res.Error
			q.wait.Done()
		}
	} else {
		// Otherwise, if the transfer was successful, notify all of the
		// watchers, and mark it as finished.
		for _, c := range q.watchers {
			c <- oid
		}

		q.meter.FinishTransfer(res.Transfer.Name)
		q.wait.Done()
	}
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

func (q *TransferQueue) Skip(size int64) {
	q.meter.Skip(size)
}

func (q *TransferQueue) transferKind() string {
	if q.direction == Download {
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

	// Progress callback - receives byte updates
	cb := func(name string, total, read int64, current int) error {
		q.meter.TransferBytes(q.transferKind(), name, read, total, current)
		return nil
	}

	tracerx.Printf("tq: starting transfer adapter %q", q.adapter.Name())
	err := q.adapter.Begin(config.Config.ConcurrentTransfers(), cb)
	if err != nil {
		return err
	}
	q.adapterInProgress = true

	return nil
}

// Wait waits for the queue to finish processing all transfers. Once Wait is
// called, Add will no longer add transferables to the queue. Any failed
// transfers will be automatically retried once.
func (q *TransferQueue) Wait() {
	close(q.incoming)

	q.wait.Wait()
	q.collectorWait.Wait()

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
	c := make(chan string, q.batchSize)
	q.watchers = append(q.watchers, c)
	return c
}

// This goroutine collects errors returned from transfers
func (q *TransferQueue) errorCollector() {
	for err := range q.errorc {
		q.errors = append(q.errors, err)
	}
	q.errorwait.Done()
}

// run starts the transfer queue, doing individual or batch transfers depending
// on the Config.BatchTransfer() value. run will transfer files sequentially or
// concurrently depending on the Config.ConcurrentTransfers() value.
func (q *TransferQueue) run() {
	tracerx.Printf("tq: running as batched queue, batch size of %d", q.batchSize)

	go q.errorCollector()
	go q.collectBatches()
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
