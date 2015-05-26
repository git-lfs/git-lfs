package lfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/github/git-lfs/Godeps/_workspace/src/github.com/cheggaaa/pb"
)

var (
	clientAuthorized = int32(0)
)

// Uploadable describes a file that can be uploaded.
type Uploadable struct {
	OIDPath  string
	Filename string
	CB       CopyCallback
	Size     int64
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

	return &Uploadable{path, filename, cb, fi.Size()}, nil
}

// UploadQueue provides a queue that will allow concurrent uploads.
type UploadQueue struct {
	uploadc  chan *Uploadable
	errorc   chan *WrappedError
	errors   []*WrappedError
	wg       sync.WaitGroup
	workers  int
	files    int
	finished int64
	size     int64
	authCond *sync.Cond
}

// NewUploadQueue builds an UploadQueue, allowing `workers` concurrent uploads.
func NewUploadQueue(workers, files int) *UploadQueue {
	return &UploadQueue{
		uploadc:  make(chan *Uploadable, files),
		errorc:   make(chan *WrappedError),
		workers:  workers,
		files:    files,
		authCond: sync.NewCond(&sync.Mutex{}),
	}
}

// Add adds an Uploadable to the upload queue.
func (q *UploadQueue) Add(u *Uploadable) {
	q.wg.Add(1)
	q.size += u.Size
	q.uploadc <- u
}

// Process starts the upload queue and displays a progress bar.
func (q *UploadQueue) Process() {
	bar := pb.New64(q.size)
	bar.SetUnits(pb.U_BYTES)
	bar.ShowBar = false
	bar.Prefix(fmt.Sprintf("(%d of %d files) ", q.finished, q.files))
	bar.Start()

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

	// This will block Process() until the worker goroutines are spun up and ready
	// to process uploads.
	workersReady := make(chan int, q.workers)

	for i := 0; i < q.workers; i++ {
		// These are the worker goroutines that process uploads
		go func(n int) {
			workersReady <- 1

			for upload := range q.uploadc {
				// If an API authorization has not occured, we wait until we're woken up.
				q.authCond.L.Lock()
				if atomic.LoadInt32(&clientAuthorized) == 0 {
					q.authCond.Wait()
				}
				q.authCond.L.Unlock()

				cb := func(total, read int64, current int) error {
					bar.Add(current)
					if upload.CB != nil {
						return upload.CB(total, read, current)
					}
					return nil
				}

				err := Upload(upload.OIDPath, upload.Filename, cb)
				if err != nil {
					q.errorc <- err
				}

				f := atomic.AddInt64(&q.finished, 1)
				bar.Prefix(fmt.Sprintf("(%d of %d files) ", f, q.files))
				q.wg.Done()
			}
		}(i)
	}

	close(q.uploadc)
	<-workersReady
	q.authCond.Signal() // Signal the first goroutine to run
	q.wg.Wait()
	close(q.errorc)

	bar.Finish()
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
