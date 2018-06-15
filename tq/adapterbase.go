package tq

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/fs"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/rubyist/tracerx"
)

// adapterBase implements the common functionality for core adapters which
// process transfers with N workers handling an oid each, and which wait for
// authentication to succeed on one worker before proceeding
type adapterBase struct {
	fs           *fs.Filesystem
	name         string
	direction    Direction
	transferImpl transferImplementation
	apiClient    *lfsapi.Client
	remote       string
	jobChan      chan *job
	debugging    bool
	cb           ProgressCallback
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
	DoTransfer(ctx interface{}, t *Transfer, cb ProgressCallback, authOkFunc func()) error
}

func newAdapterBase(f *fs.Filesystem, name string, dir Direction, ti transferImplementation) *adapterBase {
	return &adapterBase{
		fs:           f,
		name:         name,
		direction:    dir,
		transferImpl: ti,
		jobWait:      new(sync.WaitGroup),
	}
}

func (a *adapterBase) Name() string {
	return a.name
}

func (a *adapterBase) Direction() Direction {
	return a.direction
}

func (a *adapterBase) Begin(cfg AdapterConfig, cb ProgressCallback) error {
	a.apiClient = cfg.APIClient()
	a.remote = cfg.Remote()
	a.cb = cb
	a.jobChan = make(chan *job, 100)
	a.debugging = a.apiClient.OSEnv().Bool("GIT_TRANSFER_TRACE", false)
	maxConcurrency := cfg.ConcurrentTransfers()

	a.Trace("xfer: adapter %q Begin() with %d workers", a.Name(), maxConcurrency)

	a.workerWait.Add(maxConcurrency)
	a.authWait.Add(1)
	for i := 0; i < maxConcurrency; i++ {
		ctx, err := a.transferImpl.WorkerStarting(i)
		if err != nil {
			return err
		}
		go a.worker(i, ctx)
	}
	a.Trace("xfer: adapter %q started", a.Name())
	return nil
}

type job struct {
	T *Transfer

	results chan<- TransferResult
	wg      *sync.WaitGroup
}

func (j *job) Done(err error) {
	j.results <- TransferResult{j.T, err}
	j.wg.Done()
}

func (a *adapterBase) Add(transfers ...*Transfer) <-chan TransferResult {
	results := make(chan TransferResult, len(transfers))

	a.jobWait.Add(len(transfers))

	go func() {
		for _, t := range transfers {
			a.jobChan <- &job{t, results, a.jobWait}
		}
		a.jobWait.Wait()

		close(results)
	}()

	return results
}

func (a *adapterBase) End() {
	a.Trace("xfer: adapter %q End()", a.Name())

	a.jobWait.Wait()
	close(a.jobChan)

	// wait for all transfers to complete
	a.workerWait.Wait()

	a.Trace("xfer: adapter %q stopped", a.Name())
}

func (a *adapterBase) Trace(format string, args ...interface{}) {
	if !a.debugging {
		return
	}
	tracerx.Printf(format, args...)
}

// worker function, many of these run per adapter
func (a *adapterBase) worker(workerNum int, ctx interface{}) {
	a.Trace("xfer: adapter %q worker %d starting", a.Name(), workerNum)
	waitForAuth := workerNum > 0
	signalAuthOnResponse := workerNum == 0

	// First worker is the only one allowed to start immediately
	// The rest wait until successful response from 1st worker to
	// make sure only 1 login prompt is presented if necessary
	// Deliberately outside jobChan processing so we know worker 0 will process 1st item
	if waitForAuth {
		a.Trace("xfer: adapter %q worker %d waiting for Auth", a.Name(), workerNum)
		a.authWait.Wait()
		a.Trace("xfer: adapter %q worker %d auth signal received", a.Name(), workerNum)
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
		a.Trace("xfer: adapter %q worker %d processing job for %q", a.Name(), workerNum, t.Oid)

		// Actual transfer happens here
		var err error
		if t.Size < 0 {
			err = fmt.Errorf("Git LFS: object %q has invalid size (got: %d)", t.Oid, t.Size)
		} else {
			err = a.transferImpl.DoTransfer(ctx, t, a.cb, authCallback)
		}

		// Mark the job as completed, and alter all listeners
		job.Done(err)

		a.Trace("xfer: adapter %q worker %d finished job for %q", a.Name(), workerNum, t.Oid)
	}
	// This will only happen if no jobs were submitted; just wake up all workers to finish
	if signalAuthOnResponse {
		a.authWait.Done()
	}
	a.Trace("xfer: adapter %q worker %d stopping", a.Name(), workerNum)
	a.transferImpl.WorkerEnding(workerNum, ctx)
	a.workerWait.Done()
}

var httpRE = regexp.MustCompile(`\Ahttps?://`)

func (a *adapterBase) newHTTPRequest(method string, rel *Action) (*http.Request, error) {
	if !httpRE.MatchString(rel.Href) {
		urlfragment := strings.SplitN(rel.Href, "?", 2)[0]
		return nil, fmt.Errorf("missing protocol: %q", urlfragment)
	}

	req, err := http.NewRequest(method, rel.Href, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range rel.Header {
		req.Header.Set(key, value)
	}

	return req, nil
}

func (a *adapterBase) doHTTP(t *Transfer, req *http.Request) (*http.Response, error) {
	if t.Authenticated {
		return a.apiClient.Do(req)
	}
	return a.apiClient.DoWithAuth(a.remote, req)
}

func advanceCallbackProgress(cb ProgressCallback, t *Transfer, numBytes int64) {
	if cb != nil {
		// Must split into max int sizes since read count is int
		const maxInt = int(^uint(0) >> 1)
		for read := int64(0); read < numBytes; {
			remainder := numBytes - read
			if remainder > int64(maxInt) {
				read += int64(maxInt)
				cb(t.Name, t.Size, read, maxInt)
			} else {
				read += remainder
				cb(t.Name, t.Size, read, int(remainder))
			}

		}
	}
}
