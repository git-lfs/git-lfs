package tq

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/tools"
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
	rel, err := t.Rel("upload")
	if err != nil {
		return err
	}
	if rel == nil {
		return errors.Errorf("No upload action for object: %s", t.Oid)
	}

	req, err := a.newHTTPRequest("PUT", rel)
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

	if err := a.fileHTTPUpload(req, t, 0, cb, authOkFunc); err != nil {
		return err
	}

	return verifyUpload(a.apiClient, a.remote, t)
}

func (a *adapterBase) fileHTTPUpload(req *http.Request, t *Transfer, offset int64, cb ProgressCallback, authOkFunc func()) error {
	req.ContentLength = t.Size - offset

	f, err := os.OpenFile(t.Path, os.O_RDONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "lfs upload")
	}

	defer f.Close()

	// Ensure progress callbacks made while uploading
	// Wrap callback to give name context
	if cb == nil {
		cb = ProgressCallback(func(n, o string, t int64, r int64, rSince int) error {
			return nil
		})
	}

	ccb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		return cb(t.Name, t.Oid, totalSize, readSoFar, readSinceLast)
	}

	cbr := tools.NewBodyWithCallback(f, t.Size, ccb)
	var reader lfsapi.ReadSeekCloser = cbr

	if authOkFunc != nil || offset > 0 {
		reader = newStartCallbackReader(reader, func() error {
			authOkFunc()

			if offset > 0 {
				// seek to the offset since lfsapi.Client rewinds the body
				if _, err := f.Seek(offset, os.SEEK_CUR); err != nil {
					return err
				}
			}

			return nil
		})
	}

	req.Body = reader
	req = a.apiClient.LogRequest(req, "lfs.data.upload")

	res, err := a.doHTTP(t, req)
	if err != nil {
		// We're about to return a retriable error, meaning that this
		// transfer will either be retried, or it will fail.
		//
		// Either way, let's decrement the number of bytes that we've
		// read _so far_, so that the next iteration doesn't re-transfer
		// those bytes, according to the progress meter.
		if perr := cbr.ResetProgress(); perr != nil {
			err = errors.Wrap(err, perr.Error())
		}

		return errors.NewRetriableError(err)
	}

	switch res.StatusCode {
	case 403, // likely an expired auth token
		503: // service unavailable
		return errors.NewRetriableError(fmt.Errorf("http: received status %d", res.StatusCode))
	}

	if res.StatusCode > 299 {
		return errors.Wrapf(nil, "Invalid status for %s %s: %d",
			req.Method,
			strings.SplitN(req.URL.String(), "?", 2)[0],
			res.StatusCode,
		)
	}

	io.Copy(ioutil.Discard, res.Body)
	return res.Body.Close()
}

// startCallbackReader is a reader wrapper which calls a function as soon as the
// first Read() call is made. This callback is only made once
type startCallbackReader struct {
	cb     func() error
	cbDone bool
	lfsapi.ReadSeekCloser
}

func (s *startCallbackReader) Read(p []byte) (n int, err error) {
	if !s.cbDone && s.cb != nil {
		if err := s.cb(); err != nil {
			return 0, err
		}
		s.cbDone = true
	}
	return s.ReadSeekCloser.Read(p)
}

func newStartCallbackReader(r lfsapi.ReadSeekCloser, cb func() error) *startCallbackReader {
	return &startCallbackReader{
		ReadSeekCloser: r,
		cb:             cb,
	}
}

func configureBasicUploadAdapter(m *Manifest) {
	m.RegisterNewAdapterFunc(BasicAdapterName, Upload, func(name string, dir Direction) Adapter {
		switch dir {
		case Upload:
			bu := &basicUploadAdapter{newAdapterBase(m.fs, name, dir, nil)}
			// self implements impl
			bu.transferImpl = bu
			return bu
		case Download:
			panic("Should never ask this func for basic download")
		}
		return nil
	})
}
