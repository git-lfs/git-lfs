package api

import (
	"net/http"

	"github.com/github/git-lfs/auth"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/httputil"
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

// doApiBatchRequest runs the request to the LFS batch API. If the API returns a
// 401, the repo will be marked as having private access and the request will be
// re-run. When the repo is marked as having private access, credentials will
// be retrieved.
func DoBatchRequest(req *http.Request) (*http.Response, []*ObjectResource, error) {
	res, err := DoRequest(req, config.Config.PrivateAccess(auth.GetOperationForRequest(req)))

	if err != nil {
		if res != nil && res.StatusCode == 401 {
			return res, nil, errutil.NewAuthError(err)
		}
		return res, nil, err
	}

	var objs map[string][]*ObjectResource
	err = httputil.DecodeResponse(res, &objs)

	if err != nil {
		httputil.SetErrorResponseContext(err, res)
	}

	return res, objs["objects"], err
}

// DoRequest runs a request to the LFS API, without parsing the response
// body. If the API returns a 401, the repo will be marked as having private
// access and the request will be re-run. When the repo is marked as having
// private access, credentials will be retrieved.
func DoRequest(req *http.Request, useCreds bool) (*http.Response, error) {
	via := make([]*http.Request, 0, 4)
	return httputil.DoHttpRequestWithRedirects(req, via, useCreds)
}
