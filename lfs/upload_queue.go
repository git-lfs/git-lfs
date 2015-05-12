package lfs

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

var (
	clientAuthorized = int32(0)
)

// Uploadable describes a file that can be uploaded.
type Uploadable struct {
	OID      string
	OIDPath  string
	Filename string
	CB       CopyCallback
	Size     int64
	object   *objectResource
}

// NewUploadable builds the Uploadable from the given information.
func NewUploadable(oid, filename string, index, totalFiles int) (*Uploadable, *WrappedError) {
	path, err := LocalMediaPath(oid)
	if err != nil {
		return nil, Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	if err := ensureFile(filename, path); err != nil {
		return nil, Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	fi, err := os.Stat(filename)
	if err != nil {
		return nil, Errorf(err, "Error uploading file %s (%s)", filename, oid)
	}

	cb, file, cbErr := CopyCallbackFile("push", filename, index, totalFiles)
	if cbErr != nil {
		fmt.Fprintln(os.Stderr, cbErr.Error())
	}

	if file != nil {
		defer file.Close()
	}

	return &Uploadable{OID: oid, OIDPath: path, Filename: filename, CB: cb, Size: fi.Size()}, nil
}

// UploadQueue provides a queue that will allow concurrent uploads.
type UploadQueue struct {
	uploadc     chan *Uploadable
	errorc      chan *WrappedError
	errors      []*WrappedError
	wg          sync.WaitGroup
	workers     int
	files       int
	finished    int64
	size        int64
	authCond    *sync.Cond
	uploadables map[string]*Uploadable
	bar         *pb.ProgressBar
}

// NewUploadQueue builds an UploadQueue, allowing `workers` concurrent uploads.
func NewUploadQueue(workers, files int) *UploadQueue {
	return &UploadQueue{
		uploadc:     make(chan *Uploadable, files),
		errorc:      make(chan *WrappedError),
		workers:     workers,
		files:       files,
		authCond:    sync.NewCond(&sync.Mutex{}),
		uploadables: make(map[string]*Uploadable),
	}
}

// Add adds an Uploadable to the upload queue.
func (q *UploadQueue) Add(u *Uploadable) {
	q.uploadables[u.OID] = u
}

// apiWorker processes the queue, making the POST calls and
// feeding the results to uploadWorkers
func (q *UploadQueue) processIndividual() {
	apic := make(chan *Uploadable, q.workers)
	workersReady := make(chan int, q.workers)
	var wg sync.WaitGroup

	for i := 0; i < q.workers; i++ {
		go func() {
			workersReady <- 1
			for u := range apic {
				// If an API authorization has not occured, we wait until we're woken up.
				q.authCond.L.Lock()
				if atomic.LoadInt32(&clientAuthorized) == 0 {
					q.authCond.Wait()
				}
				q.authCond.L.Unlock()

				obj, err := UploadCheck(u.OIDPath)
				if err != nil {
					q.errorc <- err
					wg.Done()
					continue
				}
				if obj != nil {
					q.wg.Add(1)
					u.object = obj
					q.uploadc <- u
				}
				wg.Done()
			}
		}()
	}

	q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", q.finished, len(q.uploadables)))
	q.bar.Start()

	for _, u := range q.uploadables {
		wg.Add(1)
		apic <- u
	}

	<-workersReady
	q.authCond.Signal() // Signal the first goroutine to run
	close(apic)
	wg.Wait()

	close(q.uploadc)
}

// batchWorker makes the batch POST call, feeding the results
// to the uploadWorkers
func (q *UploadQueue) processBatch() {
	q.files = 0
	uploads := make([]*objectResource, 0, len(q.uploadables))
	for _, u := range q.uploadables {
		uploads = append(uploads, &objectResource{Oid: u.OID, Size: u.Size})
	}

	objects, err := Batch(uploads)
	if err != nil {
		q.errorc <- err
		sendApiEvent(apiEventFail)
		return
	}

	for _, o := range objects {
		if _, ok := o.Links["upload"]; ok {
			// This object needs to be uploaded
			if uploadable, ok := q.uploadables[o.Oid]; ok {
				q.files++
				q.wg.Add(1)
				uploadable.object = o
				q.uploadc <- uploadable
			}
		}
	}

	close(q.uploadc)
	q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", q.finished, q.files))
	q.bar.Start()
	sendApiEvent(apiEventSuccess) // Wake up upload workers
}

// Process starts the upload queue and displays a progress bar.
func (q *UploadQueue) Process() {
	q.bar = pb.New64(q.size)
	q.bar.SetUnits(pb.U_BYTES)
	q.bar.ShowBar = false

	// This goroutine collects errors returned from uploads
	go func() {
		for err := range q.errorc {
			q.errors = append(q.errors, err)
		}
	}()

	// This goroutine watches for apiEvents. In order to prevent multiple
	// credential requests from happening, the queue is processed sequentially
	// until an API request succeeds (meaning authenication has happened successfully).
	// Once the an API request succeeds, all worker goroutines are woken up and allowed
	// to process uploads. Once a success happens, this goroutine exits.
	go func() {
		for {
			event := <-apiEvent
			switch event {
			case apiEventSuccess:
				atomic.StoreInt32(&clientAuthorized, 1)
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

			for upload := range q.uploadc {
				cb := func(total, read int64, current int) error {
					q.bar.Add(current)
					if upload.CB != nil {
						return upload.CB(total, read, current)
					}
					return nil
				}

				err := UploadObject(upload.object, cb)
				if err != nil {
					q.errorc <- err
				}

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
func (q *UploadQueue) Errors() []*WrappedError {
	return q.errors
}

// ensureFile makes sure that the cleanPath exists before pushing it.  If it
// does not exist, it attempts to clean it by reading the file at smudgePath.
func ensureFile(smudgePath, cleanPath string) error {
	if _, err := os.Stat(cleanPath); err == nil {
		return nil
	}

	expectedOid := filepath.Base(cleanPath)
	localPath := filepath.Join(LocalWorkingDir, smudgePath)
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	cleaned, err := PointerClean(file, stat.Size(), nil)
	if err != nil {
		return err
	}

	cleaned.Close()

	if expectedOid != cleaned.Oid {
		return fmt.Errorf("Expected %s to have an OID of %s, got %s", smudgePath, expectedOid, cleaned.Oid)
	}

	return nil
}
