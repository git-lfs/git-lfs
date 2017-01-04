package tq

import (
	"net/http"

	"github.com/git-lfs/git-lfs/api"
	"github.com/git-lfs/git-lfs/lfsapi"
)

type tqClient struct {
	*lfsapi.Client
}

type batchRequest struct {
	Operation            string                `json:"operation"`
	Objects              []*api.ObjectResource `json:"objects"`
	TransferAdapterNames []string              `json:"transfers,omitempty"`
}

type batchResponse struct {
	TransferAdapterName string                `json:"transfer"`
	Objects             []*api.ObjectResource `json:"objects"`
}

func (c *tqClient) Batch(remote string, bReq *batchRequest) (*batchResponse, *http.Response, error) {
	bRes := &batchResponse{}
	if len(bReq.Objects) == 0 {
		return bRes, nil, nil
	}

	if len(bReq.TransferAdapterNames) == 1 && bReq.TransferAdapterNames[0] == "basic" {
		bReq.TransferAdapterNames = nil
	}

	e := c.Endpoints.Endpoint(bReq.Operation, remote)
	req, err := c.NewRequest("POST", e, "objects/batch", bReq)
	if err != nil {
		return nil, nil, err
	}

	res, err := c.DoWithAuth(remote, req)
	if err != nil {
		return nil, nil, err
	}

	return bRes, res, lfsapi.DecodeJSON(res, bRes)
}
