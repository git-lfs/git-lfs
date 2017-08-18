package tq

import (
	"io"
	"net/http"
	"time"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/rubyist/tracerx"
)

type tqClient struct {
	*lfsapi.Client
}

type batchRequest struct {
	Operation            string      `json:"operation"`
	Objects              []*Transfer `json:"objects"`
	TransferAdapterNames []string    `json:"transfers,omitempty"`
}

type BatchResponse struct {
	Objects             []*Transfer `json:"objects"`
	TransferAdapterName string      `json:"transfer"`
	endpoint            lfsapi.Endpoint
}

func Batch(m *Manifest, dir Direction, remote string, objects []*Transfer) (*BatchResponse, error) {
	if len(objects) == 0 {
		return &BatchResponse{}, nil
	}

	return m.batchClient().Batch(remote, &batchRequest{
		Operation:            dir.String(),
		Objects:              objects,
		TransferAdapterNames: m.GetAdapterNames(dir),
	})
}

func (c *tqClient) Batch(remote string, bReq *batchRequest) (*BatchResponse, error) {
	bRes := &BatchResponse{}
	if len(bReq.Objects) == 0 {
		return bRes, nil
	}

	if len(bReq.TransferAdapterNames) == 1 && bReq.TransferAdapterNames[0] == "basic" {
		bReq.TransferAdapterNames = nil
	}

	bRes.endpoint = c.Endpoints.Endpoint(bReq.Operation, remote)
	requestedAt := time.Now()

	var res *http.Response
	for {
		req, err := c.NewRequest("POST", bRes.endpoint, "objects/batch", bReq)
		if err != nil {
			return nil, errors.Wrap(err, "batch request")
		}

		tracerx.Printf("api: batch %d files", len(bReq.Objects))

		req = c.LogRequest(req, "lfs.batch")
		res, err = c.DoWithAuth(remote, req)

		if err == nil {
			break
		} else if err == io.EOF {
			tracerx.Printf("api: retrying batch request after io.EOF")
			continue
		}

		tracerx.Printf("api error: %s", err)
		return nil, errors.Wrap(err, "batch response")
	}

	if err := lfsapi.DecodeJSON(res, bRes); err != nil {
		return bRes, errors.Wrap(err, "batch response")
	}

	if res.StatusCode != 200 {
		return nil, lfsapi.NewStatusCodeError(res)
	}

	for _, obj := range bRes.Objects {
		for _, a := range obj.Actions {
			a.createdAt = requestedAt
		}
	}

	return bRes, nil
}
