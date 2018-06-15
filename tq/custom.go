package tq

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/fs"
	"github.com/git-lfs/git-lfs/tools"

	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/rubyist/tracerx"
)

// Adapter for custom transfer via external process
type customAdapter struct {
	*adapterBase
	path                string
	args                string
	concurrent          bool
	originalConcurrency int
	standalone          bool
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
	cmd         *subprocess.Cmd
	stdout      io.ReadCloser
	bufferedOut *bufio.Reader
	stdin       io.WriteCloser
	errTracer   *traceWriter
}

type customAdapterInitRequest struct {
	Event               string `json:"event"`
	Operation           string `json:"operation"`
	Remote              string `json:"remote"`
	Concurrent          bool   `json:"concurrent"`
	ConcurrentTransfers int    `json:"concurrenttransfers"`
}

func NewCustomAdapterInitRequest(
	op string, remote string, concurrent bool, concurrentTransfers int,
) *customAdapterInitRequest {
	return &customAdapterInitRequest{"init", op, remote, concurrent, concurrentTransfers}
}

type customAdapterTransferRequest struct {
	// common between upload/download
	Event  string  `json:"event"`
	Oid    string  `json:"oid"`
	Size   int64   `json:"size"`
	Path   string  `json:"path,omitempty"`
	Action *Action `json:"action"`
}

func NewCustomAdapterUploadRequest(oid string, size int64, path string, action *Action) *customAdapterTransferRequest {
	return &customAdapterTransferRequest{"upload", oid, size, path, action}
}
func NewCustomAdapterDownloadRequest(oid string, size int64, action *Action) *customAdapterTransferRequest {
	return &customAdapterTransferRequest{"download", oid, size, "", action}
}

type customAdapterTerminateRequest struct {
	Event string `json:"event"`
}

func NewCustomAdapterTerminateRequest() *customAdapterTerminateRequest {
	return &customAdapterTerminateRequest{"terminate"}
}

// A common struct that allows all types of response to be identified
type customAdapterResponseMessage struct {
	Event          string       `json:"event"`
	Error          *ObjectError `json:"error"`
	Oid            string       `json:"oid"`
	Path           string       `json:"path,omitempty"` // always blank for upload
	BytesSoFar     int64        `json:"bytesSoFar"`
	BytesSinceLast int          `json:"bytesSinceLast"`
}

func (a *customAdapter) Begin(cfg AdapterConfig, cb ProgressCallback) error {
	a.originalConcurrency = cfg.ConcurrentTransfers()
	if a.concurrent {
		// Use common workers impl, but downgrade workers to number of processes
		return a.adapterBase.Begin(cfg, cb)
	}

	// If config says not to launch multiple processes, downgrade incoming value
	return a.adapterBase.Begin(&customAdapterConfig{AdapterConfig: cfg}, cb)
}

func (a *customAdapter) ClearTempStorage() error {
	// no action requred
	return nil
}

func (a *customAdapter) WorkerStarting(workerNum int) (interface{}, error) {
	// Start a process per worker
	// If concurrent = false we have already dialled back workers to 1
	a.Trace("xfer: starting up custom transfer process %q for worker %d", a.name, workerNum)
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
	initReq := NewCustomAdapterInitRequest(
		a.getOperationName(), a.remote, a.concurrent, a.originalConcurrency,
	)
	resp, err := a.exchangeMessage(ctx, initReq)
	if err != nil {
		a.abortWorkerProcess(ctx)
		return nil, err
	}
	if resp.Error != nil {
		a.abortWorkerProcess(ctx)
		return nil, fmt.Errorf("Error initializing custom adapter %q worker %d: %v", a.name, workerNum, resp.Error)
	}

	a.Trace("xfer: started custom adapter process %q for worker %d OK", a.path, workerNum)

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
	a.Trace("xfer: Custom adapter worker %d sending message: %v", ctx.workerNum, string(b))
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
	a.Trace("xfer: Custom adapter worker %d received response: %v", ctx.workerNum, strings.TrimSpace(line))
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

	a.Trace("xfer: Shutting down adapter worker %d", ctx.workerNum)

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
	a.Trace("xfer: Aborting worker process: %d", ctx.workerNum)
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

func (a *customAdapter) DoTransfer(ctx interface{}, t *Transfer, cb ProgressCallback, authOkFunc func()) error {
	if ctx == nil {
		return fmt.Errorf("Custom transfer %q was not properly initialized, see previous errors", a.name)
	}

	customCtx, ok := ctx.(*customAdapterWorkerContext)
	if !ok {
		return fmt.Errorf("Context object for custom transfer %q was of the wrong type", a.name)
	}
	var authCalled bool

	rel, err := t.Rel(a.getOperationName())
	if err != nil {
		return err
	}
	if rel == nil && !a.standalone {
		return errors.Errorf("Object %s not found on the server.", t.Oid)
	}
	var req *customAdapterTransferRequest
	if a.direction == Upload {
		req = NewCustomAdapterUploadRequest(t.Oid, t.Size, t.Path, rel)
	} else {
		req = NewCustomAdapterDownloadRequest(t.Oid, t.Size, rel)
	}
	if err = a.sendMessage(customCtx, req); err != nil {
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
			if resp.Oid != t.Oid {
				return fmt.Errorf("Unexpected oid %q in response, expecting %q", resp.Oid, t.Oid)
			}
			if cb != nil {
				cb(t.Name, t.Size, resp.BytesSoFar, resp.BytesSinceLast)
			}
			wasAuthOk = resp.BytesSoFar > 0
		case "complete":
			// Download/Upload complete
			if resp.Oid != t.Oid {
				return fmt.Errorf("Unexpected oid %q in response, expecting %q", resp.Oid, t.Oid)
			}
			if resp.Error != nil {
				return fmt.Errorf("Error transferring %q: %v", t.Oid, resp.Error)
			}
			if a.direction == Download {
				// So we don't have to blindly trust external providers, check SHA
				if err = tools.VerifyFileHash(t.Oid, resp.Path); err != nil {
					return fmt.Errorf("Downloaded file failed checks: %v", err)
				}
				// Move file to final location
				if err = tools.RenameFileCopyPermissions(resp.Path, t.Path); err != nil {
					return fmt.Errorf("Failed to copy downloaded file: %v", err)
				}
			} else if a.direction == Upload {
				if err = verifyUpload(a.apiClient, a.remote, t); err != nil {
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

func newCustomAdapter(f *fs.Filesystem, name string, dir Direction, path, args string, concurrent, standalone bool) *customAdapter {
	c := &customAdapter{newAdapterBase(f, name, dir, nil), path, args, concurrent, 3, standalone}
	// self implements impl
	c.transferImpl = c
	return c
}

// Initialise custom adapters based on current config
func configureCustomAdapters(git Env, m *Manifest) {
	pathRegex := regexp.MustCompile(`lfs.customtransfer.([^.]+).path`)
	for k, _ := range git.All() {
		match := pathRegex.FindStringSubmatch(k)
		if match == nil {
			continue
		}

		name := match[1]
		path, _ := git.Get(k)
		// retrieve other values
		args, _ := git.Get(fmt.Sprintf("lfs.customtransfer.%s.args", name))
		concurrent := git.Bool(fmt.Sprintf("lfs.customtransfer.%s.concurrent", name), true)
		direction, _ := git.Get(fmt.Sprintf("lfs.customtransfer.%s.direction", name))
		if len(direction) == 0 {
			direction = "both"
		} else {
			direction = strings.ToLower(direction)
		}

		// Separate closure for each since we need to capture vars above
		newfunc := func(name string, dir Direction) Adapter {
			standalone := m.standaloneTransferAgent != ""
			return newCustomAdapter(m.fs, name, dir, path, args, concurrent, standalone)
		}

		if direction == "download" || direction == "both" {
			m.RegisterNewAdapterFunc(name, Download, newfunc)
		}
		if direction == "upload" || direction == "both" {
			m.RegisterNewAdapterFunc(name, Upload, newfunc)
		}
	}
}

type customAdapterConfig struct {
	AdapterConfig
}

func (c *customAdapterConfig) ConcurrentTransfers() int {
	return 1
}
