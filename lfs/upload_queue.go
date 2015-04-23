package lfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Uploadable describes a file that can be uploaded.
type Uploadable struct {
	OIDPath  string
	Filename string
	CB       CopyCallback
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

	cb, file, _ := CopyCallbackFile("push", filename, index, totalFiles)
	// TODO: fix this, Error() is in `commands`. This is not a fatal error, it should display
	// but not return.
	// if cbErr != nil {
	// 	Error(cbErr.Error())
	// }

	if file != nil {
		defer file.Close()
	}

	return &Uploadable{path, filename, cb}, nil
}

// UploadQueue provides a queue that will allow concurrent uploads.
type UploadQueue struct {
	uc     chan *Uploadable
	ec     chan *WrappedError
	wg     sync.WaitGroup
	errors []*WrappedError
}

// NewUploadQueue builds an UploadQueue, allowing `workers` concurrent uploads.
func NewUploadQueue(workers int) *UploadQueue {
	q := &UploadQueue{uc: make(chan *Uploadable, workers), ec: make(chan *WrappedError)}

	go func() {
		for err := range q.ec {
			q.errors = append(q.errors, err)
		}
	}()

	for i := 0; i < workers; i++ {
		go func(n int) {
			for upload := range q.uc {
				q.wg.Add(1)
				fmt.Fprintf(os.Stderr, "Uploading %s\n", upload.Filename)
				err := Upload(upload.OIDPath, upload.Filename, upload.CB)
				if err != nil {
					q.ec <- err
				}
				q.wg.Done()
			}
		}(i)
	}

	return q
}

// Upload adds an Uploadable to the upload queue. Uploads may start immediately
// when added to the queue.
func (q *UploadQueue) Upload(u *Uploadable) {
	q.uc <- u
}

// Wait waits for the upload queue to finish. Once Wait() is called, Upload() must
// not be called.
func (q *UploadQueue) Wait() {
	close(q.uc)
	q.wg.Wait()
	close(q.ec)
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
