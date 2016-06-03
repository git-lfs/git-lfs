package transfer

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/httputil"
	"github.com/github/git-lfs/progress"
	"github.com/github/git-lfs/tools"
)

const (
	BasicAdapterName = "basic"
)

// Base implementation of basic all-or-nothing HTTP upload / download adapter
type basicAdapter struct {
	*adapterBase
}

func (a *basicAdapter) ClearTempStorage() error {
	// Should be empty already but also remove dir
	return os.RemoveAll(a.tempDir())
}

func (a *basicAdapter) tempDir() string {
	// Must be dedicated to this adapter as deleted by ClearTempStorage
	d := filepath.Join(os.TempDir(), "git-lfs-basic-temp")
	if err := os.MkdirAll(d, 0755); err != nil {
		return os.TempDir()
	}
	return d
}

// Adapter for basic downloads
type basicDownloadAdapter struct {
	*basicAdapter
}

func (a *basicDownloadAdapter) DoTransfer(t *Transfer, cb TransferProgressCallback, authOkFunc func()) error {
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
	if authOkFunc != nil {
		authOkFunc()
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
		if cb != nil {
			return cb(t.Name, totalSize, readSoFar, readSinceLast)
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

// Adapter for basic uploads
type basicUploadAdapter struct {
	*basicAdapter
}

func (a *basicUploadAdapter) DoTransfer(t *Transfer, cb TransferProgressCallback, authOkFunc func()) error {
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
		if cb != nil {
			return cb(t.Name, totalSize, readSoFar, readSinceLast)
		}
		return nil
	}
	var reader io.Reader
	reader = &progress.CallbackReader{
		C:         ccb,
		TotalSize: t.Object.Size,
		Reader:    f,
	}

	// Signal auth was ok on first read; this frees up other workers to start
	if authOkFunc != nil {
		reader = newStartCallbackReader(reader, func(*startCallbackReader) {
			authOkFunc()
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

func newBasicAdapter(name string, dir Direction) *basicAdapter {
	return &basicAdapter{newAdapterBase(name, dir, nil)}
}

func init() {
	newfunc := func(name string, dir Direction) TransferAdapter {
		switch dir {
		case Download:
			bd := &basicDownloadAdapter{newBasicAdapter(name, dir)}
			// self implements impl
			bd.transferImpl = bd
			return bd
		case Upload:
			bu := &basicUploadAdapter{newBasicAdapter(name, dir)}
			// self implements impl
			bu.transferImpl = bu
			return bu
		}
		return nil
	}
	RegisterNewTransferAdapterFunc(BasicAdapterName, Upload, newfunc)
	RegisterNewTransferAdapterFunc(BasicAdapterName, Download, newfunc)
}
