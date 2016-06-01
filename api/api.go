// Package api provides the interface for querying LFS servers (metadata)
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/httputil"
	"github.com/github/git-lfs/tools"

	"github.com/rubyist/tracerx"
)

// BatchOrLegacy calls the Batch API and falls back on the Legacy API
// This is for simplicity, legacy route is not most optimal (serial)
// TODO LEGACY API: remove when legacy API removed
func BatchOrLegacy(objects []*ObjectResource, operation string, transferAdapters []string) (*BatchResponse, error) {
	if !config.Config.BatchTransfer() {
		bresp := &BatchResponse{TransferAdapterName: "basic"}
		objs, err := Legacy(objects, operation)
		bresp.Objects = objs
		return bresp, err
	}
	bresp, err := Batch(objects, operation, transferAdapters)
	if err != nil {
		if errutil.IsNotImplementedError(err) {
			git.Config.SetLocal("", "lfs.batch", "false")
			bresp = &BatchResponse{TransferAdapterName: "basic"}
			objs, err := Legacy(objects, operation)
			bresp.Objects = objs
			return bresp, err
		}
		return nil, err
	}
	return bresp, nil
}

func BatchOrLegacySingle(inobj *ObjectResource, operation string, transferAdapters []string) (*ObjectResource, error) {
	bresp, err := BatchOrLegacy([]*ObjectResource{inobj}, operation, transferAdapters)
	if err != nil {
		return nil, err
	}
	if bresp != nil && len(bresp.Objects) > 0 {
		return bresp.Objects[0], nil
	}
	return nil, fmt.Errorf("Object not found")
}

// Batch calls the batch API and returns object results
func Batch(objects []*ObjectResource, operation string, transferAdapters []string) (*BatchResponse, error) {
	if len(objects) == 0 {
		return nil, nil
	}

	o := &BatchRequest{Operation: operation, Objects: objects, TransferAdapterNames: transferAdapters}
	by, err := json.Marshal(o)
	if err != nil {
		return nil, errutil.Error(err)
	}

	req, err := NewBatchRequest(operation)
	if err != nil {
		return nil, errutil.Error(err)
	}

	req.Header.Set("Content-Type", MediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = tools.NewReadSeekCloserWrapper(bytes.NewReader(by))

	tracerx.Printf("api: batch %d files", len(objects))

	res, bresp, err := DoBatchRequest(req)

	if err != nil {

		if res == nil {
			return nil, errutil.NewRetriableError(err)
		}

		if res.StatusCode == 0 {
			return nil, errutil.NewRetriableError(err)
		}

		if errutil.IsAuthError(err) {
			httputil.SetAuthType(req, res)
			return Batch(objects, operation, transferAdapters)
		}

		switch res.StatusCode {
		case 404, 410:
			tracerx.Printf("api: batch not implemented: %d", res.StatusCode)
			return nil, errutil.NewNotImplementedError(nil)
		}

		tracerx.Printf("api error: %s", err)
		return nil, errutil.Error(err)
	}
	httputil.LogTransfer("lfs.batch", res)

	if res.StatusCode != 200 {
		return nil, errutil.Error(fmt.Errorf("Invalid status for %s: %d", httputil.TraceHttpReq(req), res.StatusCode))
	}

	return bresp, nil
}

// Legacy calls the legacy API serially and returns ObjectResources
// TODO LEGACY API: remove when legacy API removed
func Legacy(objects []*ObjectResource, operation string) ([]*ObjectResource, error) {
	retobjs := make([]*ObjectResource, 0, len(objects))
	dl := operation == "download"
	var globalErr error
	for _, o := range objects {
		var ret *ObjectResource
		var err error
		if dl {
			ret, err = DownloadCheck(o.Oid)
		} else {
			ret, err = UploadCheck(o.Oid, o.Size)
		}
		if err != nil {
			// Store for the end, likely only one
			globalErr = err
		}
		retobjs = append(retobjs, ret)
	}
	return retobjs, globalErr
}

// TODO LEGACY API: remove when legacy API removed
func DownloadCheck(oid string) (*ObjectResource, error) {
	req, err := NewRequest("GET", oid)
	if err != nil {
		return nil, errutil.Error(err)
	}

	res, obj, err := DoLegacyRequest(req)
	if err != nil {
		return nil, err
	}
	httputil.LogTransfer("lfs.download", res)

	_, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, errutil.Error(err)
	}

	return obj, nil
}

// TODO LEGACY API: remove when legacy API removed
func UploadCheck(oid string, size int64) (*ObjectResource, error) {

	reqObj := &ObjectResource{
		Oid:  oid,
		Size: size,
	}

	by, err := json.Marshal(reqObj)
	if err != nil {
		return nil, errutil.Error(err)
	}

	req, err := NewRequest("POST", oid)
	if err != nil {
		return nil, errutil.Error(err)
	}

	req.Header.Set("Content-Type", MediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = tools.NewReadSeekCloserWrapper(bytes.NewReader(by))

	tracerx.Printf("api: uploading (%s)", oid)
	res, obj, err := DoLegacyRequest(req)

	if err != nil {
		if errutil.IsAuthError(err) {
			httputil.SetAuthType(req, res)
			return UploadCheck(oid, size)
		}

		return nil, errutil.NewRetriableError(err)
	}
	httputil.LogTransfer("lfs.upload", res)

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
