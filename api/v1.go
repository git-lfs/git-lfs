package api

import (
	"net/http"
	"net/url"
	"path"

	"github.com/git-lfs/git-lfs/auth"
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/httputil"
	"github.com/git-lfs/git-lfs/lfsapi"

	"github.com/rubyist/tracerx"
)

const (
	MediaType = "application/vnd.git-lfs+json; charset=utf-8"
)

type batchRequest struct {
	TransferAdapterNames []string          `json:"transfers,omitempty"`
	Operation            string            `json:"operation"`
	Objects              []*ObjectResource `json:"objects"`
}
type batchResponse struct {
	TransferAdapterName string            `json:"transfer"`
	Objects             []*ObjectResource `json:"objects"`
}

// doApiBatchRequest runs the request to the LFS batch API. If the API returns a
// 401, the repo will be marked as having private access and the request will be
// re-run. When the repo is marked as having private access, credentials will
// be retrieved.
func DoBatchRequest(cfg *config.Configuration, req *http.Request) (*http.Response, *batchResponse, error) {
	res, err := DoRequest(req, cfg.PrivateAccess(auth.GetOperationForRequest(req)))

	if err != nil {
		if res != nil && res.StatusCode == 401 {
			return res, nil, errors.NewAuthError(err)
		}
		return res, nil, err
	}

	resp := &batchResponse{}
	err = httputil.DecodeResponse(res, resp)

	if err != nil {
		httputil.SetErrorResponseContext(cfg, err, res)
	}

	return res, resp, err
}

// DoRequest runs a request to the LFS API, without parsing the response
// body. If the API returns a 401, the repo will be marked as having private
// access and the request will be re-run. When the repo is marked as having
// private access, credentials will be retrieved.
func DoRequest(req *http.Request, useCreds bool) (*http.Response, error) {
	via := make([]*http.Request, 0, 4)
	return httputil.DoHttpRequestWithRedirects(config.Config, req, via, useCreds)
}

func NewRequest(cfg *config.Configuration, method, oid string) (*http.Request, error) {
	objectOid := oid
	operation := "download"
	if method == "POST" {
		if oid != "batch" {
			objectOid = ""
			operation = "upload"
		}
	}

	res, endpoint, err := auth.SshAuthenticate(cfg, operation, oid)
	if err != nil {
		tracerx.Printf("ssh: %s with %s failed, error: %s, message: %s",
			operation, endpoint.SshUserAndHost, err.Error(), res.Message,
		)
		return nil, errors.Wrap(errors.New(res.Message), err.Error())
	}

	if len(res.Href) > 0 {
		endpoint.Url = res.Href
	}

	u, err := ObjectUrl(endpoint, objectOid)
	if err != nil {
		return nil, err
	}

	req, err := httputil.NewHttpRequest(method, u.String(), res.Header)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", MediaType)
	return req, nil
}

func NewBatchRequest(cfg *config.Configuration, operation string) (*http.Request, error) {
	res, endpoint, err := auth.SshAuthenticate(cfg, operation, "")
	if err != nil {
		tracerx.Printf("ssh: %s with %s failed, error: %s, message: %s",
			operation, endpoint.SshUserAndHost, err.Error(), res.Message,
		)
		return nil, errors.Wrap(errors.New(res.Message), err.Error())
	}

	if len(res.Href) > 0 {
		endpoint.Url = res.Href
	}

	u, err := ObjectUrl(endpoint, "batch")
	if err != nil {
		return nil, err
	}

	req, err := httputil.NewHttpRequest("POST", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", MediaType)
	if res.Header != nil {
		for key, value := range res.Header {
			req.Header.Set(key, value)
		}
	}

	return req, nil
}

func ObjectUrl(e lfsapi.Endpoint, oid string) (*url.URL, error) {
	u, err := url.Parse(e.Url)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "objects")
	if len(oid) > 0 {
		u.Path = path.Join(u.Path, oid)
	}
	return u, nil
}
