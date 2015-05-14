package lfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/cheggaaa/pb"
)

// DownloadQueue provides a queue that will allow concurrent uploads.
type DownloadQueue struct {
	downloadc        chan *wrappedPointer
	errorc           chan *WrappedError
	errors           []*WrappedError
	wg               sync.WaitGroup
	workers          int
	files            int
	finished         int64
	size             int64
	authCond         *sync.Cond
	pointers         []*wrappedPointer
	bar              *pb.ProgressBar
	clientAuthorized int32
}

// NewDownloadQueue builds a DownloadQueue, allowing `workers` concurrent downloads.
func NewDownloadQueue(workers, files int) *DownloadQueue {
	return &DownloadQueue{
		downloadc: make(chan *wrappedPointer, files),
		errorc:    make(chan *WrappedError),
		workers:   workers,
		files:     files,
		authCond:  sync.NewCond(&sync.Mutex{}),
	}
}

// Add adds an object to the download queue.
func (q *DownloadQueue) Add(p *wrappedPointer) {
	q.pointers = append(q.pointers, p)
}

// apiWorker processes the queue, making the POST calls and
// feeding the results to uploadWorkers
func (q *DownloadQueue) processIndividual() {
	apic := make(chan *wrappedPointer, q.files)
	workersReady := make(chan int, q.workers)
	var wg sync.WaitGroup

	for i := 0; i < q.workers; i++ {
		go func() {
			workersReady <- 1
			for p := range apic {
				// If an API authorization has not occured, we wait until we're woken up.
				q.authCond.L.Lock()
				if atomic.LoadInt32(&q.clientAuthorized) == 0 {
					q.authCond.Wait()
				}
				q.authCond.L.Unlock()

				_, err := DownloadCheck(p.Oid)
				if err != nil {
					q.errorc <- err
					wg.Done()
					continue
				}

				q.wg.Add(1)
				q.downloadc <- p
				wg.Done()
			}
		}()
	}

	q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", q.finished, len(q.pointers)))
	q.bar.Start()

	for _, p := range q.pointers {
		wg.Add(1)
		apic <- p
	}

	<-workersReady
	q.authCond.Signal() // Signal the first goroutine to run
	close(apic)
	wg.Wait()

	close(q.downloadc)
}

// Process starts the download queue and displays a progress bar.
func (q *DownloadQueue) Process() {
	q.bar = pb.New64(q.size)
	q.bar.SetUnits(pb.U_BYTES)
	q.bar.ShowBar = false

	// This goroutine collects errors returned from downloads
	go func() {
		for err := range q.errorc {
			q.errors = append(q.errors, err)
		}
	}()

	// This goroutine watches for apiEvents. In order to prevent multiple
	// credential requests from happening, the queue is processed sequentially
	// until an API request succeeds (meaning authenication has happened successfully).
	// Once the an API request succeeds, all worker goroutines are woken up and allowed
	// to process downloads. Once a success happens, this goroutine exits.
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
		// These are the worker goroutines that process uploads
		go func(n int) {

			for ptr := range q.downloadc {
				fullPath := filepath.Join(LocalWorkingDir, ptr.Name)
				output, err := os.Create(fullPath)
				if err != nil {
					q.errorc <- Error(err)
					f := atomic.AddInt64(&q.finished, 1)
					q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", f, q.files))
					q.wg.Done()
					continue
				}

				cb := func(total, read int64, current int) error {
					q.bar.Add(current)
					return nil
				}
				// TODO need a callback
				if err := PointerSmudge(output, ptr.Pointer, ptr.Name, cb); err != nil {
					q.errorc <- Error(err)
				}

				f := atomic.AddInt64(&q.finished, 1)
				q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", f, q.files))
				q.wg.Done()
			}
		}(i)
	}

	//	if Config.BatchTransfer() {
	//		q.processBatch()
	//	} else {
	q.processIndividual()
	//	}

	q.wg.Wait()
	close(q.errorc)

	q.bar.Finish()
}

// Errors returns any errors encountered during uploading.
func (q *DownloadQueue) Errors() []*WrappedError {
	return q.errors
}
