package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/httputil"
)

// VerifyUpload calls the "verify" API link relation on obj if it exists
func VerifyUpload(obj *ObjectResource) error {

	// Do we need to do verify?
	if _, ok := obj.Rel("verify"); !ok {
		return nil
	}

	req, err := obj.NewRequest("verify", "POST")
	if err != nil {
		return errutil.Error(err)
	}

	by, err := json.Marshal(obj)
	if err != nil {
		return errutil.Error(err)
	}

	req.Header.Set("Content-Type", MediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = ioutil.NopCloser(bytes.NewReader(by))
	res, err := DoRequest(req, true)
	if err != nil {
		return err
	}

	httputil.LogTransfer("lfs.data.verify", res)
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return err
}
