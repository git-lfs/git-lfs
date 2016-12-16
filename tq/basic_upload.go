package tq

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/httputil"
	"github.com/git-lfs/git-lfs/progress"
)

const (
	BasicAdapterName = "basic"
)

// Adapter for basic uploads (non resumable)
type basicUploadAdapter struct {
	*adapterBase
}

func (a *basicUploadAdapter) ClearTempStorage() error {
	// Should be empty already but also remove dir
	return os.RemoveAll(a.tempDir())
}

func (a *basicUploadAdapter) tempDir() string {
	// Must be dedicated to this adapter as deleted by ClearTempStorage
	d := filepath.Join(os.TempDir(), "git-lfs-basic-temp")
	if err := os.MkdirAll(d, 0755); err != nil {
		return os.TempDir()
	}
	return d
}

func (a *basicUploadAdapter) WorkerStarting(workerNum int) (interface{}, error) {
	return nil, nil
}
func (a *basicUploadAdapter) WorkerEnding(workerNum int, ctx interface{}) {
}

func (a *basicUploadAdapter) DoTransfer(ctx interface{}, t *Transfer, cb ProgressCallback, authOkFunc func()) error {
	rel, err := t.Actions.Get("upload")
	if err != nil {
		return err
		// return fmt.Errorf("No upload action for this object.")
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
		req.Header.Set("Content-Length", strconv.FormatInt(t.Size, 10))
	}

	req.ContentLength = t.Size

	f, err := os.OpenFile(t.Path, os.O_RDONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "basic upload")
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
		TotalSize: t.Size,
		Reader:    f,
	}

	// Signal auth was ok on first read; this frees up other workers to start
	if authOkFunc != nil {
		reader = newStartCallbackReader(reader, func(*startCallbackReader) {
			authOkFunc()
		})
	}

	req.Body = ioutil.NopCloser(reader)

	res, err := httputil.DoHttpRequest(config.Config, req, !t.Authenticated)
	if err != nil {
		return errors.NewRetriableError(err)
	}
	httputil.LogTransfer(config.Config, "lfs.data.upload", res)

	// A status code of 403 likely means that an authentication token for the
	// upload has expired. This can be safely retried.
	if res.StatusCode == 403 {
		err = errors.New("http: received status 403")
		return errors.NewRetriableError(err)
	}

	if res.StatusCode > 299 {
		return errors.Wrapf(nil, "Invalid status for %s: %d", httputil.TraceHttpReq(req), res.StatusCode)
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return api.VerifyUpload(config.Config, toApiObject(t))
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

func configureBasicUploadAdapter(m *Manifest) {
	m.RegisterNewAdapterFunc(BasicAdapterName, Upload, func(name string, dir Direction) Adapter {
		switch dir {
		case Upload:
			bu := &basicUploadAdapter{newAdapterBase(name, dir, nil)}
			// self implements impl
			bu.transferImpl = bu
			return bu
		case Download:
			panic("Should never ask this func for basic download")
		}
		return nil
	})
}
