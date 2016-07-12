package transfer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/github/git-lfs/localstorage"
	"github.com/github/git-lfs/tools"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/subprocess"
	"github.com/rubyist/tracerx"

	"github.com/github/git-lfs/config"
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
	buf bytes.Buffer
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
			tracerx.Printf("xfer_custom_stderr: %v", strings.TrimSpace(s))
		}
	}
}

type customAdapterWorkerContext struct {
	cmd         *exec.Cmd
	stdout      io.ReadCloser
	bufferedOut *bufio.Reader
	stdin       io.WriteCloser
	errTracer   *traceWriter
}

type customAdapterInitRequest struct {
	Operation           string `json:"operation"`
	Concurrent          bool   `json:"concurrent"`
	ConcurrentTransfers int    `json:"concurrenttransfers"`
}
type customAdapterInitResponse struct {
	Error *api.ObjectError `json:"error,omitempty"`
}
type customAdapterTransferRequest struct { // common between upload/download
	Oid    string            `json:"oid"`
	Size   int64             `json:"size"`
	Path   string            `json:"path,omitempty"`
	Action *api.LinkRelation `json:"action"`
}
type customAdapterTransferResponse struct { // common between upload/download
	Oid   string           `json:"oid"`
	Path  string           `json:"path,omitempty"` // always blank for upload
	Error *api.ObjectError `json:"error,omitempty"`
}
type customAdapterTerminateRequest struct {
	Complete bool `json:"complete"`
}
type customAdapterProgressResponse struct {
	Oid            string `json:"oid"`
	BytesSoFar     int64  `json:"bytesSoFar"`
	BytesSinceLast int    `json:"bytesSinceLast"`
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
	cmd.Stderr = tracer
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("Failed to start custom transfer command %q remote: %v", a.path, err)
	}
	// Set up buffered reader/writer since we operate on lines
	ctx := &customAdapterWorkerContext{cmd, outp, bufio.NewReader(outp), inp, tracer}

	// send initiate message
	initReq := &customAdapterInitRequest{a.getOperationName(), a.concurrent, a.originalConcurrency}
	var initResp customAdapterInitResponse
	err = a.exchangeMessage(ctx, initReq, &initResp)
	if err != nil {
		a.abortWorkerProcess(ctx)
		return nil, err
	}

	tracerx.Printf("xfer: %q for worker %d started OK", a.name, workerNum)

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
	// Line oriented JSON
	b = append(b, '\n')
	_, err = ctx.stdin.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// readResponse reads one of 1..N possible responses and populates the first one which
// was unmarshalled correctly. This allows us to listen for one of possibly many responses
// Returns the index of the item which was populated
func (a *customAdapter) readResponse(ctx *customAdapterWorkerContext, possResps []interface{}) (int, error) {
	line, err := ctx.bufferedOut.ReadString('\n')
	if err != nil {
		return 0, err
	}
	for i, resp := range possResps {
		if json.Unmarshal([]byte(line), resp) == nil {
			return i, nil
		}
	}
	return 0, fmt.Errorf("Response %q did not match any of possible responses %v", string(line), possResps)

}

// exchangeMessage sends a message to a process and reads a response if resp != nil
// Only fatal errors to communicate return an error, errors may be embedded in reply
func (a *customAdapter) exchangeMessage(ctx *customAdapterWorkerContext, req, resp interface{}) error {

	err := a.sendMessage(ctx, req)
	if err != nil {
		return err
	}
	// Read response if needed
	if resp != nil {
		_, err = a.readResponse(ctx, []interface{}{resp})
		return err
	}
	return nil
}

// shutdownWorkerProcess terminates gracefully a custom adapter process
// returns an error if it couldn't shut down gracefully (caller may abortWorkerProcess)
func (a *customAdapter) shutdownWorkerProcess(ctx *customAdapterWorkerContext) error {
	defer ctx.errTracer.Flush()

	termReq := &customAdapterTerminateRequest{true}
	err := a.exchangeMessage(ctx, termReq, nil)
	if err != nil {
		return err
	}
	ctx.stdin.Close()
	ctx.stdout.Close()
	return ctx.cmd.Wait()
}

// abortWorkerProcess terminates & aborts untidily, most probably breakdown of comms or internal error
func (a *customAdapter) abortWorkerProcess(ctx *customAdapterWorkerContext) {
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
		tracerx.Printf("xfer: error finishing up custom transfer process %q, aborting: %v", a.name, err)
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
	req := &customAdapterTransferRequest{t.Object.Oid, t.Object.Size, "", rel}
	if a.direction == Upload {
		req.Path = localstorage.Objects().ObjectPath(t.Object.Oid)
	}
	err := a.sendMessage(customCtx, req)
	if err != nil {
		return err
	}

	// 1..N replies (including progress & one of download / upload)
	possResps := []interface{}{&customAdapterProgressResponse{}, &customAdapterTransferResponse{}}
	var complete bool
	for !complete {
		respIdx, err := a.readResponse(customCtx, possResps)
		if err != nil {
			return err
		}
		var wasAuthOk bool
		switch respIdx {
		case 0:
			// Progress
			prog := possResps[respIdx].(customAdapterProgressResponse)
			if prog.Oid != t.Object.Oid {
				return fmt.Errorf("Unexpected oid %q in response, expecting %q", prog.Oid, t.Object.Oid)
			}
			if cb != nil {
				cb(t.Name, t.Object.Size, prog.BytesSoFar, prog.BytesSinceLast)
			}
			wasAuthOk = prog.BytesSoFar > 0
		case 1:
			// Download/Upload complete
			comp := possResps[respIdx].(customAdapterTransferResponse)
			if comp.Oid != t.Object.Oid {
				return fmt.Errorf("Unexpected oid %q in response, expecting %q", comp.Oid, t.Object.Oid)
			}
			if comp.Error != nil {
				return fmt.Errorf("Error transferring %q: %v", t.Object.Oid, comp.Error.Error())
			}
			if a.direction == Download {
				// So we don't have to blindly trust external providers, check SHA
				if err = tools.VerifyFileHash(t.Object.Oid, comp.Path); err != nil {
					return fmt.Errorf("Downloaded file failed checks: %v", err)
				}
				// Move file to final location
				if err = tools.RenameFileCopyPermissions(comp.Path, t.Path); err != nil {
					return fmt.Errorf("Failed to copy downloaded file: %v", err)
				}
			} else if a.direction == Upload {
				if err = api.VerifyUpload(t.Object); err != nil {
					return err
				}
			}
			wasAuthOk = true
			complete = true
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
func ConfigureCustomAdapters() {
	pathRegex := regexp.MustCompile(`lfs.customtransfer.([^.]+).path`)
	for k, v := range config.Config.AllGitConfig() {
		if match := pathRegex.FindStringSubmatch(k); match != nil {
			name := match[1]
			path := v
			var args string
			var concurrent bool
			var direction string
			// retrieve other values
			args, _ = config.Config.GitConfig(fmt.Sprintf("lfs.customtransfer.%s.args", name))
			concurrent = config.Config.GitConfigBool(fmt.Sprintf("lfs.customtransfer.%s.concurrent", name), true)
			direction, _ = config.Config.GitConfig(fmt.Sprintf("lfs.customtransfer.%s.direction", name))
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
				RegisterNewTransferAdapterFunc(name, Download, newfunc)
			}
			if direction == "upload" || direction == "both" {
				RegisterNewTransferAdapterFunc(name, Upload, newfunc)
			}

		}
	}

}

func init() {
	ConfigureCustomAdapters()
}
