package transfer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/tools"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/rubyist/tracerx"

	"github.com/git-lfs/git-lfs/config"
)

// Adapter for custom transfer via external process
type customAdapter struct {
	*adapterBase
	path                string
	args                string
	concurrent          bool
	originalConcurrency int
}

// Struct to capture stderr and write to trace
type traceWriter struct {
	buf         bytes.Buffer
	processName string
}

func (t *traceWriter) Write(b []byte) (int, error) {
	n, err := t.buf.Write(b)
	t.Flush()
	return n, err
}
func (t *traceWriter) Flush() {
	var err error
	for err == nil {
		var s string
		s, err = t.buf.ReadString('\n')
		if len(s) > 0 {
			tracerx.Printf("xfer[%v]: %v", t.processName, strings.TrimSpace(s))
		}
	}
}

type customAdapterWorkerContext struct {
	workerNum   int
	cmd         *exec.Cmd
	stdout      io.ReadCloser
	bufferedOut *bufio.Reader
	stdin       io.WriteCloser
	errTracer   *traceWriter
}

type customAdapterInitRequest struct {
	Event               string `json:"event"`
	Operation           string `json:"operation"`
	Concurrent          bool   `json:"concurrent"`
	ConcurrentTransfers int    `json:"concurrenttransfers"`
}

func NewCustomAdapterInitRequest(op string, concurrent bool, concurrentTransfers int) *customAdapterInitRequest {
	return &customAdapterInitRequest{"init", op, concurrent, concurrentTransfers}
}

type customAdapterTransferRequest struct { // common between upload/download
	Event  string            `json:"event"`
	Oid    string            `json:"oid"`
	Size   int64             `json:"size"`
	Path   string            `json:"path,omitempty"`
	Action *api.LinkRelation `json:"action"`
}

func NewCustomAdapterUploadRequest(oid string, size int64, path string, action *api.LinkRelation) *customAdapterTransferRequest {
	return &customAdapterTransferRequest{"upload", oid, size, path, action}
}
func NewCustomAdapterDownloadRequest(oid string, size int64, action *api.LinkRelation) *customAdapterTransferRequest {
	return &customAdapterTransferRequest{"download", oid, size, "", action}
}

type customAdapterTerminateRequest struct {
	MessageType string `json:"type"`
}

func NewCustomAdapterTerminateRequest() *customAdapterTerminateRequest {
	return &customAdapterTerminateRequest{"terminate"}
}

// A common struct that allows all types of response to be identified
type customAdapterResponseMessage struct {
	Event          string           `json:"event"`
	Error          *api.ObjectError `json:"error"`
	Oid            string           `json:"oid"`
	Path           string           `json:"path,omitempty"` // always blank for upload
	BytesSoFar     int64            `json:"bytesSoFar"`
	BytesSinceLast int              `json:"bytesSinceLast"`
}

func (a *customAdapter) Begin(maxConcurrency int, cb TransferProgressCallback, completion chan TransferResult) error {
	// If config says not to launch multiple processes, downgrade incoming value
	useConcurrency := maxConcurrency
	if !a.concurrent {
		useConcurrency = 1
	}
	a.originalConcurrency = maxConcurrency

	tracerx.Printf("xfer: Custom transfer adapter %q using concurrency %d", a.name, useConcurrency)

	// Use common workers impl, but downgrade workers to number of processes
	return a.adapterBase.Begin(useConcurrency, cb, completion)
}

func (a *customAdapter) ClearTempStorage() error {
	// no action requred
	return nil
}

func (a *customAdapter) WorkerStarting(workerNum int) (interface{}, error) {

	// Start a process per worker
	// If concurrent = false we have already dialled back workers to 1
	tracerx.Printf("xfer: starting up custom transfer process %q for worker %d", a.name, workerNum)
	cmd := subprocess.ExecCommand(a.path, a.args)
	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to get stdout for custom transfer command %q remote: %v", a.path, err)
	}
	inp, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to get stdin for custom transfer command %q remote: %v", a.path, err)
	}
	// Capture stderr to trace
	tracer := &traceWriter{}
	tracer.processName = filepath.Base(a.path)
	cmd.Stderr = tracer
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("Failed to start custom transfer command %q remote: %v", a.path, err)
	}
	// Set up buffered reader/writer since we operate on lines
	ctx := &customAdapterWorkerContext{workerNum, cmd, outp, bufio.NewReader(outp), inp, tracer}

	// send initiate message
	initReq := NewCustomAdapterInitRequest(a.getOperationName(), a.concurrent, a.originalConcurrency)
	resp, err := a.exchangeMessage(ctx, initReq)
	if err != nil {
		a.abortWorkerProcess(ctx)
		return nil, err
	}
	if resp.Error != nil {
		a.abortWorkerProcess(ctx)
		return nil, fmt.Errorf("Error initializing custom adapter %q worker %d: %v", a.name, workerNum, resp.Error)
	}

	tracerx.Printf("xfer: started custom adapter process %q for worker %d OK", a.path, workerNum)

	// Save this process context and use in future callbacks
	return ctx, nil
}

func (a *customAdapter) getOperationName() string {
	if a.direction == Download {
		return "download"
	}
	return "upload"
}

// sendMessage sends a JSON message to the custom adapter process
func (a *customAdapter) sendMessage(ctx *customAdapterWorkerContext, req interface{}) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	tracerx.Printf("xfer: Custom adapter worker %d sending message: %v", ctx.workerNum, string(b))
	// Line oriented JSON
	b = append(b, '\n')
	_, err = ctx.stdin.Write(b)
	return err
}

func (a *customAdapter) readResponse(ctx *customAdapterWorkerContext) (*customAdapterResponseMessage, error) {
	line, err := ctx.bufferedOut.ReadString('\n')
	if err != nil {
		return nil, err
	}
	tracerx.Printf("xfer: Custom adapter worker %d received response: %v", ctx.workerNum, strings.TrimSpace(line))
	resp := &customAdapterResponseMessage{}
	err = json.Unmarshal([]byte(line), resp)
	return resp, err
}

// exchangeMessage sends a message to a process and reads a response if resp != nil
// Only fatal errors to communicate return an error, errors may be embedded in reply
func (a *customAdapter) exchangeMessage(ctx *customAdapterWorkerContext, req interface{}) (*customAdapterResponseMessage, error) {

	err := a.sendMessage(ctx, req)
	if err != nil {
		return nil, err
	}
	return a.readResponse(ctx)
}

// shutdownWorkerProcess terminates gracefully a custom adapter process
// returns an error if it couldn't shut down gracefully (caller may abortWorkerProcess)
func (a *customAdapter) shutdownWorkerProcess(ctx *customAdapterWorkerContext) error {
	defer ctx.errTracer.Flush()

	tracerx.Printf("xfer: Shutting down adapter worker %d", ctx.workerNum)

	finishChan := make(chan error, 1)
	go func() {
		termReq := NewCustomAdapterTerminateRequest()
		err := a.sendMessage(ctx, termReq)
		if err != nil {
			finishChan <- err
		}
		ctx.stdin.Close()
		ctx.stdout.Close()
		finishChan <- ctx.cmd.Wait()
	}()
	select {
	case err := <-finishChan:
		return err
	case <-time.After(30 * time.Second):
		return fmt.Errorf("Timeout while shutting down worker process %d", ctx.workerNum)
	}
}

// abortWorkerProcess terminates & aborts untidily, most probably breakdown of comms or internal error
func (a *customAdapter) abortWorkerProcess(ctx *customAdapterWorkerContext) {
	tracerx.Printf("xfer: Aborting worker process: %d", ctx.workerNum)
	ctx.stdin.Close()
	ctx.stdout.Close()
	ctx.cmd.Process.Kill()
}
func (a *customAdapter) WorkerEnding(workerNum int, ctx interface{}) {
	customCtx, ok := ctx.(*customAdapterWorkerContext)
	if !ok {
		tracerx.Printf("Context object for custom transfer %q was of the wrong type", a.name)
		return
	}

	err := a.shutdownWorkerProcess(customCtx)
	if err != nil {
		tracerx.Printf("xfer: error finishing up custom transfer process %q worker %d, aborting: %v", a.path, customCtx.workerNum, err)
		a.abortWorkerProcess(customCtx)
	}
}

func (a *customAdapter) DoTransfer(ctx interface{}, t *Transfer, cb TransferProgressCallback, authOkFunc func()) error {
	if ctx == nil {
		return fmt.Errorf("Custom transfer %q was not properly initialized, see previous errors", a.name)
	}

	customCtx, ok := ctx.(*customAdapterWorkerContext)
	if !ok {
		return fmt.Errorf("Context object for custom transfer %q was of the wrong type", a.name)
	}
	var authCalled bool

	rel, ok := t.Object.Rel(a.getOperationName())
	if !ok {
		return errors.New("Object not found on the server.")
	}
	var req *customAdapterTransferRequest
	if a.direction == Upload {
		req = NewCustomAdapterUploadRequest(t.Object.Oid, t.Object.Size, t.Path, rel)
	} else {
		req = NewCustomAdapterDownloadRequest(t.Object.Oid, t.Object.Size, rel)
	}
	err := a.sendMessage(customCtx, req)
	if err != nil {
		return err
	}

	// 1..N replies (including progress & one of download / upload)
	var complete bool
	for !complete {
		resp, err := a.readResponse(customCtx)
		if err != nil {
			return err
		}
		var wasAuthOk bool
		switch resp.Event {
		case "progress":
			// Progress
			if resp.Oid != t.Object.Oid {
				return fmt.Errorf("Unexpected oid %q in response, expecting %q", resp.Oid, t.Object.Oid)
			}
			if cb != nil {
				cb(t.Name, t.Object.Size, resp.BytesSoFar, resp.BytesSinceLast)
			}
			wasAuthOk = resp.BytesSoFar > 0
		case "complete":
			// Download/Upload complete
			if resp.Oid != t.Object.Oid {
				return fmt.Errorf("Unexpected oid %q in response, expecting %q", resp.Oid, t.Object.Oid)
			}
			if resp.Error != nil {
				return fmt.Errorf("Error transferring %q: %v", t.Object.Oid, resp.Error)
			}
			if a.direction == Download {
				// So we don't have to blindly trust external providers, check SHA
				if err = tools.VerifyFileHash(t.Object.Oid, resp.Path); err != nil {
					return fmt.Errorf("Downloaded file failed checks: %v", err)
				}
				// Move file to final location
				if err = tools.RenameFileCopyPermissions(resp.Path, t.Path); err != nil {
					return fmt.Errorf("Failed to copy downloaded file: %v", err)
				}
			} else if a.direction == Upload {
				if err = api.VerifyUpload(config.Config, t.Object); err != nil {
					return err
				}
			}
			wasAuthOk = true
			complete = true
		default:
			return fmt.Errorf("Invalid message %q from custom adapter %q", resp.Event, a.name)
		}
		// Fall through from both progress and completion messages
		// Call auth on first progress or success to free up other workers
		if wasAuthOk && authOkFunc != nil && !authCalled {
			authOkFunc()
			authCalled = true
		}
	}

	return nil
}

func newCustomAdapter(name string, dir Direction, path, args string, concurrent bool) *customAdapter {
	c := &customAdapter{newAdapterBase(name, dir, nil), path, args, concurrent, 3}
	// self implements impl
	c.transferImpl = c
	return c
}

// Initialise custom adapters based on current config
func configureCustomAdapters(cfg *config.Configuration, m *Manifest) {
	pathRegex := regexp.MustCompile(`lfs.customtransfer.([^.]+).path`)
	for k, v := range cfg.Git.All() {
		match := pathRegex.FindStringSubmatch(k)
		if match == nil {
			continue
		}

		name := match[1]
		path := v
		// retrieve other values
		args, _ := cfg.Git.Get(fmt.Sprintf("lfs.customtransfer.%s.args", name))
		concurrent := cfg.Git.Bool(fmt.Sprintf("lfs.customtransfer.%s.concurrent", name), true)
		direction, _ := cfg.Git.Get(fmt.Sprintf("lfs.customtransfer.%s.direction", name))
		if len(direction) == 0 {
			direction = "both"
		} else {
			direction = strings.ToLower(direction)
		}

		// Separate closure for each since we need to capture vars above
		newfunc := func(name string, dir Direction) TransferAdapter {
			return newCustomAdapter(name, dir, path, args, concurrent)
		}

		if direction == "download" || direction == "both" {
			m.RegisterNewTransferAdapterFunc(name, Download, newfunc)
		}
		if direction == "upload" || direction == "both" {
			m.RegisterNewTransferAdapterFunc(name, Upload, newfunc)
		}
	}
}
