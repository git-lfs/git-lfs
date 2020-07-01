package tq

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/creds"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfshttp"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

const (
	defaultBatchSize = 100
	baseRetryDelayMs = 250
)

type retryCounter struct {
	MaxRetries    int
	MaxRetryDelay int

	// cmu guards count
	cmu sync.Mutex
	// count maps OIDs to number of retry attempts
	count map[string]int
}

// newRetryCounter instantiates a new *retryCounter.
func newRetryCounter() *retryCounter {
	return &retryCounter{
		MaxRetries:    defaultMaxRetries,
		MaxRetryDelay: defaultMaxRetryDelay,
		count:         make(map[string]int),
	}
}

// Increment increments the number of retries for a given OID and returns the
// new value. It is safe to call across multiple goroutines.
func (r *retryCounter) Increment(oid string) int {
	r.cmu.Lock()
	defer r.cmu.Unlock()

	r.count[oid]++
	return r.count[oid]
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

// ReadyTime returns the time from now when the current retry can occur or the
// zero time if the retry can occur immediately.
func (r *retryCounter) ReadyTime(oid string) time.Time {
	count := r.CountFor(oid)
	if count < 1 {
		return time.Time{}
	}

	maxDelayMs := 1000 * uint64(r.MaxRetryDelay)
	delay := uint64(baseRetryDelayMs) * (1 << uint(count-1))
	if delay == 0 || delay > maxDelayMs {
		delay = maxDelayMs
	}
	return time.Now().Add(time.Duration(delay) * time.Millisecond)
}

// batch implements the sort.Interface interface and enables sorting on a slice
// of `*Transfer`s by object size.
//
// This interface is implemented here so that the largest objects can be
// processed first. Since adding a new batch is unable to occur until the
// current batch has finished processing, this enables us to reduce the risk of
// a single worker getting tied up on a large item at the end of a batch while
// all other workers are sitting idle.
type batch []*objectTuple

// Concat concatenates two batches together, returning a single, clamped batch as
// "left", and the remainder of elements as "right". If the union of the
// receiver and "other" has cardinality less than "size", "right" will be
// returned as nil. Any object tuple that is not currently able to be retried
// (ie Retry-After response), will also go into the right batch. Also, when object(s)
// are returend that are rate-limited, return the minimum duration required to wait until
// a object is ready.
func (b batch) Concat(other batch, size int) (left, right batch, minWait time.Duration) {
	u := batch(append(b, other...))
	for _, ot := range u {
		if time.Now().After(ot.ReadyTime) {
			// The current time is past the time the object should
			// be available.
			left = append(left, ot)
		} else {
			// The time hasn't passed for the object.
			right = append(right, ot)
			wait := time.Until(ot.ReadyTime)
			if minWait == 0 {
				minWait = wait
			} else if wait < minWait {
				minWait = wait
			}
		}
	}
	if len(left) <= size {
		// If the size of left fits the given size limit, return with no adjustments.
		return left, right, minWait
	}
	// If left is too large, trip left up to size and append the rest to right.
	right = append(right, left[size:]...)
	left = left[:size]
	return left, right, minWait
}

func (b batch) ToTransfers() []*Transfer {
	transfers := make([]*Transfer, 0, len(b))
	for _, t := range b {
		transfers = append(transfers, &Transfer{Oid: t.Oid, Size: t.Size, Missing: t.Missing})
	}
	return transfers
}

func (b batch) Len() int           { return len(b) }
func (b batch) Less(i, j int) bool { return b[i].Size < b[j].Size }
func (b batch) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

type abortableWaitGroup struct {
	wq      sync.WaitGroup
	counter int
	mu      sync.Mutex
}

func newAbortableWaitGroup() *abortableWaitGroup {
	return &abortableWaitGroup{}
}

func (q *abortableWaitGroup) Add(delta int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.counter += delta
	q.wq.Add(delta)
}

func (q *abortableWaitGroup) Done() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.counter -= 1
	q.wq.Done()
}

func (q *abortableWaitGroup) Abort() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.wq.Add(-q.counter)
}

func (q *abortableWaitGroup) Wait() {
	q.wq.Wait()
}

// TransferQueue organises the wider process of uploading and downloading,
// including calling the API, passing the actual transfer request to transfer
// adapters, and dealing with progress, errors and retries.
type TransferQueue struct {
	direction         Direction
	client            *tqClient
	remote            string
	ref               *git.Ref
	adapter           Adapter
	adapterInProgress bool
	adapterInitMutex  sync.Mutex
	dryRun            bool
	cb                tools.CopyCallback
	meter             *Meter
	errors            []error
	transfers         map[string]*objects
	batchSize         int
	bufferDepth       int
	incoming          chan *objectTuple // Channel for processing incoming items
	errorc            chan error        // Channel for processing errors
	watchers          []chan *Transfer
	trMutex           *sync.Mutex
	collectorWait     sync.WaitGroup
	errorwait         sync.WaitGroup
	// wait is used to keep track of pending transfers. It is incremented
	// once per unique OID on Add(), and is decremented when that transfer
	// is marked as completed or failed, but not retried.
	wait     *abortableWaitGroup
	manifest *Manifest
	rc       *retryCounter

	// unsupportedContentType indicates whether the transfer queue ever saw
	// an HTTP 422 response indicating that their upload destination does
	// not support Content-Type detection.
	unsupportedContentType bool
}

// objects holds a set of objects.
type objects struct {
	completed bool
	objects   []*objectTuple
}

// All returns all *objectTuple's contained in the *objects set.
func (s *objects) All() []*objectTuple {
	return s.objects
}

// Append returns a new *objects with the given *objectTuple(s) appended to the
// end of the known objects.
func (s *objects) Append(os ...*objectTuple) *objects {
	return &objects{
		completed: s.completed,
		objects:   append(s.objects, os...),
	}
}

// First returns the first *objectTuple in the chain of objects.
func (s *objects) First() *objectTuple {
	if len(s.objects) == 0 {
		return nil
	}
	return s.objects[0]
}

type objectTuple struct {
	Name, Path, Oid string
	Size            int64
	Missing         bool
	ReadyTime       time.Time
}

func (o *objectTuple) ToTransfer() *Transfer {
	return &Transfer{
		Name:    o.Name,
		Path:    o.Path,
		Oid:     o.Oid,
		Size:    o.Size,
		Missing: o.Missing,
	}
}

type Option func(*TransferQueue)

func DryRun(dryRun bool) Option {
	return func(tq *TransferQueue) {
		tq.dryRun = dryRun
	}
}

func WithProgress(m *Meter) Option {
	return func(tq *TransferQueue) {
		tq.meter = m
	}
}

func RemoteRef(ref *git.Ref) Option {
	return func(tq *TransferQueue) {
		tq.ref = ref
	}
}

func WithProgressCallback(cb tools.CopyCallback) Option {
	return func(tq *TransferQueue) {
		tq.cb = cb
	}
}

func WithBatchSize(size int) Option {
	return func(tq *TransferQueue) { tq.batchSize = size }
}

func WithBufferDepth(depth int) Option {
	return func(tq *TransferQueue) { tq.bufferDepth = depth }
}

// NewTransferQueue builds a TransferQueue, direction and underlying mechanism determined by adapter
func NewTransferQueue(dir Direction, manifest *Manifest, remote string, options ...Option) *TransferQueue {
	q := &TransferQueue{
		direction: dir,
		client:    &tqClient{Client: manifest.APIClient()},
		remote:    remote,
		errorc:    make(chan error),
		transfers: make(map[string]*objects),
		trMutex:   &sync.Mutex{},
		manifest:  manifest,
		rc:        newRetryCounter(),
		wait:      newAbortableWaitGroup(),
	}

	for _, opt := range options {
		opt(q)
	}

	q.rc.MaxRetries = q.manifest.maxRetries
	q.rc.MaxRetryDelay = q.manifest.maxRetryDelay
	q.client.MaxRetries = q.manifest.maxRetries

	if q.batchSize <= 0 {
		q.batchSize = defaultBatchSize
	}
	if q.bufferDepth <= 0 {
		q.bufferDepth = q.batchSize
	}
	if q.meter != nil {
		q.meter.Direction = q.direction
	}

	q.incoming = make(chan *objectTuple, q.bufferDepth)
	q.collectorWait.Add(1)
	q.errorwait.Add(1)
	q.run()

	return q
}

// Add adds a *Transfer to the transfer queue. It only increments the amount
// of waiting the TransferQueue has to do if the *Transfer "t" is new.
//
// If another transfer(s) with the same OID has been added to the *TransferQueue
// already, the given transfer will not be enqueued, but will be sent to any
// channel created by Watch() once the oldest transfer has completed.
//
// Only one file will be transferred to/from the Path element of the first
// transfer.
func (q *TransferQueue) Add(name, path, oid string, size int64, missing bool, err error) {
	if err != nil {
		q.errorc <- err
		return
	}

	t := &objectTuple{
		Name:    name,
		Path:    path,
		Oid:     oid,
		Size:    size,
		Missing: missing,
	}

	if objs := q.remember(t); len(objs.objects) > 1 {
		if objs.completed {
			// If there is already a completed transfer chain for
			// this OID, then this object is already "done", and can
			// be sent through as completed to the watchers.
			for _, w := range q.watchers {
				w <- t.ToTransfer()
			}
		}

		// If the chain is not done, there is no reason to enqueue this
		// transfer into 'q.incoming'.
		tracerx.Printf("already transferring %q, skipping duplicate", t.Oid)
		return
	}

	q.incoming <- t
}

// remember remembers the *Transfer "t" if the *TransferQueue doesn't already
// know about a Transfer with the same OID.
//
// It returns if the value is new or not.
func (q *TransferQueue) remember(t *objectTuple) objects {
	q.trMutex.Lock()
	defer q.trMutex.Unlock()

	if _, ok := q.transfers[t.Oid]; !ok {
		q.wait.Add(1)
		q.transfers[t.Oid] = &objects{
			objects: []*objectTuple{t},
		}

		return *q.transfers[t.Oid]
	}

	q.transfers[t.Oid] = q.transfers[t.Oid].Append(t)

	return *q.transfers[t.Oid]
}

// collectBatches collects batches in a loop, prioritizing failed items from the
// previous before adding new items. The process works as follows:
//
//   1. Create a new batch, of size `q.batchSize`, and containing no items
//   2. While the batch contains less items than `q.batchSize` AND the channel
//      is open, read one item from the `q.incoming` channel.
//      a. If the read was a channel close, go to step 4.
//      b. If the read was a transferable item, go to step 3.
//   3. Append the item to the batch.
//   4. Sort the batch by descending object size, make a batch API call, send
//      the items to the `*adapterBase`.
//   5. In a separate goroutine, process the worker results, incrementing and
//      appending retries if possible. On the main goroutine, accept new items
//      into "pending".
//   6. Concat() the "next" and "pending" batches such that no more items than
//      the maximum allowed per batch are in next, and the rest are in pending.
//   7. If the `q.incoming` channel is open, go to step 2.
//   8. If the next batch is empty AND the `q.incoming` channel is closed,
//      terminate immediately.
//
// collectBatches runs in its own goroutine.
func (q *TransferQueue) collectBatches() {
	defer q.collectorWait.Done()

	var closing bool
	next := q.makeBatch()
	pending := q.makeBatch()

	for {
		for !closing && (len(next) < q.batchSize) {
			t, ok := <-q.incoming
			if !ok {
				closing = true
				break
			}

			next = append(next, t)
		}

		// Before enqueuing the next batch, sort by descending object
		// size.
		sort.Sort(sort.Reverse(next))

		done := make(chan struct{})

		var retries batch
		var err error

		go func() {
			defer close(done)

			if len(next) == 0 {
				return
			}

			retries, err = q.enqueueAndCollectRetriesFor(next)
			if err != nil {
				q.errorc <- err
			}
		}()

		var collected batch
		collected, closing = q.collectPendingUntil(done)

		// If we've encountered a serious error here, abort immediately;
		// don't process further batches.  Abort the wait queue so that
		// we don't deadlock waiting for objects to complete when they
		// never will.
		if err != nil && !errors.IsRetriableError(err) {
			q.wait.Abort()
			break
		}

		// Ensure the next batch is filled with, in order:
		//
		// - retries from the previous batch,
		// - new additions that were enqueued behind retries, &
		// - items collected while the batch was processing.
		var minWaitTime time.Duration
		next, pending, minWaitTime = retries.Concat(append(pending, collected...), q.batchSize)
		if len(next) == 0 && len(pending) != 0 {
			// There are some pending that cound not be queued.
			// Wait the requested time before resuming loop.
			time.Sleep(minWaitTime)
		} else if len(next) == 0 && len(pending) == 0 && closing {
			// There are no items remaining, it is safe to break
			break
		}
	}
}

// collectPendingUntil collects items from q.incoming into a "pending" batch
// until the given "done" channel is written to, or is closed.
//
// A "pending" batch is returned, along with whether or not "q.incoming" is
// closed.
func (q *TransferQueue) collectPendingUntil(done <-chan struct{}) (pending batch, closing bool) {
	for {
		select {
		case t, ok := <-q.incoming:
			if !ok {
				closing = true
				<-done
				return
			}

			pending = append(pending, t)
		case <-done:
			return
		}
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
func (q *TransferQueue) enqueueAndCollectRetriesFor(batch batch) (batch, error) {
	next := q.makeBatch()
	tracerx.Printf("tq: sending batch of size %d", len(batch))

	enqueueRetry := func(t *objectTuple, err error, readyTime *time.Time) {
		count := q.rc.Increment(t.Oid)

		if readyTime == nil {
			t.ReadyTime = q.rc.ReadyTime(t.Oid)
		} else {
			t.ReadyTime = *readyTime
		}
		delay := time.Until(t.ReadyTime).Seconds()

		var errMsg string
		if err != nil {
			errMsg = fmt.Sprintf(": %s", err)
		}
		tracerx.Printf("tq: enqueue retry #%d after %.2fs for %q (size: %d)%s", count, delay, t.Oid, t.Size, errMsg)
		next = append(next, t)
	}

	q.meter.Pause()
	var bRes *BatchResponse
	if q.manifest.standaloneTransferAgent != "" {
		// Trust the external transfer agent can do everything by itself.
		objects := make([]*Transfer, 0, len(batch))
		for _, t := range batch {
			objects = append(objects, &Transfer{Oid: t.Oid, Size: t.Size, Path: t.Path, Missing: t.Missing})
		}
		bRes = &BatchResponse{
			Objects:             objects,
			TransferAdapterName: q.manifest.standaloneTransferAgent,
		}
	} else {
		// Query the Git LFS server for what transfer method to use and
		// details such as URLs, authentication, etc.
		var err error
		bRes, err = Batch(q.manifest, q.direction, q.remote, q.ref, batch.ToTransfers())
		if err != nil {
			// If there was an error making the batch API call, mark all of
			// the objects for retry, and return them along with the error
			// that was encountered. If any of the objects couldn't be
			// retried, they will be marked as failed.
			for _, t := range batch {
				if q.canRetryObject(t.Oid, err) {
					enqueueRetry(t, err, nil)
				} else if readyTime, canRetry := q.canRetryObjectLater(t.Oid, err); canRetry {
					err = nil
					enqueueRetry(t, err, &readyTime)
				} else {
					q.wait.Done()
				}
			}

			return next, errors.NewRetriableError(err)
		}
	}

	if len(bRes.Objects) == 0 {
		return next, nil
	}

	// We check first that all of the objects we want to upload are present,
	// and abort if any are missing. We'll never have any objects marked as
	// missing except possibly on upload, so just skip iterating over the
	// objects in that case.
	if q.direction == Upload {
		for _, o := range bRes.Objects {
			// If the server already has the object, the list of
			// actions will be empty. It's fine if the file is
			// missing in that case, since we don't need to upload
			// it.
			if o.Missing && len(o.Actions) != 0 {
				return nil, errors.Errorf("Unable to find source for object %v (try running git lfs fetch --all)", o.Oid)
			}
		}
	}

	q.useAdapter(bRes.TransferAdapterName)
	q.meter.Start()

	toTransfer := make([]*Transfer, 0, len(bRes.Objects))

	for _, o := range bRes.Objects {
		if o.Error != nil {
			q.errorc <- errors.Wrapf(o.Error, "[%v] %v", o.Oid, o.Error.Message)
			q.Skip(o.Size)
			q.wait.Done()

			continue
		}

		q.trMutex.Lock()
		objects, ok := q.transfers[o.Oid]
		q.trMutex.Unlock()
		if !ok {
			// If we couldn't find any associated
			// Transfer object, then we give up on the
			// transfer by telling the progress meter to
			// skip the number of bytes in "o".
			q.errorc <- errors.Errorf("[%v] The server returned an unknown OID.", o.Oid)

			q.Skip(o.Size)
			q.wait.Done()
		} else {
			// Pick t[0], since it will cover all transfers with the
			// same OID.
			tr := newTransfer(o, objects.First().Name, objects.First().Path)

			if a, err := tr.Rel(q.direction.String()); err != nil {
				if q.canRetryObject(tr.Oid, err) {
					enqueueRetry(objects.First(), err, nil)
				} else {
					q.errorc <- errors.Errorf("[%v] %v", tr.Name, err)

					q.Skip(o.Size)
					q.wait.Done()
				}
			} else if a == nil && q.manifest.standaloneTransferAgent == "" {
				q.Skip(o.Size)
				q.wait.Done()
			} else {
				q.meter.StartTransfer(objects.First().Name)
				toTransfer = append(toTransfer, tr)
			}
		}
	}

	retries := q.addToAdapter(bRes.endpoint, toTransfer)
	for t := range retries {
		enqueueRetry(t, nil, nil)
	}

	return next, nil
}

// makeBatch returns a new, empty batch, with a capacity equal to the maximum
// batch size designated by the `*TransferQueue`.
func (q *TransferQueue) makeBatch() batch { return make(batch, 0, q.batchSize) }

// addToAdapter adds the given "pending" transfers to the transfer adapters and
// returns a channel of Transfers that are to be retried in the next batch.
// After all of the items in the batch have been processed, the channel is
// closed.
//
// addToAdapter returns immediately, and does not block.
func (q *TransferQueue) addToAdapter(e lfshttp.Endpoint, pending []*Transfer) <-chan *objectTuple {
	retries := make(chan *objectTuple, len(pending))

	if err := q.ensureAdapterBegun(e); err != nil {
		close(retries)

		q.errorc <- err
		for _, t := range pending {
			q.Skip(t.Size)
			q.wait.Done()
		}

		return retries
	}

	present, missingResults := q.partitionTransfers(pending)

	go func() {
		defer close(retries)

		var results <-chan TransferResult
		if q.dryRun {
			results = q.makeDryRunResults(present)
		} else {
			results = q.adapter.Add(present...)
		}

		for _, res := range missingResults {
			q.handleTransferResult(res, retries)
		}
		for res := range results {
			q.handleTransferResult(res, retries)
		}
	}()

	return retries
}

func (q *TransferQueue) partitionTransfers(transfers []*Transfer) (present []*Transfer, results []TransferResult) {
	if q.direction != Upload {
		return transfers, nil
	}

	present = make([]*Transfer, 0, len(transfers))
	results = make([]TransferResult, 0, len(transfers))

	for _, t := range transfers {
		var err error

		if t.Size < 0 {
			err = errors.Errorf("Git LFS: object %q has invalid size (got: %d)", t.Oid, t.Size)
		} else {
			fd, serr := os.Stat(t.Path)
			if serr != nil {
				if os.IsNotExist(serr) {
					err = newObjectMissingError(t.Name, t.Oid)
				} else {
					err = serr
				}
			} else if t.Size != fd.Size() {
				err = newCorruptObjectError(t.Name, t.Oid)
			}
		}

		if err != nil {
			results = append(results, TransferResult{
				Transfer: t,
				Error:    err,
			})
		} else {
			present = append(present, t)
		}
	}

	return
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
	res TransferResult, retries chan<- *objectTuple,
) {
	oid := res.Transfer.Oid

	if res.Error != nil {
		// If there was an error encountered when processing the
		// transfer (res.Transfer), handle the error as is appropriate:
		if readyTime, canRetry := q.canRetryObjectLater(oid, res.Error); canRetry {
			// If the object can't be retried now, but can be
			// after a certain period of time, send it to
			// the retry channel with a time when it's ready.
			tracerx.Printf("tq: retrying object %s after %s seconds.", oid, time.Until(readyTime).Seconds())
			q.trMutex.Lock()
			objects, ok := q.transfers[oid]
			q.trMutex.Unlock()

			if ok {
				t := objects.First()
				t.ReadyTime = readyTime
				retries <- t
			} else {
				q.errorc <- res.Error
			}
		} else if q.canRetryObject(oid, res.Error) {
			// If the object can be retried, send it on the retries
			// channel, where it will be read at the call-site and
			// its retry count will be incremented.
			tracerx.Printf("tq: retrying object %s: %s", oid, res.Error)

			q.trMutex.Lock()
			objects, ok := q.transfers[oid]
			q.trMutex.Unlock()

			if ok {
				retries <- objects.First()
			} else {
				q.errorc <- res.Error
			}
		} else {
			// If the error wasn't retriable, OR the object has
			// exceeded its retry budget, it will be NOT be sent to
			// the retry channel, and the error will be reported
			// immediately (unless the error is in response to a
			// HTTP 422).
			if errors.IsUnprocessableEntityError(res.Error) {
				q.unsupportedContentType = true
			} else {
				q.errorc <- res.Error
			}
			q.wait.Done()
		}
	} else {
		q.trMutex.Lock()
		objects := q.transfers[oid]
		objects.completed = true

		// Otherwise, if the transfer was successful, notify all of the
		// watchers, and mark it as finished.
		for _, c := range q.watchers {
			// Send one update for each transfer with the
			// same OID.
			for _, t := range objects.All() {
				c <- &Transfer{
					Name: t.Name,
					Path: t.Path,
					Oid:  t.Oid,
					Size: t.Size,
				}
			}
		}

		q.trMutex.Unlock()

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

// BatchSize returns the batch size of the receiving *TransferQueue, or, the
// number of transfers to accept before beginning work on them.
func (q *TransferQueue) BatchSize() int {
	return q.batchSize
}

func (q *TransferQueue) Skip(size int64) {
	q.meter.Skip(size)
}

func (q *TransferQueue) ensureAdapterBegun(e lfshttp.Endpoint) error {
	q.adapterInitMutex.Lock()
	defer q.adapterInitMutex.Unlock()

	if q.adapterInProgress {
		return nil
	}

	// Progress callback - receives byte updates
	cb := func(name string, total, read int64, current int) error {
		q.meter.TransferBytes(q.direction.String(), name, read, total, current)
		if q.cb != nil {
			// NOTE: this is the mechanism by which the logpath
			// specified by GIT_LFS_PROGRESS is written to.
			//
			// See: lfs.downloadFile() for more.
			q.cb(total, read, current)
		}
		return nil
	}

	tracerx.Printf("tq: starting transfer adapter %q", q.adapter.Name())
	err := q.adapter.Begin(q.toAdapterCfg(e), cb)
	if err != nil {
		return err
	}
	q.adapterInProgress = true

	return nil
}

func (q *TransferQueue) toAdapterCfg(e lfshttp.Endpoint) AdapterConfig {
	apiClient := q.manifest.APIClient()
	concurrency := q.manifest.ConcurrentTransfers()
	access := apiClient.Endpoints.AccessFor(e.Url)
	if access.Mode() == creds.NTLMAccess {
		concurrency = 1
	}

	return &adapterConfig{
		concurrentTransfers: concurrency,
		apiClient:           apiClient,
		remote:              q.remote,
	}
}

var (
	// contentTypeWarning is the message printed when a server returns an
	// HTTP 422 at the end of a push.
	contentTypeWarning = []string{
		"Uploading failed due to unsupported Content-Type header(s).",
		"Consider disabling Content-Type detection with:",
		"",
		"  $ git config lfs.contenttype false",
	}
)

// Wait waits for the queue to finish processing all transfers. Once Wait is
// called, Add will no longer add transfers to the queue. Any failed
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

	q.meter.Flush()
	q.errorwait.Wait()

	if q.unsupportedContentType {
		for _, line := range contentTypeWarning {
			fmt.Fprintf(os.Stderr, "info: %s\n", line)
		}
	}
}

// Watch returns a channel where the queue will write the value of each transfer
// as it completes. If multiple transfers exist with the same OID, they will all
// be recorded here, even though only one actual transfer took place. The
// channel will be closed when the queue finishes processing.
func (q *TransferQueue) Watch() chan *Transfer {
	c := make(chan *Transfer, q.batchSize)
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

// run begins the transfer queue. It transfers files sequentially or
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

// canRetryLater returns the number of seconds until an error can be retried and if the error
// is a delayed-retriable error.
func (q *TransferQueue) canRetryLater(err error) (time.Time, bool) {
	return errors.IsRetriableLaterError(err)
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

func (q *TransferQueue) canRetryObjectLater(oid string, err error) (time.Time, bool) {
	if count, ok := q.rc.CanRetry(oid); !ok {
		tracerx.Printf("tq: refusing to retry %q, too many retries (%d)", oid, count)
		return time.Time{}, false
	}

	return q.canRetryLater(err)
}

// Errors returns any errors encountered during transfer.
func (q *TransferQueue) Errors() []error {
	return q.errors
}
