package tq

import (
	"net/http"
	"strings"

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
	endpoint            lfsapi.Endpoint
	transferAdapterName string `json:"transfer"`
}

func Batch(m *Manifest, dir Direction, remote string, objects []*Transfer) (*BatchResponse, error) {
	if len(objects) == 0 {
		return nil, nil
	}

	breq := &batchRequest{
		Operation:            dir.String(),
		Objects:              objects,
		TransferAdapterNames: m.GetAdapterNames(dir),
	}

	cli := &tqClient{Client: m.APIClient()}
	bres, _, err := cli.Batch(remote, breq)
	return bres, err
}

func (c *tqClient) Batch(remote string, bReq *batchRequest) (*BatchResponse, *http.Response, error) {
	bRes := &BatchResponse{}
	if len(bReq.Objects) == 0 {
		return bRes, nil, nil
	}

	if len(bReq.TransferAdapterNames) == 1 && bReq.TransferAdapterNames[0] == "basic" {
		bReq.TransferAdapterNames = nil
	}

	bRes.endpoint = c.Endpoints.Endpoint(bReq.Operation, remote)
	req, err := c.NewRequest("POST", bRes.endpoint, "objects/batch", bReq)
	if err != nil {
		return nil, nil, errors.Wrap(err, "batch request")
	}

	tracerx.Printf("api: batch %d files", len(bReq.Objects))

	res, err := c.DoWithAuth(remote, req)
	if err != nil {
		tracerx.Printf("api error: %s", err)
		return nil, nil, errors.Wrap(err, "batch response")
	}
	c.LogResponse("lfs.batch", res)

	if err := lfsapi.DecodeJSON(res, bRes); err != nil {
		return bRes, res, errors.Wrap(err, "batch response")
	}

	if res.StatusCode != 200 {
		return nil, res, errors.Errorf("Invalid status for %s %s: %d",
			req.Method,
			strings.SplitN(req.URL.String(), "?", 2)[0],
			res.StatusCode)
	}

	return bRes, res, nil
}
