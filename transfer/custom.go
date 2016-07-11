package transfer

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

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

type customAdapterWorkerContext struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stdin  io.WriteCloser
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
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("Failed to start custom transfer command %q remote: %v", a.path, err)
	}

	// TODO send initiate message

	tracerx.Printf("xfer: %q for worker %d started OK", a.name, workerNum)

	// Save this process context and use in future callbacks
	return &customAdapterWorkerContext{cmd, outp, inp}, nil
}
func (a *customAdapter) WorkerEnding(workerNum int, ctx interface{}) {
	customCtx, ok := ctx.(*customAdapterWorkerContext)
	if !ok {
		tracerx.Printf("Context object for custom transfer %q was of the wrong type", a.name)
		return
	}

	// TODO send finish message

	err := customCtx.cmd.Wait()
	if err != nil {
		tracerx.Printf("xfer: error finishing up custom transfer process %q: %v", a.name, err)
	}
}

func (a *customAdapter) DoTransfer(ctx interface{}, t *Transfer, cb TransferProgressCallback, authOkFunc func()) error {
	if ctx == nil {
		return fmt.Errorf("Custom transfer %q was not properly initialized, see previous errors", a.name)
	}
	// TODO
	// customCtx, ok := ctx.(*customAdapterWorkerContext)
	// if !ok {
	// 	return fmt.Errorf("Context object for custom transfer %q was of the wrong type", a.name)
	// }

	// TODO call authOK on first non-zero progress
	if authOkFunc != nil {
		authOkFunc()
	}

	// TODO send transfer request, receive progress and completion
	if cb != nil {
		advanceCallbackProgress(cb, t, t.Object.Size)
	}

	if a.direction == Upload {
		return api.VerifyUpload(t.Object)
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
