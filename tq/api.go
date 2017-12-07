package tq

import (
	"time"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/rubyist/tracerx"
)

type tqClient struct {
	MaxRetries int
	*lfsapi.Client
}

type batchRef struct {
	Name string `json:"name,omitempty"`
}

type batchRequest struct {
	Operation            string      `json:"operation"`
	Objects              []*Transfer `json:"objects"`
	TransferAdapterNames []string    `json:"transfers,omitempty"`
	Ref                  *batchRef   `json:"ref"`
}

type BatchResponse struct {
	Objects             []*Transfer `json:"objects"`
	TransferAdapterName string      `json:"transfer"`
	endpoint            lfsapi.Endpoint
}

func Batch(m *Manifest, dir Direction, remote string, remoteRef *git.Ref, objects []*Transfer) (*BatchResponse, error) {
	if len(objects) == 0 {
		return &BatchResponse{}, nil
	}

	return m.batchClient().Batch(remote, &batchRequest{
		Operation:            dir.String(),
		Objects:              objects,
		TransferAdapterNames: m.GetAdapterNames(dir),
		Ref:                  &batchRef{Name: remoteRef.Refspec()},
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

	req, err := c.NewRequest("POST", bRes.endpoint, "objects/batch", bReq)
	if err != nil {
		return nil, errors.Wrap(err, "batch request")
	}

	tracerx.Printf("api: batch %d files", len(bReq.Objects))

	req = c.LogRequest(req, "lfs.batch")
	res, err := c.DoWithAuth(remote, lfsapi.WithRetries(req, c.MaxRetries))
	if err != nil {
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
