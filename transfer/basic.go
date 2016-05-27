package transfer

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/httputil"
	"github.com/github/git-lfs/progress"
	"github.com/github/git-lfs/tools"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

const (
	BasicAdapterName = "basic"
)

// Base implementation of basic all-or-nothing HTTP upload / download adapter
type basicAdapter struct {
	direction      Direction
	jobChan        chan *Transfer
	cb             TransferProgressCallback
	outChan        chan TransferResult
	maxConcurrency int
	// WaitGroup to sync the completion of all workers
	workerWait sync.WaitGroup
	// WaitGroup to serialise the first transfer response to perform login if needed
	authWait sync.WaitGroup
}

func newBasicAdapter(d Direction) *basicAdapter {
	return &basicAdapter{
		direction:      d,
		maxConcurrency: config.Config.ConcurrentTransfers(),
	}
}

func (a *basicAdapter) Direction() Direction {
	return a.direction
}

func (a *basicAdapter) Name() string {
	return BasicAdapterName
}

func (a *basicAdapter) Begin(maxConcurrency int, cb TransferProgressCallback, completion chan TransferResult) error {
	a.cb = cb
	a.outChan = completion
	a.jobChan = make(chan *Transfer, 100)

	tracerx.Printf("xfer: adapter %q Begin()", a.Name())

	a.workerWait.Add(maxConcurrency)
	a.authWait.Add(1)
	for i := 0; i < maxConcurrency; i++ {
		go a.worker(i)
	}
	return nil
}

func (a *basicAdapter) Add(t *Transfer) {
	tracerx.Printf("xfer: adapter %q Add() for %q", a.Name(), t.Object.Oid)
	a.jobChan <- t
}

func (a *basicAdapter) End() {
	tracerx.Printf("xfer: adapter %q End()", a.Name())
	close(a.jobChan)
	// wait for all transfers to complete
	a.workerWait.Wait()
}

func (a *basicAdapter) ClearTempStorage() error {
	// Should be empty already but also remove dir
	return os.RemoveAll(a.tempDir())
}

// worker function, many of these run per adapter
func (a *basicAdapter) worker(workerNum int) {

	tracerx.Printf("xfer: adapter %q worker %d starting", a.Name(), workerNum)
	waitForAuth := workerNum > 0
	signalAuthOnResponse := workerNum == 0

	// First worker is the only one allowed to start immediately
	// The rest wait until successful response from 1st worker to
	// make sure only 1 login prompt is presented if necessary
	// Deliberately outside jobChan processing so we know worker 0 will process 1st item
	if waitForAuth {
		tracerx.Printf("xfer: adapter %q worker %d waiting for Auth", a.Name(), workerNum)
		a.authWait.Wait()
		tracerx.Printf("xfer: adapter %q worker %d auth signal received", a.Name(), workerNum)
	}

	for t := range a.jobChan {
		tracerx.Printf("xfer: adapter %q worker %d processing job for %q", a.Name(), workerNum, t.Object.Oid)
		var err error
		switch a.Direction() {
		case Download:
			err = a.download(t, signalAuthOnResponse)
		case Upload:
			err = a.upload(t, signalAuthOnResponse)
		}

		if a.outChan != nil {
			res := TransferResult{t, err}
			a.outChan <- res
		}

		// Only need to signal for auth once
		signalAuthOnResponse = false

		tracerx.Printf("xfer: adapter %q worker %d finished job for %q", a.Name(), workerNum, t.Object.Oid)
	}
	// This will only happen if no jobs were submitted; just wake up all workers to finish
	if signalAuthOnResponse {
		a.authWait.Done()
	}
	a.workerWait.Done()
}

func (a *basicAdapter) tempDir() string {
	// Must be dedicated to this adapter as deleted by ClearTempStorage
	d := filepath.Join(os.TempDir(), "git-lfs-basic-temp")
	if err := os.MkdirAll(d, 0755); err != nil {
		return os.TempDir()
	}
	return d
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
	defer res.Body.Close()

	// Signal auth OK on success response, before starting download to free up
	// other workers immediately
	if signalAuthOnResponse {
		a.authWait.Done()
	}

	// Now do transfer of content
	f, err := ioutil.TempFile(a.tempDir(), t.Object.Oid+"-")
	if err != nil {
		return fmt.Errorf("cannot create temp file: %v", err)
	}

	defer func() {
		if err != nil {
			// Don't leave the temp file lying around on error.
			_ = os.Remove(f.Name()) // yes, ignore the error, not much we can do about it.
		}
	}()

	hasher := tools.NewHashingReader(res.Body)

	// ensure we always close f. Note that this does not conflict with  the
	// close below, as close is idempotent.
	defer f.Close()
	tempfilename := f.Name()
	// Wrap callback to give name context
	ccb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		if a.cb != nil {
			return a.cb(t.Name, totalSize, readSoFar, readSinceLast)
		}
		return nil
	}
	written, err := tools.CopyWithCallback(f, hasher, res.ContentLength, ccb)
	if err != nil {
		return fmt.Errorf("cannot write data to tempfile %q: %v", tempfilename, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("can't close tempfile %q: %v", tempfilename, err)
	}

	if actual := hasher.Hash(); actual != t.Object.Oid {
		return fmt.Errorf("Expected OID %s, got %s after %d bytes written", t.Object.Oid, actual, written)
	}

	return tools.RenameFileCopyPermissions(tempfilename, t.Path)

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
	// Wrap callback to give name context
	ccb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		if a.cb != nil {
			return a.cb(t.Name, totalSize, readSoFar, readSinceLast)
		}
		return nil
	}
	var reader io.Reader
	reader = &progress.CallbackReader{
		C:         ccb,
		TotalSize: t.Object.Size,
		Reader:    f,
	}

	if signalAuthOnResponse {
		// Signal auth was ok on first read; this frees up other workers to start
		reader = newStartCallbackReader(reader, func(*startCallbackReader) {
			a.authWait.Done()
		})
	}

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

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return api.VerifyUpload(t.Object)
}

// startCallbackReader is a reader wrapper which calls a function as soon as the
// first Read() call is made. This callback is only made once
type startCallbackReader struct {
	r      io.Reader
	cb     func(*startCallbackReader)
	cbDone bool
}

func (s *startCallbackReader) Read(p []byte) (n int, err error) {
	if !s.cbDone && s.cb != nil {
		s.cb(s)
		s.cbDone = true
	}
	return s.r.Read(p)
}
func newStartCallbackReader(r io.Reader, cb func(*startCallbackReader)) *startCallbackReader {
	return &startCallbackReader{r, cb, false}
}

func init() {
	ul := newBasicAdapter(Upload)
	RegisterAdapter(ul)
	dl := newBasicAdapter(Download)
	RegisterAdapter(dl)
}
