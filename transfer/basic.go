package transfer

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/httputil"

	"github.com/github/git-lfs/progress"
)

// Base implementation of basic all-or-nothing HTTP upload / download adapter
type basicAdapter struct {
	direction Direction
	jobChan   chan *Transfer
	cb        progress.CopyCallback
	outChan   chan TransferResult
	// WaitGroup to sync the completion of all workers
	workerWait sync.WaitGroup
	// WaitGroup to serialise the first transfer response to perform login if needed
	authWait sync.WaitGroup
}

func newBasicAdapter(d Direction) *basicAdapter {
	return &basicAdapter{
		direction: d,
		jobChan:   make(chan *Transfer, 100),
	}
}

func (a *basicAdapter) Direction() Direction {
	return a.direction
}

func (a *basicAdapter) Name() string {
	return "basic"
}

func (a *basicAdapter) Begin(cb progress.CopyCallback, completion chan TransferResult) error {
	a.cb = cb
	a.outChan = completion

	numworkers := config.Config.ConcurrentTransfers()
	a.workerWait.Add(numworkers)
	a.authWait.Add(1)
	for i := 0; i < numworkers; i++ {
		go a.worker(i)
	}
	return nil
}

func (a *basicAdapter) Add(t *Transfer) {
	a.jobChan <- t
}

func (a *basicAdapter) End() {
	// wait for all transfers to complete
	a.workerWait.Wait()
}

func (a *basicAdapter) ClearTempStorage() error {
	// TODO @sinbad
	return nil
}

// worker function, many of these run per adapter
func (a *basicAdapter) worker(workerNum int) {

	isFirstWorker := workerNum == 0
	signalAuthOnResponse := isFirstWorker

	for t := range a.jobChan {
		if !isFirstWorker {
			// First worker is the only one allowed to start immediately
			// The rest wait until successful response from 1st worker to
			// make sure only 1 login prompt is presented if necessary
			a.authWait.Wait()
		}
		var err error
		switch a.Direction() {
		case Download:
			err = a.download(t, signalAuthOnResponse)
		case Upload:
			err = a.upload(t, signalAuthOnResponse)
		}

		res := TransferResult{t, err}
		a.outChan <- res

		signalAuthOnResponse = false
	}
	a.workerWait.Done()
}

func (a *basicAdapter) download(t *Transfer, signalAuthOnResponse bool) error {
	rel, ok := t.Object.Rel("download")
	if !ok {
		return errors.New("Object not found on the server.")
	}

	req, err := httputil.NewHttpRequest("GET", rel.Href, rel.Header)
	if err != nil {
		return err
	}

	res, err := httputil.DoHttpRequest(req, true)
	if err != nil {
		return errutil.NewRetriableError(err)
	}
	httputil.LogTransfer("lfs.data.download", res)

	// Signal auth OK on success response, before starting download to free up
	// other workers immediately
	if signalAuthOnResponse {
		a.authWait.Done()
	}

	// Now do transfer of content
	// TODO @sinbad - re-use bufferDownloadedFile?

	return nil

}
func (a *basicAdapter) upload(t *Transfer, signalAuthOnResponse bool) error {
	rel, ok := t.Object.Rel("upload")
	if !ok {
		return fmt.Errorf("No upload action for this object.")
	}

	req, err := httputil.NewHttpRequest("PUT", rel.Href, rel.Header)
	if err != nil {
		return err
	}

	if len(req.Header.Get("Content-Type")) == 0 {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	if req.Header.Get("Transfer-Encoding") == "chunked" {
		req.TransferEncoding = []string{"chunked"}
	} else {
		req.Header.Set("Content-Length", strconv.FormatInt(t.Object.Size, 10))
	}

	req.ContentLength = t.Object.Size

	f, err := os.OpenFile(t.Path, os.O_RDONLY, 0644)
	if err != nil {
		return errutil.Error(err)
	}
	defer f.Close()

	// Ensure progress callbacks made while uploading
	reader := &progress.CallbackReader{
		C:         a.cb,
		TotalSize: t.Object.Size,
		Reader:    f,
	}

	// TODO @sinbad - use extra custom wrapper to signalAuthOnResponse earlier
	req.Body = ioutil.NopCloser(reader)

	res, err := httputil.DoHttpRequest(req, true)
	if err != nil {
		return errutil.NewRetriableError(err)
	}
	httputil.LogTransfer("lfs.data.upload", res)

	// A status code of 403 likely means that an authentication token for the
	// upload has expired. This can be safely retried.
	if res.StatusCode == 403 {
		return errutil.NewRetriableError(err)
	}

	if res.StatusCode > 299 {
		return errutil.Errorf(nil, "Invalid status for %s: %d", httputil.TraceHttpReq(req), res.StatusCode)
	}

	// Signal auth OK on success response, before starting download to free up
	// other workers immediately
	// TODO @sinbad remove this when custom readcloser wrapper does it instead
	if signalAuthOnResponse {
		a.authWait.Done()
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return api.VerifyUpload(t.Object)
}

func init() {
	ul := newBasicAdapter(Upload)
	RegisterAdapter(ul)
	dl := newBasicAdapter(Download)
	RegisterAdapter(dl)
}
