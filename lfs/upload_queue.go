package lfs

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
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
}

// NewUploadQueue builds an UploadQueue, allowing `workers` concurrent uploads.
func NewUploadQueue(workers, files int) *UploadQueue {
	return &UploadQueue{
		uploadc: make(chan *Uploadable, files),
		errorc:  make(chan *WrappedError),
		workers: workers,
		files:   files,
	}
}

// Upload adds an Uploadable to the upload queue.
func (q *UploadQueue) Upload(u *Uploadable) {
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

	go func() {
		for err := range q.errorc {
			q.errors = append(q.errors, err)
		}
	}()

	for i := 0; i < q.workers; i++ {
		go func(n int) {
			for upload := range q.uploadc {
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
