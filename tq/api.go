package tq

import (
	"time"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

type tqClient struct {
	maxRetries int
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
	HashAlgorithm        string      `json:"hash_algo"`
}

type BatchResponse struct {
	Objects             []*Transfer `json:"objects"`
	TransferAdapterName string      `json:"transfer"`
	HashAlgorithm       string      `json:"hash_algo"`
	endpoint            lfshttp.Endpoint
}

func Batch(m Manifest, dir Direction, remote string, remoteRef *git.Ref, objects []*Transfer) (*BatchResponse, error) {
	if len(objects) == 0 {
		return &BatchResponse{}, nil
	}

	cm := m.Upgrade()

	return cm.batchClient().Batch(remote, &batchRequest{
		Operation:            dir.String(),
		Objects:              objects,
		TransferAdapterNames: m.GetAdapterNames(dir),
		Ref:                  &batchRef{Name: remoteRef.Refspec()},
		HashAlgorithm:        "sha256",
	})
}

type BatchClient interface {
	Batch(remote string, bReq *batchRequest) (*BatchResponse, error)
	MaxRetries() int
	SetMaxRetries(n int)
}

func (c *tqClient) MaxRetries() int {
	return c.maxRetries
}

func (c *tqClient) SetMaxRetries(n int) {
	c.maxRetries = n
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
		return nil, errors.Wrap(err, tr.Tr.Get("batch request"))
	}

	tracerx.Printf("api: batch %d files", len(bReq.Objects))

	req = c.Client.LogRequest(req, "lfs.batch")
	res, err := c.DoAPIRequestWithAuth(remote, lfshttp.WithRetries(req, c.MaxRetries()))
	if err != nil {
		tracerx.Printf("api error: %s", err)
		return nil, errors.Wrap(err, tr.Tr.Get("batch response"))
	}

	if err := lfshttp.DecodeJSON(res, bRes); err != nil {
		return bRes, errors.Wrap(err, tr.Tr.Get("batch response"))
	}

	if bRes.HashAlgorithm != "" && bRes.HashAlgorithm != "sha256" {
		return bRes, errors.Wrap(errors.New(tr.Tr.Get("unsupported hash algorithm")), tr.Tr.Get("batch response"))
	}

	if res.StatusCode != 200 {
		return nil, lfshttp.NewStatusCodeError(res)
	}

	for _, obj := range bRes.Objects {
		for _, a := range obj.Actions {
			a.createdAt = requestedAt
		}
	}

	return bRes, nil
}
