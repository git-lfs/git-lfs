package lfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/httputil"
	"github.com/github/git-lfs/progress"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

// Download will attempt to download the object with the given oid. The batched
// API will be used, but if the server does not implement the batch operations
// it will fall back to the legacy API.
func Download(oid string, size int64) (io.ReadCloser, int64, error) {
	if !config.Config.BatchTransfer() {
		return DownloadLegacy(oid)
	}

	objects := []*api.ObjectResource{
		&api.ObjectResource{Oid: oid, Size: size},
	}

	objs, err := Batch(objects, "download")
	if err != nil {
		if errutil.IsNotImplementedError(err) {
			git.Config.SetLocal("", "lfs.batch", "false")
			return DownloadLegacy(oid)
		}
		return nil, 0, err
	}

	if len(objs) != 1 { // Expecting to find one object
		return nil, 0, errutil.Error(fmt.Errorf("Object not found: %s", oid))
	}

	return DownloadObject(objs[0])
}

// DownloadLegacy attempts to download the object for the given oid using the
// legacy API.
func DownloadLegacy(oid string) (io.ReadCloser, int64, error) {
	req, err := api.NewRequest("GET", oid)
	if err != nil {
		return nil, 0, errutil.Error(err)
	}

	res, obj, err := api.DoLegacyRequest(req)
	if err != nil {
		return nil, 0, err
	}
	httputil.LogTransfer("lfs.api.download", res)
	req, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, 0, errutil.Error(err)
	}

	res, err = httputil.DoHttpRequest(req, true)
	if err != nil {
		return nil, 0, err
	}
	httputil.LogTransfer("lfs.data.download", res)

	return res.Body, res.ContentLength, nil
}

type byteCloser struct {
	*bytes.Reader
}

func DownloadCheck(oid string) (*api.ObjectResource, error) {
	req, err := api.NewRequest("GET", oid)
	if err != nil {
		return nil, errutil.Error(err)
	}

	res, obj, err := api.DoLegacyRequest(req)
	if err != nil {
		return nil, err
	}
	httputil.LogTransfer("lfs.api.download", res)

	_, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, errutil.Error(err)
	}

	return obj, nil
}

func DownloadObject(obj *api.ObjectResource) (io.ReadCloser, int64, error) {
	req, err := obj.NewRequest("download", "GET")
	if err != nil {
		return nil, 0, errutil.Error(err)
	}

	res, err := httputil.DoHttpRequest(req, true)
	if err != nil {
		return nil, 0, errutil.NewRetriableError(err)
	}
	httputil.LogTransfer("lfs.data.download", res)

	return res.Body, res.ContentLength, nil
}

func (b *byteCloser) Close() error {
	return nil
}

func Batch(objects []*api.ObjectResource, operation string) ([]*api.ObjectResource, error) {
	if len(objects) == 0 {
		return nil, nil
	}

	o := map[string]interface{}{"objects": objects, "operation": operation}

	by, err := json.Marshal(o)
	if err != nil {
		return nil, errutil.Error(err)
	}

	req, err := api.NewBatchRequest(operation)
	if err != nil {
		return nil, errutil.Error(err)
	}

	req.Header.Set("Content-Type", api.MediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = &byteCloser{bytes.NewReader(by)}

	tracerx.Printf("api: batch %d files", len(objects))

	res, objs, err := api.DoBatchRequest(req)

	if err != nil {

		if res == nil {
			return nil, errutil.NewRetriableError(err)
		}

		if res.StatusCode == 0 {
			return nil, errutil.NewRetriableError(err)
		}

		if errutil.IsAuthError(err) {
			httputil.SetAuthType(req, res)
			return Batch(objects, operation)
		}

		switch res.StatusCode {
		case 404, 410:
			tracerx.Printf("api: batch not implemented: %d", res.StatusCode)
			return nil, errutil.NewNotImplementedError(nil)
		}

		tracerx.Printf("api error: %s", err)
		return nil, errutil.Error(err)
	}
	httputil.LogTransfer("lfs.api.batch", res)

	if res.StatusCode != 200 {
		return nil, errutil.Error(fmt.Errorf("Invalid status for %s: %d", httputil.TraceHttpReq(req), res.StatusCode))
	}

	return objs, nil
}

func UploadCheck(oidPath string) (*api.ObjectResource, error) {
	oid := filepath.Base(oidPath)

	stat, err := os.Stat(oidPath)
	if err != nil {
		return nil, errutil.Error(err)
	}

	reqObj := &api.ObjectResource{
		Oid:  oid,
		Size: stat.Size(),
	}

	by, err := json.Marshal(reqObj)
	if err != nil {
		return nil, errutil.Error(err)
	}

	req, err := api.NewRequest("POST", oid)
	if err != nil {
		return nil, errutil.Error(err)
	}

	req.Header.Set("Content-Type", api.MediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = &byteCloser{bytes.NewReader(by)}

	tracerx.Printf("api: uploading (%s)", oid)
	res, obj, err := api.DoLegacyRequest(req)

	if err != nil {
		if errutil.IsAuthError(err) {
			httputil.SetAuthType(req, res)
			return UploadCheck(oidPath)
		}

		return nil, errutil.NewRetriableError(err)
	}
	httputil.LogTransfer("lfs.api.upload", res)

	if res.StatusCode == 200 {
		return nil, nil
	}

	if obj.Oid == "" {
		obj.Oid = oid
	}
	if obj.Size == 0 {
		obj.Size = reqObj.Size
	}

	return obj, nil
}

func UploadObject(o *api.ObjectResource, cb progress.CopyCallback) error {
	path, err := LocalMediaPath(o.Oid)
	if err != nil {
		return errutil.Error(err)
	}

	file, err := os.Open(path)
	if err != nil {
		return errutil.Error(err)
	}
	defer file.Close()

	reader := &progress.CallbackReader{
		C:         cb,
		TotalSize: o.Size,
		Reader:    file,
	}

	req, err := o.NewRequest("upload", "PUT")
	if err != nil {
		return errutil.Error(err)
	}

	if len(req.Header.Get("Content-Type")) == 0 {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	if req.Header.Get("Transfer-Encoding") == "chunked" {
		req.TransferEncoding = []string{"chunked"}
	} else {
		req.Header.Set("Content-Length", strconv.FormatInt(o.Size, 10))
	}

	req.ContentLength = o.Size
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

	if _, ok := o.Rel("verify"); !ok {
		return nil
	}

	req, err = o.NewRequest("verify", "POST")
	if err != nil {
		return errutil.Error(err)
	}

	by, err := json.Marshal(o)
	if err != nil {
		return errutil.Error(err)
	}

	req.Header.Set("Content-Type", api.MediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = ioutil.NopCloser(bytes.NewReader(by))
	res, err = api.DoRequest(req, true)
	if err != nil {
		return err
	}

	httputil.LogTransfer("lfs.data.verify", res)
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return err
}
