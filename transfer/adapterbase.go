package transfer

import (
	"fmt"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

const (
	// objectExpirationToTransfer is the duration we expect to have passed
	// from the time that the object's expires_at property is checked to
	// when the transfer is executed.
	objectExpirationToTransfer = 5 * time.Second
)

// adapterBase implements the common functionality for core adapters which
// process transfers with N workers handling an oid each, and which wait for
// authentication to succeed on one worker before proceeding
type adapterBase struct {
	name         string
	direction    Direction
	transferImpl transferImplementation
	jobChan      chan *job
	cb           TransferProgressCallback
	outChan      chan TransferResult
	// WaitGroup to sync the completion of all workers
	workerWait sync.WaitGroup
	// WaitGroup to sync the completion of all in-flight jobs
	jobWait *sync.WaitGroup
	// WaitGroup to serialise the first transfer response to perform login if needed
	authWait sync.WaitGroup
}

// transferImplementation must be implemented to provide the actual upload/download
// implementation for all core transfer approaches that use adapterBase for
// convenience. This function will be called on multiple goroutines so it
// must be either stateless or thread safe. However it will never be called
// for the same oid in parallel.
// If authOkFunc is not nil, implementations must call it as early as possible
// when authentication succeeded, before the whole file content is transferred
type transferImplementation interface {
	// WorkerStarting is called when a worker goroutine starts to process jobs
	// Implementations can run some startup logic here & return some context if needed
	WorkerStarting(workerNum int) (interface{}, error)
	// WorkerEnding is called when a worker goroutine is shutting down
	// Implementations can clean up per-worker resources here, context is as returned from WorkerStarting
	WorkerEnding(workerNum int, ctx interface{})
	// DoTransfer performs a single transfer within a worker. ctx is any context returned from WorkerStarting
	DoTransfer(ctx interface{}, t *Transfer, cb TransferProgressCallback, authOkFunc func()) error
}

func newAdapterBase(name string, dir Direction, ti transferImplementation) *adapterBase {
	return &adapterBase{
		name:         name,
		direction:    dir,
		transferImpl: ti,

		jobWait: new(sync.WaitGroup),
	}
}

func (a *adapterBase) Name() string {
	return a.name
}

func (a *adapterBase) Direction() Direction {
	return a.direction
}

func (a *adapterBase) Begin(maxConcurrency int, cb TransferProgressCallback, completion chan TransferResult) error {
	a.cb = cb
	a.outChan = completion
	a.jobChan = make(chan *job, 100)

	tracerx.Printf("xfer: adapter %q Begin() with %d workers", a.Name(), maxConcurrency)

	a.workerWait.Add(maxConcurrency)
	a.authWait.Add(1)
	for i := 0; i < maxConcurrency; i++ {
		ctx, err := a.transferImpl.WorkerStarting(i)
		if err != nil {
			return err
		}
		go a.worker(i, ctx)
	}
	tracerx.Printf("xfer: adapter %q started", a.Name())
	return nil
}

type job struct {
	T *Transfer

	listeners []chan<- TransferResult
	wg        *sync.WaitGroup
}

func (j *job) Done(err error) {
	for _, l := range j.listeners {
		l <- TransferResult{j.T, err}
	}

	j.wg.Done()
}

func (a *adapterBase) Add(transfers ...*Transfer) <-chan TransferResult {
	results := make(chan TransferResult, len(transfers))

	listeners := []chan<- TransferResult{results}
	if a.outChan != nil {
		listeners = append(listeners, a.outChan)
	}

	a.jobWait.Add(len(transfers))

	go func() {
		defer close(results)

		for _, t := range transfers {
			// BUG(taylor): End() is race-y here, and can close
			// jobChan before we want it to
			a.jobChan <- &job{t, listeners, a.jobWait}
		}

		a.jobWait.Wait()
	}()

	return results
}

func (a *adapterBase) End() {
	tracerx.Printf("xfer: adapter %q End()", a.Name())

	a.jobWait.Wait()
	close(a.jobChan)

	// wait for all transfers to complete
	a.workerWait.Wait()
	if a.outChan != nil {
		close(a.outChan)
	}

	tracerx.Printf("xfer: adapter %q stopped", a.Name())
}

// worker function, many of these run per adapter
func (a *adapterBase) worker(workerNum int, ctx interface{}) {

	tracerx.Printf("xfer: adapter %q worker %d starting", a.Name(), workerNum)
	waitForAuth := workerNum > 0
	signalAuthOnResponse := workerNum == 0

	// First worker is the only one allowed to start immediately
	// The rest wait until successful response from 1st worker to
	// make sure only 1 login prompt is presented if necessary
	// Deliberately outside jobChan processing so we know worker 0 will process 1st item
	if waitForAuth {
		tracerx.Printf("xfer: adapter %q worker %d waiting for Auth", a.Name(), workerNum)
		a.authWait.Wait()
		tracerx.Printf("xfer: adapter %q worker %d auth signal received", a.Name(), workerNum)
	}

	for job := range a.jobChan {
		t := job.T

		var authCallback func()
		if signalAuthOnResponse {
			authCallback = func() {
				a.authWait.Done()
				signalAuthOnResponse = false
			}
		}
		tracerx.Printf("xfer: adapter %q worker %d processing job for %q", a.Name(), workerNum, t.Object.Oid)

		// transferTime is the time that we are to compare the transfer's
		// `expired_at` property against.
		//
		// We add the `objectExpirationToTransfer` since there will be
		// some time lost from this comparison to the time we actually
		// transfer the object
		transferTime := time.Now().Add(objectExpirationToTransfer)

		// Actual transfer happens here
		var err error
		if expAt, expired := t.Object.IsExpired(transferTime); expired {
			tracerx.Printf("xfer: adapter %q worker %d found job for %q expired, retrying...", a.Name(), workerNum, t.Object.Oid)
			err = errors.NewRetriableError(errors.Errorf(
				"lfs/transfer: object %q expires at %s",
				t.Object.Oid, expAt.In(time.Local).Format(time.RFC822),
			))
		} else if t.Object.Size < 0 {
			tracerx.Printf("xfer: adapter %q worker %d found invalid size for %q (got: %d), retrying...", a.Name(), workerNum, t.Object.Oid, t.Object.Size)
			err = fmt.Errorf("Git LFS: object %q has invalid size (got: %d)", t.Object.Oid, t.Object.Size)
		} else {
			err = a.transferImpl.DoTransfer(ctx, t, a.cb, authCallback)
		}

		// Mark the job as completed, and alter all listeners
		job.Done(err)

		tracerx.Printf("xfer: adapter %q worker %d finished job for %q", a.Name(), workerNum, t.Object.Oid)
	}
	// This will only happen if no jobs were submitted; just wake up all workers to finish
	if signalAuthOnResponse {
		a.authWait.Done()
	}
	tracerx.Printf("xfer: adapter %q worker %d stopping", a.Name(), workerNum)
	a.transferImpl.WorkerEnding(workerNum, ctx)
	a.workerWait.Done()
}

func advanceCallbackProgress(cb TransferProgressCallback, t *Transfer, numBytes int64) {
	if cb != nil {
		// Must split into max int sizes since read count is int
		const maxInt = int(^uint(0) >> 1)
		for read := int64(0); read < numBytes; {
			remainder := numBytes - read
			if remainder > int64(maxInt) {
				read += int64(maxInt)
				cb(t.Name, t.Object.Size, read, maxInt)
			} else {
				read += remainder
				cb(t.Name, t.Object.Size, read, int(remainder))
			}

		}
	}
}
