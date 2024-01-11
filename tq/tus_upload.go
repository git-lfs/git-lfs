package tq

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
)

const (
	TusAdapterName = "tus"
	TusVersion     = "1.0.0"
)

// Adapter for tus.io protocol resumaable uploads
type tusUploadAdapter struct {
	*adapterBase
}

func (a *tusUploadAdapter) WorkerStarting(workerNum int) (interface{}, error) {
	return nil, nil
}
func (a *tusUploadAdapter) WorkerEnding(workerNum int, ctx interface{}) {
}

func (a *tusUploadAdapter) DoTransfer(ctx interface{}, t *Transfer, cb ProgressCallback, authOkFunc func()) error {
	rel, err := t.Rel("upload")
	if err != nil {
		return err
	}
	if rel == nil {
		return errors.Errorf(tr.Tr.Get("No upload action for object: %s", t.Oid))
	}

	// Note not supporting the Creation extension since the batch API generates URLs
	// Also not supporting Concatenation to support parallel uploads of chunks; forward only

	// 1. Send HEAD request to determine upload start point
	//    Request must include Tus-Resumable header (version)
	a.Trace("xfer: sending tus.io HEAD request for %q", t.Oid)
	req, err := a.newHTTPRequest("HEAD", rel)
	if err != nil {
		return err
	}

	req.Header.Set("Tus-Resumable", TusVersion)

	res, err := a.doHTTP(t, req)
	if err != nil {
		return errors.NewRetriableError(err)
	}

	//    Response will contain Upload-Offset if supported
	offHdr := res.Header.Get("Upload-Offset")
	if len(offHdr) == 0 {
		return errors.New(tr.Tr.Get("missing Upload-Offset header from tus.io HEAD response at %q, contact server admin", rel.Href))
	}
	offset, err := strconv.ParseInt(offHdr, 10, 64)
	if err != nil || offset < 0 {
		return errors.New(tr.Tr.Get("invalid Upload-Offset value %q in response from tus.io HEAD at %q, contact server admin", offHdr, rel.Href))
	}
	// Upload-Offset=size means already completed (skip)
	// Batch API will probably already detect this, but handle just in case
	if offset >= t.Size {
		a.Trace("xfer: tus.io HEAD offset %d indicates %q is already fully uploaded, skipping", offset, t.Oid)
		advanceCallbackProgress(cb, t, t.Size)
		return nil
	}

	// Open file for uploading
	f, err := os.OpenFile(t.Path, os.O_RDONLY, 0644)
	if err != nil {
		return errors.Wrap(err, tr.Tr.Get("tus.io upload"))
	}
	defer f.Close()

	// Upload-Offset=0 means start from scratch, but still send PATCH
	if offset == 0 {
		a.Trace("xfer: tus.io uploading %q from start", t.Oid)
	} else {
		a.Trace("xfer: tus.io resuming upload %q from %d", t.Oid, offset)
		advanceCallbackProgress(cb, t, offset)
	}

	// 2. Send PATCH request with byte start point (even if 0) in Upload-Offset
	//    Response status must be 204
	//    Response Upload-Offset must be request Upload-Offset plus sent bytes
	//    Response may include Upload-Expires header in which case check not passed

	a.Trace("xfer: sending tus.io PATCH request for %q", t.Oid)
	req, err = a.newHTTPRequest("PATCH", rel)
	if err != nil {
		return err
	}

	req.Header.Set("Tus-Resumable", TusVersion)
	req.Header.Set("Upload-Offset", strconv.FormatInt(offset, 10))
	req.Header.Set("Content-Type", "application/offset+octet-stream")
	req.Header.Set("Content-Length", strconv.FormatInt(t.Size-offset, 10))
	req.ContentLength = t.Size - offset

	// Ensure progress callbacks made while uploading
	// Wrap callback to give name context
	ccb := func(totalSize int64, readSoFar int64, readSinceLast int) error {
		if cb != nil {
			return cb(t.Name, totalSize, readSoFar, readSinceLast)
		}
		return nil
	}

	var reader lfsapi.ReadSeekCloser = tools.NewBodyWithCallback(f, t.Size, ccb)
	reader = newStartCallbackReader(reader, func() error {
		// seek to the offset since lfsapi.Client rewinds the body
		if _, err := f.Seek(offset, io.SeekCurrent); err != nil {
			return err
		}
		// Signal auth was ok on first read; this frees up other workers to start
		if authOkFunc != nil {
			authOkFunc()
		}
		return nil
	})

	req.Body = reader

	req = a.apiClient.LogRequest(req, "lfs.data.upload")
	res, err = a.doHTTP(t, req)
	if err != nil {
		return errors.NewRetriableError(err)
	}

	// A status code of 403 likely means that an authentication token for the
	// upload has expired. This can be safely retried.
	if res.StatusCode == 403 {
		err = errors.New(tr.Tr.Get("Received status %d", res.StatusCode))
		return errors.NewRetriableError(err)
	}

	if res.StatusCode > 299 {
		return errors.Wrapf(nil, tr.Tr.Get("Invalid status for %s %s: %d",
			req.Method,
			strings.SplitN(req.URL.String(), "?", 2)[0],
			res.StatusCode,
		))
	}

	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	return verifyUpload(a.apiClient, a.remote, t)
}

func configureTusAdapter(m *concreteManifest) {
	m.RegisterNewAdapterFunc(TusAdapterName, Upload, func(name string, dir Direction) Adapter {
		switch dir {
		case Upload:
			bu := &tusUploadAdapter{newAdapterBase(m.fs, name, dir, nil)}
			// self implements impl
			bu.transferImpl = bu
			return bu
		case Download:
			panic(tr.Tr.Get("Should never ask this function to download"))
		}
		return nil
	})
}
