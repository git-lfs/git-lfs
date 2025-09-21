package tq

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/fs"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

// Bulk transfer adapter that processes multiple files in batches
type customBulkAdapter struct {
	fs         *fs.Filesystem
	name       string
	direction  Direction
	apiClient  *lfsapi.Client
	remote     string
	path       string
	args       string
	concurrent bool
	bulkSize   int
	debugging  bool

	// transferChan is a buffered channel used to queue transfer requests (*Transfer objects)
	// for processing by the bulk adapter. Workers consume items from this channel to handle
	// transfers in batches, ensuring efficient and concurrent processing.
	transferChan chan *Transfer

	// resultChan is a buffered channel used to send the results of transfer operations.
	// Each worker writes the outcome of a transfer (success or failure) to this channel,
	// which is then consumed by the bulk adapter to handle and report transfer results.
	resultChan chan TransferResult

	// done is a channel used to signal the termination of the bulk adapter's operations.
	// It is closed when the adapter is shutting down, allowing goroutines to detect
	// and respond to the shutdown event.
	done chan struct{}

	// workerWait is a WaitGroup used to manage the lifecycle of worker goroutines.
	// It ensures that all worker processes complete their tasks before the bulk adapter
	// shuts down, preventing premature termination.
	workerWait sync.WaitGroup

	// Worker management
	workers             []*customBulkAdapterWorkerContext
	originalConcurrency int
	workerBusy          []bool // Track which workers are busy

	// Result routing for multiple concurrent Add calls
	resultRoutingMu sync.RWMutex
	pendingAddCalls map[string]chan TransferResult // Maps OID to result channel
	pendingCounts   map[chan TransferResult]int    // Maps result channel to expected count
}

type customBulkAdapterWorkerContext struct {
	workerNum   int
	cmd         *subprocess.Cmd
	stdout      io.ReadCloser
	bufferedOut *bufio.Reader
	stdin       io.WriteCloser
	errTracer   *traceWriter
}

// Bulk-specific message types
type bulkAdapterHeaderRequest struct {
	Event string `json:"event"`
	Oid   string `json:"oid"`
	Size  int64  `json:"size"`
}

type bulkAdapterTransferRequest struct {
	// common between upload/download
	Event  string  `json:"event"`
	Oid    string  `json:"oid"`
	Size   int64   `json:"size"`
	Path   string  `json:"path,omitempty"`
	Action *Action `json:"action"`
}

type bulkAdapterFooterRequest struct {
	Event string `json:"event"`
	Oid   string `json:"oid"`
	Size  int64  `json:"size"`
}

type bulkAdapterBulkResponse struct {
	Event          string       `json:"event"`
	Error          *ObjectError `json:"error"`
	Oid            string       `json:"oid"`
	Path           string       `json:"path,omitempty"`
	BytesSoFar     int64        `json:"bytesSoFar"`
	BytesSinceLast int          `json:"bytesSinceLast"`
}

// Bulk state tracking
type bulkState struct {
	id            string
	transfers     []*Transfer
	totalSize     int64
	completedSize int64
	itemsComplete int
	worker        *customBulkAdapterWorkerContext
}

func NewCustomBulkAdapter(f *fs.Filesystem, name string, dir Direction, path, args string, concurrent bool, bulkSize int) *customBulkAdapter {
	return &customBulkAdapter{
		fs:              f,
		name:            name,
		direction:       dir,
		path:            path,
		args:            args,
		concurrent:      concurrent,
		bulkSize:        bulkSize,
		transferChan:    make(chan *Transfer, bulkSize),
		resultChan:      make(chan TransferResult, bulkSize),
		done:            make(chan struct{}),
		pendingAddCalls: make(map[string]chan TransferResult),
		pendingCounts:   make(map[chan TransferResult]int),
	}
}

func newBulkCustomAdapterUploadRequest(oid string, size int64, path string, action *Action) *bulkAdapterTransferRequest {
	return &bulkAdapterTransferRequest{"upload", oid, size, path, action}
}
func newBulkCustomAdapterDownloadRequest(oid string, size int64, action *Action) *bulkAdapterTransferRequest {
	return &bulkAdapterTransferRequest{"download", oid, size, "", action}
}

// Name returns the name of this bulk adapter instance.
func (a *customBulkAdapter) Name() string {
	return a.name
}

// Direction returns the transfer direction (Upload or Download) for this adapter instance.
func (a *customBulkAdapter) Direction() Direction {
	return a.direction
}

// Begin initializes the bulk adapter with the given configuration and progress callback.
// It starts worker processes and begins the bulk processing goroutine.
func (a *customBulkAdapter) Begin(cfg AdapterConfig, cb ProgressCallback) error {
	a.apiClient = cfg.APIClient()
	a.remote = cfg.Remote()
	a.debugging = a.apiClient.OSEnv().Bool("GIT_TRANSFER_TRACE", false) ||
		a.apiClient.OSEnv().Bool("GIT_CURL_VERBOSE", false)

	maxConcurrency := cfg.ConcurrentTransfers()
	a.originalConcurrency = maxConcurrency

	if !a.concurrent {
		maxConcurrency = 1
	}

	a.Trace("xfer: bulk adapter %q Begin() with %d workers, bulk size %d", a.Name(), maxConcurrency, a.bulkSize)

	// Start worker processes
	a.workers = make([]*customBulkAdapterWorkerContext, maxConcurrency)
	a.workerBusy = make([]bool, maxConcurrency)
	for i := 0; i < maxConcurrency; i++ {
		worker, err := a.startWorker(i)
		if err != nil {
			// Clean up already started workers
			for j := 0; j < i; j++ {
				a.shutdownWorker(a.workers[j])
			}
			return err
		}
		a.workers[i] = worker
	}

	// Start the bulk processor goroutine
	a.Trace("xfer: bulk adapter starting bulk processor goroutine")
	a.workerWait.Add(1)
	go a.bulkProcessor(cb)

	a.Trace("xfer: bulk adapter %q started successfully", a.Name())
	return nil
}

// Add queues one or more transfers for bulk processing.
// Returns a channel that will receive the transfer results.
func (a *customBulkAdapter) Add(transfers ...*Transfer) <-chan TransferResult {
	a.Trace("xfer: bulk adapter Add() called with %d transfers", len(transfers))
	results := make(chan TransferResult, len(transfers))

	// Register this result channel for each transfer OID
	a.resultRoutingMu.Lock()
	a.pendingCounts[results] = len(transfers) // Track expected count for this channel
	for _, t := range transfers {
		a.Trace("xfer: bulk adapter registering result channel for OID %s", t.Oid)
		a.pendingAddCalls[t.Oid] = results
	}
	a.resultRoutingMu.Unlock()

	// Queue the transfers
	queued := 0
	for _, t := range transfers {
		a.Trace("xfer: bulk adapter attempting to queue transfer %s", t.Oid)
		select {
		case a.transferChan <- t:
			// Transfer queued successfully
			a.Trace("xfer: bulk adapter successfully queued transfer %s", t.Oid)
			queued++
		case <-a.done:
			// Adapter is shutting down - clean up and send error
			a.Trace("xfer: bulk adapter shutting down while queuing %s", t.Oid)
			a.resultRoutingMu.Lock()
			delete(a.pendingAddCalls, t.Oid)
			a.resultRoutingMu.Unlock()
			results <- TransferResult{Transfer: t, Error: errors.New("adapter shutting down")}
		}
	}

	a.Trace("xfer: bulk adapter Add() completed, queued %d/%d transfers", queued, len(transfers))
	return results
}

// routeResult sends a transfer result to the appropriate Add call's result channel
func (a *customBulkAdapter) routeResult(result TransferResult) {
	a.Trace("xfer: routing result for OID %s, error: %v", result.Transfer.Oid, result.Error)
	a.resultRoutingMu.Lock()
	resultChan, exists := a.pendingAddCalls[result.Transfer.Oid]
	if exists {
		delete(a.pendingAddCalls, result.Transfer.Oid)
	}
	a.resultRoutingMu.Unlock()

	if exists {
		a.Trace("xfer: sending result to registered channel for OID %s", result.Transfer.Oid)
		// Send to the specific Add call's result channel
		select {
		case resultChan <- result:
			a.Trace("xfer: result sent successfully for OID %s", result.Transfer.Oid)

			// Check if this was the last result for this channel
			a.resultRoutingMu.Lock()
			if count, hasCount := a.pendingCounts[resultChan]; hasCount {
				a.pendingCounts[resultChan] = count - 1
				if a.pendingCounts[resultChan] <= 0 {
					// All results sent, close the channel
					delete(a.pendingCounts, resultChan)
					close(resultChan)
					a.Trace("xfer: closed result channel after sending all results")
				}
			}
			a.resultRoutingMu.Unlock()
		case <-a.done:
			a.Trace("xfer: adapter shutting down while sending result for OID %s", result.Transfer.Oid)
		}
	} else {
		// No registered Add call for this OID - this shouldn't happen in normal operation
		a.Trace("xfer: received result for unregistered OID %s", result.Transfer.Oid)
	}
}

// End signals that no more transfers will be added and waits for all processing to complete.
// It gracefully shuts down all worker processes and closes channels.
func (a *customBulkAdapter) End() {
	a.Trace("xfer: bulk adapter End() called - closing transfer channel")
	close(a.transferChan)

	a.Trace("xfer: bulk adapter End() - closing done channel to signal shutdown")
	close(a.done)

	a.Trace("xfer: bulk adapter End() - waiting for workers to complete")
	a.workerWait.Wait()
	a.Trace("xfer: bulk adapter End() - all workers completed")

	a.Trace("xfer: bulk adapter End() - closing result channel")
	close(a.resultChan)

	// Close any remaining result channels
	a.resultRoutingMu.Lock()
	for resultChan := range a.pendingCounts {
		close(resultChan)
		a.Trace("xfer: closed remaining result channel during shutdown")
	}
	a.pendingCounts = make(map[chan TransferResult]int)
	a.pendingAddCalls = make(map[string]chan TransferResult)
	a.resultRoutingMu.Unlock()

	a.Trace("xfer: bulk adapter End() - shutting down %d workers", len(a.workers))
	// Shutdown all workers
	for i, worker := range a.workers {
		if worker != nil {
			a.Trace("xfer: bulk adapter End() - shutting down worker %d", i)
			a.shutdownWorker(worker)
		}
	}
	a.Trace("xfer: bulk adapter End() completed")
}

// startWorker creates and initializes a new worker process for handling bulk transfers.
// It establishes communication pipes and sends the initialization message.
func (a *customBulkAdapter) startWorker(workerNum int) (*customBulkAdapterWorkerContext, error) {
	a.Trace("xfer: starting bulk transfer process %q for worker %d", a.name, workerNum)

	cmdName, cmdArgs := subprocess.FormatForShell(subprocess.ShellQuoteSingle(a.path), a.args)
	cmd, err := subprocess.ExecCommand(cmdName, cmdArgs...)
	if err != nil {
		return nil, errors.New(tr.Tr.Get("failed to find bulk transfer command %q: %v", a.path, err))
	}

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.New(tr.Tr.Get("failed to get stdout for bulk transfer command %q: %v", a.path, err))
	}

	inp, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.New(tr.Tr.Get("failed to get stdin for bulk transfer command %q: %v", a.path, err))
	}

	// Capture stderr to trace
	tracer := &traceWriter{}
	tracer.processName = filepath.Base(a.path)
	cmd.Stderr = tracer

	err = cmd.Start()
	if err != nil {
		return nil, errors.New(tr.Tr.Get("failed to start bulk transfer command %q: %v", a.path, err))
	}

	worker := &customBulkAdapterWorkerContext{
		workerNum:   workerNum,
		cmd:         cmd,
		stdout:      outp,
		bufferedOut: bufio.NewReader(outp),
		stdin:       inp,
		errTracer:   tracer,
	}

	// Send initialization message
	initReq := NewCustomAdapterInitRequest(
		a.getOperationName(), a.remote, a.concurrent, a.originalConcurrency,
	)
	resp, err := a.exchangeMessage(worker, initReq)
	if err != nil {
		a.abortWorker(worker)
		return nil, err
	}
	if resp.Error != nil {
		a.abortWorker(worker)
		return nil, errors.New(tr.Tr.Get("error initializing bulk adapter %q worker %d: %v", a.name, workerNum, resp.Error))
	}

	a.Trace("xfer: started bulk adapter process %q for worker %d OK", a.path, workerNum)
	return worker, nil
}

// findAvailableWorker returns the index of an available worker, or -1 if none available
func (a *customBulkAdapter) findAvailableWorker() int {
	for i, busy := range a.workerBusy {
		if !busy {
			return i
		}
	}
	return -1
}

// markWorkerBusy marks a worker as busy or free
func (a *customBulkAdapter) markWorkerBusy(workerIndex int, busy bool) {
	if workerIndex >= 0 && workerIndex < len(a.workerBusy) {
		a.workerBusy[workerIndex] = busy
	}
}

// bulkProcessor is the main goroutine that groups incoming transfers into bulks
// and distributes them to available workers for processing.
func (a *customBulkAdapter) bulkProcessor(cb ProgressCallback) {
	defer a.workerWait.Done()
	a.Trace("xfer: bulk processor started")

	var pendingTransfers []*Transfer

	// Timer to periodically process pending transfers even if bulk isn't full
	flushTimer := time.NewTimer(100 * time.Millisecond)
	defer flushTimer.Stop()

	for {
		select {
		case transfer, ok := <-a.transferChan:
			if !ok {
				// Channel closed, process remaining transfers
				a.Trace("xfer: transfer channel closed, processing %d remaining transfers", len(pendingTransfers))
				if len(pendingTransfers) > 0 {
					// For shutdown, use any available worker (block if needed)
					availableWorker := a.findAvailableWorker()
					if availableWorker >= 0 {
						a.markWorkerBusy(availableWorker, true)
						a.processBulk(pendingTransfers, availableWorker, cb)
					} else {
						// Wait for a worker to become available
						a.Trace("xfer: waiting for worker to become available for final bulk")
						for {
							availableWorker = a.findAvailableWorker()
							if availableWorker >= 0 {
								a.markWorkerBusy(availableWorker, true)
								a.processBulk(pendingTransfers, availableWorker, cb)
								break
							}
							time.Sleep(10 * time.Millisecond)
						}
					}
				}
				a.Trace("xfer: bulk processor exiting")
				return
			}

			a.Trace("xfer: received transfer %s, pending count: %d", transfer.Oid, len(pendingTransfers)+1)
			pendingTransfers = append(pendingTransfers, transfer)

			// Process bulk when we reach the bulk size
			if len(pendingTransfers) >= a.bulkSize {
				// Find available worker
				availableWorker := a.findAvailableWorker()
				if availableWorker >= 0 {
					a.Trace("xfer: bulk size reached (%d), processing with worker %d", len(pendingTransfers), availableWorker)
					a.markWorkerBusy(availableWorker, true)
					a.processBulk(pendingTransfers, availableWorker, cb)
					pendingTransfers = nil
					// Reset timer since we just processed
					if !flushTimer.Stop() {
						<-flushTimer.C
					}
					flushTimer.Reset(100 * time.Millisecond)
				} else {
					a.Trace("xfer: bulk size reached but no workers available, waiting...")
					// Don't reset timer, will try again on next timer tick
				}
			} else if len(pendingTransfers) == 1 {
				// Start/restart timer when we get the first transfer in a new batch
				if !flushTimer.Stop() {
					<-flushTimer.C
				}
				flushTimer.Reset(100 * time.Millisecond)
			}

		case <-flushTimer.C:
			// Timer expired, process any pending transfers
			if len(pendingTransfers) > 0 {
				availableWorker := a.findAvailableWorker()
				if availableWorker >= 0 {
					a.Trace("xfer: timer expired, processing %d pending transfers with worker %d", len(pendingTransfers), availableWorker)
					a.markWorkerBusy(availableWorker, true)
					a.processBulk(pendingTransfers, availableWorker, cb)
					pendingTransfers = nil
				} else {
					a.Trace("xfer: timer expired but no workers available, waiting...")
				}
			}
			// Reset timer for next batch
			flushTimer.Reset(100 * time.Millisecond)

		case <-a.done:
			a.Trace("xfer: bulk processor received done signal, processing %d remaining transfers", len(pendingTransfers))
			if len(pendingTransfers) > 0 {
				// For shutdown, use any available worker (block if needed)
				availableWorker := a.findAvailableWorker()
				if availableWorker >= 0 {
					a.markWorkerBusy(availableWorker, true)
					a.processBulk(pendingTransfers, availableWorker, cb)
				} else {
					// Wait for a worker to become available
					a.Trace("xfer: waiting for worker to become available for final bulk")
					for {
						availableWorker = a.findAvailableWorker()
						if availableWorker >= 0 {
							a.markWorkerBusy(availableWorker, true)
							a.processBulk(pendingTransfers, availableWorker, cb)
							break
						}
						time.Sleep(10 * time.Millisecond)
					}
				}
			}
			a.Trace("xfer: bulk processor exiting due to done signal")
			return
		}
	}
}

// processBulk handles the complete bulk transfer process for a group of transfers.
// It sends the bulk definition (header, items, footer) and processes all responses.
func (a *customBulkAdapter) processBulk(transfers []*Transfer, workerIndex int, cb ProgressCallback) {
	if len(transfers) == 0 {
		a.Trace("xfer: processBulk called with empty transfers")
		return
	}

	a.Trace("xfer: processBulk starting with %d transfers, worker %d", len(transfers), workerIndex)
	worker := a.workers[workerIndex]
	bulk := &bulkState{
		id:        a.generateBulkId(),
		transfers: transfers,
		worker:    worker,
	}

	// Calculate total size
	for _, t := range transfers {
		bulk.totalSize += t.Size
	}

	a.Trace("xfer: processing bulk %s with %d transfers, total size %d", bulk.id, len(transfers), bulk.totalSize)

	// Send bulk header
	headerReq := &bulkAdapterHeaderRequest{
		Event: "bulk-header",
		Oid:   bulk.id,
		Size:  int64(len(transfers)),
	}

	if err := a.sendMessage(worker, headerReq); err != nil {
		a.markWorkerBusy(workerIndex, false)
		a.failBulk(bulk, err)
		return
	}

	// Send individual transfer requests
	for _, t := range transfers {
		rel, err := t.Rel(a.getOperationName())
		if err != nil {
			a.markWorkerBusy(workerIndex, false)
			a.failBulk(bulk, err)
			return
		}

		var req *bulkAdapterTransferRequest
		if a.direction == Upload {
			req = newBulkCustomAdapterUploadRequest(t.Oid, t.Size, t.Path, rel)
		} else {
			req = newBulkCustomAdapterDownloadRequest(t.Oid, t.Size, rel)
		}

		if err := a.sendMessage(worker, req); err != nil {
			a.markWorkerBusy(workerIndex, false)
			a.failBulk(bulk, err)
			return
		}
	}

	// Send bulk footer
	a.Trace("xfer: sending bulk footer for bulk %s", bulk.id)
	footerReq := &bulkAdapterFooterRequest{
		Event: "bulk-footer",
		Oid:   bulk.id,
		Size:  bulk.totalSize,
	}

	if err := a.sendMessage(worker, footerReq); err != nil {
		a.Trace("xfer: failed to send bulk footer: %v", err)
		a.markWorkerBusy(workerIndex, false)
		a.failBulk(bulk, err)
		return
	}

	// Process responses in parallel - don't block the bulkProcessor
	a.Trace("xfer: starting to process bulk responses for bulk %s in parallel", bulk.id)
	a.workerWait.Add(1)
	go func() {
		defer a.workerWait.Done()
		defer a.markWorkerBusy(workerIndex, false) // Free the worker when done
		a.processBulkResponses(bulk, cb)
		a.Trace("xfer: processBulk completed for bulk %s", bulk.id)
	}()
}

// processBulkResponses handles all response messages from the external process
// for a bulk transfer, including progress updates, item completions, and bulk completion.
func (a *customBulkAdapter) processBulkResponses(bulk *bulkState, cb ProgressCallback) {
	a.Trace("xfer: processBulkResponses starting for bulk %s with %d transfers", bulk.id, len(bulk.transfers))
	itemsCompleted := make(map[string]*Transfer)
	var bulkComplete bool

	for !bulkComplete {
		a.Trace("xfer: waiting for response from bulk adapter...")
		resp, err := a.readResponse(bulk.worker)
		if err != nil {
			a.Trace("xfer: error reading response: %v", err)
			a.failBulk(bulk, err)
			return
		}

		a.Trace("xfer: received response: event=%s, oid=%s, error=%v", resp.Event, resp.Oid, resp.Error)
		switch resp.Event {
		case "progress":
			if resp.Oid == bulk.id {
				a.Trace("xfer: bulk progress update: %d/%d bytes", resp.BytesSoFar, bulk.totalSize)
				// Bulk progress
				if cb != nil {
					// Report progress for each transfer proportionally
					for _, t := range bulk.transfers {
						proportion := float64(t.Size) / float64(bulk.totalSize)
						transferProgress := int64(float64(resp.BytesSoFar) * proportion)
						transferIncrement := int(float64(resp.BytesSinceLast) * proportion)
						cb(t.Name, t.Size, transferProgress, transferIncrement)
					}
				}
				bulk.completedSize = resp.BytesSoFar
			}

		case "complete":
			a.Trace("xfer: item completion received for OID %s", resp.Oid)
			// Individual item completed
			transfer := a.findTransferByOid(bulk.transfers, resp.Oid)
			if transfer == nil {
				a.Trace("xfer: unexpected OID %s in complete event", resp.Oid)
				a.failBulk(bulk, errors.New(tr.Tr.Get("unexpected OID %q in complete", resp.Oid)))
				return
			}

			itemsCompleted[resp.Oid] = transfer
			bulk.itemsComplete++
			a.Trace("xfer: marked item %s as complete (%d/%d)", resp.Oid, bulk.itemsComplete, len(bulk.transfers))

			if resp.Error != nil {
				// Individual item failed
				a.routeResult(TransferResult{
					Transfer: transfer,
					Error:    errors.New(tr.Tr.Get("error transferring %q: %v", transfer.Oid, resp.Error)),
				})
			} else {
				// Individual item succeeded
				if a.direction == Download {
					// Verify and move downloaded file
					if err := tools.VerifyFileHash(transfer.Oid, resp.Path); err != nil {
						a.routeResult(TransferResult{
							Transfer: transfer,
							Error:    errors.New(tr.Tr.Get("downloaded file failed checks: %v", err)),
						})
						continue
					}

					if err := tools.RenameFileCopyPermissions(resp.Path, transfer.Path); err != nil {
						a.routeResult(TransferResult{
							Transfer: transfer,
							Error:    errors.New(tr.Tr.Get("failed to copy downloaded file: %v", err)),
						})
						continue
					}
				} else if a.direction == Upload {
					if err := verifyUpload(a.apiClient, a.remote, transfer); err != nil {
						a.routeResult(TransferResult{
							Transfer: transfer,
							Error:    err,
						})
						continue
					}
				}

				a.routeResult(TransferResult{Transfer: transfer, Error: nil})
			}

		case "bulk-complete":
			a.Trace("xfer: bulk completion received for bulk %s", resp.Oid)
			if resp.Oid != bulk.id {
				a.Trace("xfer: unexpected bulk ID %s, expected %s", resp.Oid, bulk.id)
				a.failBulk(bulk, errors.New(tr.Tr.Get("unexpected bulk ID %q in bulk-complete", resp.Oid)))
				return
			}

			if resp.Error != nil {
				a.Trace("xfer: bulk transfer failed: %v", resp.Error)
				a.failBulk(bulk, errors.New(tr.Tr.Get("bulk transfer failed: %v", resp.Error)))
				return
			}

			// Check that all items were completed
			if bulk.itemsComplete != len(bulk.transfers) {
				a.Trace("xfer: bulk completed but only %d of %d items finished", bulk.itemsComplete, len(bulk.transfers))
				a.failBulk(bulk, errors.New(tr.Tr.Get("bulk completed but only %d of %d items finished", bulk.itemsComplete, len(bulk.transfers))))
				return
			}

			a.Trace("xfer: bulk %s completed successfully with all %d items", bulk.id, bulk.itemsComplete)
			bulkComplete = true

		default:
			a.Trace("xfer: invalid message %s from bulk adapter", resp.Event)
			a.failBulk(bulk, errors.New(tr.Tr.Get("invalid message %q from bulk adapter", resp.Event)))
			return
		}
	}

	a.Trace("xfer: processBulkResponses completed successfully for bulk %s", bulk.id)
}

// failBulk marks all transfers in a bulk as failed with the given error.
// It sends error results for each transfer in the bulk to the result channel.
func (a *customBulkAdapter) failBulk(bulk *bulkState, err error) {
	a.Trace("xfer: bulk %s failed: %v", bulk.id, err)

	for _, t := range bulk.transfers {
		a.routeResult(TransferResult{Transfer: t, Error: err})
	}
}

// findTransferByOid searches for a transfer with the given OID within a slice of transfers.
// Returns the matching transfer or nil if not found.
func (a *customBulkAdapter) findTransferByOid(transfers []*Transfer, oid string) *Transfer {
	for _, t := range transfers {
		if t.Oid == oid {
			return t
		}
	}
	return nil
}

// generateBulkId creates a unique identifier for a bulk transfer using random bytes.
// The ID is used to track the bulk throughout its lifecycle.
func (a *customBulkAdapter) generateBulkId() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("bulk-%s", hex.EncodeToString(bytes))
}

// getOperationName returns the operation name string ("download" or "upload")
// based on the adapter's direction for use in protocol messages.
func (a *customBulkAdapter) getOperationName() string {
	if a.direction == Download {
		return "download"
	}
	return "upload"
}

// Message handling methods (reuse from custom adapter)
// sendMessage serializes a message to JSON and sends it to the worker process via stdin.
// Each message is sent on a single line followed by a newline (line-delimited JSON).
func (a *customBulkAdapter) sendMessage(worker *customBulkAdapterWorkerContext, req interface{}) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	a.Trace("xfer: Bulk adapter worker %d sending message: %v", worker.workerNum, string(b))
	// Line oriented JSON
	b = append(b, '\n')
	_, err = worker.stdin.Write(b)
	return err
}

// readResponse reads a single line-delimited JSON response from the worker process stdout
// and deserializes it into a bulkAdapterBulkResponse struct.
func (a *customBulkAdapter) readResponse(worker *customBulkAdapterWorkerContext) (*bulkAdapterBulkResponse, error) {
	line, err := worker.bufferedOut.ReadString('\n')
	if err != nil {
		return nil, err
	}
	a.Trace("xfer: Bulk adapter worker %d received response: %v", worker.workerNum, strings.TrimSpace(line))
	resp := &bulkAdapterBulkResponse{}
	err = json.Unmarshal([]byte(line), resp)
	return resp, err
}

// exchangeMessage sends a message to the worker process and waits for a response.
// This is a convenience method that combines sendMessage and readResponse.
func (a *customBulkAdapter) exchangeMessage(worker *customBulkAdapterWorkerContext, req interface{}) (*bulkAdapterBulkResponse, error) {
	err := a.sendMessage(worker, req)
	if err != nil {
		return nil, err
	}
	return a.readResponse(worker)
}

// shutdownWorker gracefully terminates a worker process by sending a terminate message
// and waiting for the process to exit. Returns an error if shutdown fails or times out.
func (a *customBulkAdapter) shutdownWorker(worker *customBulkAdapterWorkerContext) error {
	defer worker.errTracer.Flush()

	a.Trace("xfer: Shutting down bulk adapter worker %d", worker.workerNum)

	finishChan := make(chan error, 1)
	go func() {
		termReq := NewCustomAdapterTerminateRequest()
		err := a.sendMessage(worker, termReq)
		if err != nil {
			finishChan <- err
		}
		worker.stdin.Close()
		worker.stdout.Close()
		finishChan <- worker.cmd.Wait()
	}()

	select {
	case err := <-finishChan:
		return err
	case <-time.After(30 * time.Second):
		return errors.New(tr.Tr.Get("timeout while shutting down bulk worker process %d", worker.workerNum))
	}
}

// abortWorker forcefully terminates a worker process by closing pipes and killing the process.
// This is used when graceful shutdown fails or when an error occurs during initialization.
func (a *customBulkAdapter) abortWorker(worker *customBulkAdapterWorkerContext) {
	a.Trace("xfer: Aborting bulk worker process: %d", worker.workerNum)
	worker.stdin.Close()
	worker.stdout.Close()
	worker.cmd.Process.Kill()
}

// Trace outputs debug trace messages if debugging is enabled.
// It uses the same format as fmt.Printf for consistency with other adapters.
func (a *customBulkAdapter) Trace(format string, args ...interface{}) {
	if a.debugging {
		tracerx.Printf(format, args...)
	}
}

// Configuration and registration
// configureBulkAdapters scans the git configuration for bulk transfer adapter definitions
// and registers them with the manifest. It looks for lfs.bulk.transfer.<name>.* settings.
func configureBulkAdapters(git Env, m *concreteManifest) {
	pathRegex := regexp.MustCompile(`lfs\.bulk\.transfer\.([^.]+)\.path`)
	tracerx.Printf("configureBulkAdapters: scanning git config for bulk adapters")
	for k, _ := range git.All() {
		tracerx.Printf("configureBulkAdapters: checking config key: %s", k)
		match := pathRegex.FindStringSubmatch(k)
		if match == nil {
			continue
		}

		name := match[1]
		path, _ := git.Get(k)
		tracerx.Printf("configureBulkAdapters: found bulk adapter '%s' with path '%s'", name, path)
		// retrieve other values
		args, _ := git.Get(fmt.Sprintf("lfs.bulk.transfer.%s.args", name))
		concurrent := git.Bool(fmt.Sprintf("lfs.bulk.transfer.%s.concurrent", name), true)
		bulkSize := git.Int(fmt.Sprintf("lfs.bulk.transfer.%s.bulkSize", name), 100)
		direction, _ := git.Get(fmt.Sprintf("lfs.bulk.transfer.%s.direction", name))
		if len(direction) == 0 {
			direction = "both"
		} else {
			direction = strings.ToLower(direction)
		}
		tracerx.Printf("configureBulkAdapters: adapter '%s': concurrent=%t, bulkSize=%d, direction=%s", name, concurrent, bulkSize, direction)

		// Separate closure for each since we need to capture vars above
		newfunc := func(name string, dir Direction) Adapter {
			return NewCustomBulkAdapter(m.fs, name, dir, path, args, concurrent, bulkSize)
		}

		if direction == "download" || direction == "both" {
			tracerx.Printf("configureBulkAdapters: registering '%s' for Download direction", name)
			m.RegisterNewAdapterFunc(name, Download, newfunc)
		}
		if direction == "upload" || direction == "both" {
			tracerx.Printf("configureBulkAdapters: registering '%s' for Upload direction", name)
			m.RegisterNewAdapterFunc(name, Upload, newfunc)
		}
	}
	tracerx.Printf("configureBulkAdapters: completed scanning git config")
}
