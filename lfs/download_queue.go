package lfs

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"sync"
	"sync/atomic"
)

type Downloadable struct {
	OID      string
	Size     int64
	Filename string
	CB       CopyCallback
	object   *objectResource
}

// DownloadQueue provides a queue that will allow concurrent uploads.
type DownloadQueue struct {
	downloadc        chan *Downloadable
	errorc           chan *WrappedError
	errors           []*WrappedError
	wg               sync.WaitGroup
	workers          int
	files            int
	finished         int64
	size             int64
	authCond         *sync.Cond
	downloadables    map[string]*Downloadable
	bar              *pb.ProgressBar
	clientAuthorized int32
}

// NewDownloadQueue builds a DownloadQueue, allowing `workers` concurrent downloads.
func NewDownloadQueue(workers, files int) *DownloadQueue {
	return &DownloadQueue{
		downloadc:     make(chan *Downloadable, files),
		errorc:        make(chan *WrappedError),
		workers:       workers,
		files:         files,
		authCond:      sync.NewCond(&sync.Mutex{}),
		downloadables: make(map[string]*Downloadable),
	}
}

// Add adds an object to the download queue.
func (q *DownloadQueue) Add(oid, filename string, size int64) {
	// TODO create the callback and such
	q.downloadables[oid] = &Downloadable{OID: oid, Filename: filename, Size: size}
}

// apiWorker processes the queue, making the POST calls and
// feeding the results to uploadWorkers
func (q *DownloadQueue) processIndividual() {
	apic := make(chan *Downloadable, q.workers)
	workersReady := make(chan int, q.workers)
	var wg sync.WaitGroup

	for i := 0; i < q.workers; i++ {
		go func() {
			workersReady <- 1
			for d := range apic {
				// If an API authorization has not occured, we wait until we're woken up.
				q.authCond.L.Lock()
				if atomic.LoadInt32(&q.clientAuthorized) == 0 {
					q.authCond.Wait()
				}
				q.authCond.L.Unlock()

				obj, err := DownloadCheck(d.OID)
				if err != nil {
					q.errorc <- err
					wg.Done()
					continue
				}
				if obj != nil {
					q.wg.Add(1)
					d.object = obj
					q.downloadc <- d
				}
				wg.Done()
			}
		}()
	}

	q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", q.finished, len(q.downloadables)))
	q.bar.Start()

	for _, d := range q.downloadables {
		wg.Add(1)
		apic <- d
	}

	<-workersReady
	q.authCond.Signal() // Signal the first goroutine to run
	close(apic)
	wg.Wait()

	close(q.downloadc)
}

// batchWorker makes the batch POST call, feeding the results
// to the uploadWorkers
func (q *DownloadQueue) processBatch() {
	q.files = 0
	downloads := make([]*objectResource, 0, len(q.downloadables))
	for _, d := range q.downloadables {
		downloads = append(downloads, &objectResource{Oid: d.OID, Size: d.Size})
	}

	objects, err := Batch(downloads)
	if err != nil {
		q.errorc <- err
		sendApiEvent(apiEventFail)
		return
	}

	for _, o := range objects {
		if _, ok := o.Links["download"]; ok {
			// This object can be downloaded
			if downloadable, ok := q.downloadables[o.Oid]; ok {
				q.files++
				q.wg.Add(1)
				downloadable.object = o
				q.downloadc <- downloadable
			}
		}
	}

	close(q.downloadc)
	q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", q.finished, q.files))
	q.bar.Start()
	sendApiEvent(apiEventSuccess) // Wake up download workers
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

			for download := range q.downloadc {
				_, _, err := DownloadObject(download.object)
				if err != nil {
					q.errorc <- err
				}

				// TODO: Process the download

				f := atomic.AddInt64(&q.finished, 1)
				q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", f, q.files))
				q.wg.Done()
			}
		}(i)
	}

	if Config.BatchTransfer() {
		q.processBatch()
	} else {
		q.processIndividual()
	}

	q.wg.Wait()
	close(q.errorc)

	q.bar.Finish()
}

// Errors returns any errors encountered during uploading.
func (q *DownloadQueue) Errors() []*WrappedError {
	return q.errors
}
