package lfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/vendor/_nuts/github.com/cheggaaa/pb"
)

type Transferable interface {
	Check() (*objectResource, *WrappedError)
	Transfer(CopyCallback) *WrappedError
	Object() *objectResource
	Oid() string
	Size() int64
	Name() string
	SetObject(*objectResource)
}

// TransferQueue provides a queue that will allow concurrent transfers.
type TransferQueue struct {
	filesAdded    int32 // Number of files added to the queue
	filesFinished int32 // Number of files that have finished transfering
	transferIdx   int32 // Used to track transfer number for progress monitor
	workers       int   // Number of transfer workers to spawn
	transferKind  string
	errors        []*WrappedError
	transferables map[string]Transferable
	bar           *pb.ProgressBar
	batcher       *Batcher
	apic          chan Transferable  // Channel for processing individual API requests
	transferc     chan Transferable  // Channel for processing transfers
	errorc        chan *WrappedError // Channel for processing errors
	progressc     chan string        // Channel for GIT_LFS_PROGRESS monitor
	watchers      []chan string
	wait          sync.WaitGroup
}

// newTransferQueue builds a TransferQueue, allowing `workers` concurrent transfers.
func newTransferQueue(workers int) *TransferQueue {
	q := &TransferQueue{
		apic:          make(chan Transferable, 100),
		transferc:     make(chan Transferable, 100),
		errorc:        make(chan *WrappedError),
		watchers:      make([]chan string, 0),
		progressc:     make(chan string, 100),
		workers:       workers,
		transferables: make(map[string]Transferable),
	}

	q.run()

	return q
}

// Add adds a Transferable to the transfer queue.
func (q *TransferQueue) Add(t Transferable) {
	atomic.AddInt32(&q.filesAdded, 1)

	// Sneak in and update the progress bar's total size
	atomic.AddInt64(&q.bar.Total, t.Size())

	q.wait.Add(1)
	q.transferables[t.Oid()] = t

	if q.batcher != nil {
		q.batcher.Add(t)
		return
	}

	q.apic <- t
}

// Wait waits for the queue to finish processing all transfers
func (q *TransferQueue) Wait() {

	q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", q.filesFinished, q.filesAdded))

	if q.batcher != nil {
		q.batcher.Exit()
	}
	q.wait.Wait()
	close(q.apic)
	close(q.transferc)
	close(q.errorc)
	for _, watcher := range q.watchers {
		close(watcher)
	}
	close(q.progressc)
	q.bar.Finish()
}

// Watch returns a channel where the queue will write the OID of each transfer
// as it completes. The channel will be closed when the queue finishes processing.
func (q *TransferQueue) Watch() chan string {
	c := make(chan string, 100)
	q.watchers = append(q.watchers, c)
	return c
}

// individualApiRoutine processes the queue of transfers one at a time by making
// a POST call for each object, feeding the results to the transfer workers.
// If configured, the object transfers can still happen concurrently, the
// sequential nature here is only for the meta POST calls.
func (q *TransferQueue) individualApiRoutine(apiWaiter chan interface{}) {
	for t := range q.apic {
		obj, err := t.Check()
		if err != nil {
			q.wait.Done()
			q.errorc <- err
			continue
		}

		if apiWaiter != nil { // Signal to launch more individual api workers
			select {
			case apiWaiter <- 1:
			default:
			}
		}

		if obj != nil {
			t.SetObject(obj)
			q.transferc <- t
		}
	}
}

// batchApiRoutine processes the queue of transfers using the batch endpoint,
// making only one POST call for all objects. The results are then handed
// off to the transfer workers.
func (q *TransferQueue) batchApiRoutine() {
	for {
		batch := q.batcher.Next()
		if batch == nil {
			break
		}

		transfers := make([]*objectResource, 0, len(batch))
		for _, t := range batch {
			transfers = append(transfers, &objectResource{Oid: t.Oid(), Size: t.Size()})
		}

		objects, err := Batch(transfers, q.transferKind)
		if err != nil {
			if isNotImplError(err) {
				configFile := filepath.Join(LocalGitDir, "config")
				git.Config.SetLocal(configFile, "lfs.batch", "false")
			}
			// TODO trigger the individual fallback
		}

		for _, o := range objects {
			if _, ok := o.Links[q.transferKind]; ok {
				// This object needs to be transfered
				if transfer, ok := q.transferables[o.Oid]; ok {
					transfer.SetObject(o)
					q.transferc <- transfer
				}
			}
		}
	}
}

// This goroutine collects errors returned from transfers
func (q *TransferQueue) errorCollector() {
	for err := range q.errorc {
		q.errors = append(q.errors, err)
	}
}

func (q *TransferQueue) progressMonitor() {
	output, err := newProgressLogger()
	if err != nil {
		q.errorc <- Error(err)
	}

	for l := range q.progressc {
		if err := output.Write([]byte(l)); err != nil {
			q.errorc <- Error(err)
			output.Shutdown()
		}
	}

	output.Close()
}

func (q *TransferQueue) transferWorker() {
	direction := "push"
	if q.transferKind == "download" {
		direction = "pull"
	}

	for transfer := range q.transferc {
		c := atomic.AddInt32(&q.transferIdx, 1)
		cb := func(total, read int64, current int) error {
			q.progressc <- fmt.Sprintf("%s %d/%d %d/%d %s\n", direction, c, q.filesAdded, read, total, transfer.Name())
			q.bar.Add(current)
			return nil
		}

		if err := transfer.Transfer(cb); err != nil {
			q.errorc <- err
		} else {
			oid := transfer.Oid()
			for _, c := range q.watchers {
				c <- oid
			}
		}

		f := atomic.AddInt32(&q.filesFinished, 1)
		q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", f, q.filesAdded))
		q.wait.Done()
	}
}

// launchIndividualApiRoutines first launches a single api worker. When it
// receives the first successful api request it launches workers - 1 more
// workers. This prevents being prompted for credentials multiple times at once
// when they're needed.
func (q *TransferQueue) launchIndividualApiRoutines() {
	go func() {
		apiWaiter := make(chan interface{})
		go q.individualApiRoutine(apiWaiter)

		<-apiWaiter

		for i := 0; i < q.workers-1; i++ {
			go q.individualApiRoutine(nil)
		}
	}()
}

// run starts the transfer queue and displays a progress bar. Process will
// do individual or batch transfers depending on the Config.BatchTransfer() value.
// Process will transfer files sequentially or concurrently depending on the
// Concig.ConcurrentTransfers() value.
func (q *TransferQueue) run() {
	// Set up the pb progress bar. The total size will be updated as files are
	// added to the queue.
	q.bar = pb.New64(0)
	q.bar.SetUnits(pb.U_BYTES)
	q.bar.ShowBar = false
	q.bar.Prefix(fmt.Sprintf("(%d of %d files) ", q.filesFinished, q.filesAdded))
	q.bar.Start()

	go q.errorCollector()
	go q.progressMonitor()

	for i := 0; i < q.workers; i++ {
		go q.transferWorker()
	}

	if Config.BatchTransfer() {
		q.batcher = NewBatcher(100)
		go q.batchApiRoutine()
	} else {
		q.launchIndividualApiRoutines()
	}
}

// Errors returns any errors encountered during transfer.
func (q *TransferQueue) Errors() []*WrappedError {
	return q.errors
}

// progressLogger provides a wrapper around an os.File that can either
// write to the file or ignore all writes completely.
type progressLogger struct {
	writeData bool
	log       *os.File
}

// Write will write to the file and perform a Sync() if writing succeeds.
func (l *progressLogger) Write(b []byte) error {
	if l.writeData {
		if _, err := l.log.Write(b); err != nil {
			return err
		}
		return l.log.Sync()
	}
	return nil
}

// Close will call Close() on the underlying file
func (l *progressLogger) Close() error {
	if l.log != nil {
		return l.log.Close()
	}
	return nil
}

// Shutdown will cause the logger to ignore any further writes. It should
// be used when writing causes an error.
func (l *progressLogger) Shutdown() {
	l.writeData = false
}

// newProgressLogger creates a progressLogger based on the presence of
// the GIT_LFS_PROGRESS environment variable. If it is present and a log file
// is able to be created, the logger will write to the file. If it is absent,
// or there is an err creating the file, the logger will ignore all writes.
func newProgressLogger() (*progressLogger, error) {
	logPath := Config.Getenv("GIT_LFS_PROGRESS")

	if len(logPath) == 0 {
		return &progressLogger{}, nil
	}
	if !filepath.IsAbs(logPath) {
		return &progressLogger{}, fmt.Errorf("GIT_LFS_PROGRESS must be an absolute path")
	}

	cbDir := filepath.Dir(logPath)
	if err := os.MkdirAll(cbDir, 0755); err != nil {
		return &progressLogger{}, err
	}

	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return &progressLogger{}, err
	}

	return &progressLogger{true, file}, nil
}
