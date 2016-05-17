package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/httputil"
	"github.com/github/git-lfs/tools"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

func Batch(objects []*ObjectResource, operation string) ([]*ObjectResource, error) {
	if len(objects) == 0 {
		return nil, nil
	}

	o := map[string]interface{}{"objects": objects, "operation": operation}

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

	res, objs, err := DoBatchRequest(req)

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
	httputil.LogTransfer("lfs.batch", res)

	if res.StatusCode != 200 {
		return nil, errutil.Error(fmt.Errorf("Invalid status for %s: %d", httputil.TraceHttpReq(req), res.StatusCode))
	}

	return objs, nil
}
