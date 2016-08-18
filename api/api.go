// Package api provides the interface for querying LFS servers (metadata)
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errors"
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/httputil"
	"github.com/github/git-lfs/tools"

	"github.com/rubyist/tracerx"
)

// BatchOrLegacy calls the Batch API and falls back on the Legacy API
// This is for simplicity, legacy route is not most optimal (serial)
// TODO LEGACY API: remove when legacy API removed
func BatchOrLegacy(cfg *config.Configuration, objects []*ObjectResource, operation string, transferAdapters []string) (objs []*ObjectResource, transferAdapter string, e error) {
	if !cfg.BatchTransfer() {
		objs, err := Legacy(cfg, objects, operation)
		return objs, "", err
	}
	objs, adapterName, err := Batch(cfg, objects, operation, transferAdapters)
	if err != nil {
		if errors.IsNotImplementedError(err) {
			git.Config.SetLocal("", "lfs.batch", "false")
			objs, err := Legacy(cfg, objects, operation)
			return objs, "", err
		}
		return nil, "", err
	}
	return objs, adapterName, nil
}

func BatchOrLegacySingle(cfg *config.Configuration, inobj *ObjectResource, operation string, transferAdapters []string) (obj *ObjectResource, transferAdapter string, e error) {
	objs, adapterName, err := BatchOrLegacy(cfg, []*ObjectResource{inobj}, operation, transferAdapters)
	if err != nil {
		return nil, "", err
	}
	if len(objs) > 0 {
		return objs[0], adapterName, nil
	}
	return nil, "", fmt.Errorf("Object not found")
}

// Batch calls the batch API and returns object results
func Batch(cfg *config.Configuration, objects []*ObjectResource, operation string, transferAdapters []string) (objs []*ObjectResource, transferAdapter string, e error) {
	if len(objects) == 0 {
		return nil, "", nil
	}

	// Compatibility; omit transfers list when only basic
	// older schemas included `additionalproperties=false`
	if len(transferAdapters) == 1 && transferAdapters[0] == "basic" {
		transferAdapters = nil
	}

	o := &batchRequest{Operation: operation, Objects: objects, TransferAdapterNames: transferAdapters}
	by, err := json.Marshal(o)
	if err != nil {
		return nil, "", errors.Wrap(err, "batch request")
	}

	req, err := NewBatchRequest(cfg, operation)
	if err != nil {
		return nil, "", errors.Wrap(err, "batch request")
	}

	req.Header.Set("Content-Type", MediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = tools.NewReadSeekCloserWrapper(bytes.NewReader(by))

	tracerx.Printf("api: batch %d files", len(objects))

	res, bresp, err := DoBatchRequest(cfg, req)

	if err != nil {
		if res == nil {
			return nil, "", errors.NewRetriableError(err)
		}

		if res.StatusCode == 0 {
			return nil, "", errors.NewRetriableError(err)
		}

		if errors.IsAuthError(err) {
			httputil.SetAuthType(cfg, req, res)
			return Batch(cfg, objects, operation, transferAdapters)
		}

		switch res.StatusCode {
		case 404, 410:
			return nil, "", errors.NewNotImplementedError(errors.Errorf("api: batch not implemented: %d", res.StatusCode))
		}

		tracerx.Printf("api error: %s", err)
		return nil, "", errors.Wrap(err, "batch response")
	}
	httputil.LogTransfer(cfg, "lfs.batch", res)

	if res.StatusCode != 200 {
		return nil, "", errors.Errorf("Invalid status for %s: %d", httputil.TraceHttpReq(req), res.StatusCode)
	}

	return bresp.Objects, bresp.TransferAdapterName, nil
}

// Legacy calls the legacy API serially and returns ObjectResources
// TODO LEGACY API: remove when legacy API removed
func Legacy(cfg *config.Configuration, objects []*ObjectResource, operation string) ([]*ObjectResource, error) {
	retobjs := make([]*ObjectResource, 0, len(objects))
	dl := operation == "download"
	var globalErr error
	for _, o := range objects {
		var ret *ObjectResource
		var err error
		if dl {
			ret, err = DownloadCheck(cfg, o.Oid)
		} else {
			ret, err = UploadCheck(cfg, o.Oid, o.Size)
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
func DownloadCheck(cfg *config.Configuration, oid string) (*ObjectResource, error) {
	req, err := NewRequest(cfg, "GET", oid)
	if err != nil {
		return nil, errors.Wrap(err, "download check")
	}

	res, obj, err := DoLegacyRequest(cfg, req)
	if err != nil {
		return nil, err
	}

	httputil.LogTransfer(cfg, "lfs.download", res)

	_, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, errors.Wrap(err, "download check")
	}

	return obj, nil
}

// TODO LEGACY API: remove when legacy API removed
func UploadCheck(cfg *config.Configuration, oid string, size int64) (*ObjectResource, error) {
	reqObj := &ObjectResource{
		Oid:  oid,
		Size: size,
	}

	by, err := json.Marshal(reqObj)
	if err != nil {
		return nil, errors.Wrap(err, "upload check")
	}

	req, err := NewRequest(cfg, "POST", oid)
	if err != nil {
		return nil, errors.Wrap(err, "upload check")
	}

	req.Header.Set("Content-Type", MediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = tools.NewReadSeekCloserWrapper(bytes.NewReader(by))

	tracerx.Printf("api: uploading (%s)", oid)
	res, obj, err := DoLegacyRequest(cfg, req)

	if err != nil {
		if errors.IsAuthError(err) {
			httputil.SetAuthType(cfg, req, res)
			return UploadCheck(cfg, oid, size)
		}

		return nil, errors.NewRetriableError(err)
	}
	httputil.LogTransfer(cfg, "lfs.upload", res)

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
