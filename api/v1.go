package api

import (
	"net/http"
	"net/url"
	"path"

	"github.com/github/git-lfs/auth"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/httputil"

	"github.com/rubyist/tracerx"
)

const (
	MediaType = "application/vnd.git-lfs+json; charset=utf-8"
)

// doLegacyApiRequest runs the request to the LFS legacy API.
func DoLegacyRequest(req *http.Request) (*http.Response, *ObjectResource, error) {
	via := make([]*http.Request, 0, 4)
	res, err := httputil.DoHttpRequestWithRedirects(req, via, true)
	if err != nil {
		return res, nil, err
	}

	obj := &ObjectResource{}
	err = httputil.DecodeResponse(res, obj)

	if err != nil {
		httputil.SetErrorResponseContext(err, res)
		return nil, nil, err
	}

	return res, obj, nil
}

type BatchRequest struct {
	TransferAdapterNames []string          `json:"transfers"`
	Operation            string            `json:"operation"`
	Objects              []*ObjectResource `json:"objects"`
}
type BatchResponse struct {
	TransferAdapterName string            `json:"transfer"`
	Objects             []*ObjectResource `json:"objects"`
}

// doApiBatchRequest runs the request to the LFS batch API. If the API returns a
// 401, the repo will be marked as having private access and the request will be
// re-run. When the repo is marked as having private access, credentials will
// be retrieved.
func DoBatchRequest(req *http.Request) (*http.Response, *BatchResponse, error) {
	res, err := DoRequest(req, config.Config.PrivateAccess(auth.GetOperationForRequest(req)))

	if err != nil {
		if res != nil && res.StatusCode == 401 {
			return res, nil, errutil.NewAuthError(err)
		}
		return res, nil, err
	}

	resp := &BatchResponse{}
	err = httputil.DecodeResponse(res, resp)

	if err != nil {
		httputil.SetErrorResponseContext(err, res)
	}

	return res, resp, err
}

// DoRequest runs a request to the LFS API, without parsing the response
// body. If the API returns a 401, the repo will be marked as having private
// access and the request will be re-run. When the repo is marked as having
// private access, credentials will be retrieved.
func DoRequest(req *http.Request, useCreds bool) (*http.Response, error) {
	via := make([]*http.Request, 0, 4)
	return httputil.DoHttpRequestWithRedirects(req, via, useCreds)
}

func NewRequest(method, oid string) (*http.Request, error) {
	objectOid := oid
	operation := "download"
	if method == "POST" {
		if oid != "batch" {
			objectOid = ""
			operation = "upload"
		}
	}
	endpoint := config.Config.Endpoint(operation)

	res, err := auth.SshAuthenticate(endpoint, operation, oid)
	if err != nil {
		tracerx.Printf("ssh: attempted with %s.  Error: %s",
			endpoint.SshUserAndHost, err.Error(),
		)
		return nil, err
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

func NewBatchRequest(operation string) (*http.Request, error) {
	endpoint := config.Config.Endpoint(operation)

	res, err := auth.SshAuthenticate(endpoint, operation, "")
	if err != nil {
		tracerx.Printf("ssh: %s attempted with %s.  Error: %s",
			operation, endpoint.SshUserAndHost, err.Error(),
		)
		return nil, err
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

func ObjectUrl(endpoint config.Endpoint, oid string) (*url.URL, error) {
	u, err := url.Parse(endpoint.Url)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "objects")
	if len(oid) > 0 {
		u.Path = path.Join(u.Path, oid)
	}
	return u, nil
}
