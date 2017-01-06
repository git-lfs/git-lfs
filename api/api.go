// Package api provides the interface for querying LFS servers (metadata)
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/httputil"

	"github.com/rubyist/tracerx"
)

// BatchSingle calls the batch API with just a single object.
func BatchSingle(cfg *config.Configuration, inobj *ObjectResource, operation string, transferAdapters []string) (obj *ObjectResource, transferAdapter string, e error) {
	objs, adapterName, err := Batch(cfg, []*ObjectResource{inobj}, operation, transferAdapters)
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
	req.Body = newByteBody(by)

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

		tracerx.Printf("api error: %s", err)
		return nil, "", errors.Wrap(err, "batch response")
	}
	httputil.LogTransfer(cfg, "lfs.batch", res)

	if res.StatusCode != 200 {
		return nil, "", errors.Errorf("Invalid status for %s: %d", httputil.TraceHttpReq(req), res.StatusCode)
	}

	return bresp.Objects, bresp.TransferAdapterName, nil
}

// temporary copied code from lfsapi, since api is going away
func newByteBody(by []byte) *closingByteReader {
	return &closingByteReader{Reader: bytes.NewReader(by)}
}

type closingByteReader struct {
	*bytes.Reader
}

func (r *closingByteReader) Close() error {
	return nil
}
