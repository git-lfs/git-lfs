package lfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
)

// An abstract interface providing state & resource management for a specific Endpoint across
// potentially multiple requests
type ApiContext interface {
	// Get the endpoint this context was constructed from
	Endpoint() Endpoint
	// Close the context & any resources it's using
	Close() error

	// Download a single object, return reader for data, size and any error
	Download(oid string) (io.ReadCloser, int64, *WrappedError)
	// Upload a single object
	Upload(oid string, sz int64, content io.Reader, cb CopyCallback) *WrappedError
}

var (
	contextCache map[string]ApiContext
)

// Return an API context appropriate for a given Endpoint
// This may return a new context, or an existing one which is compatible with the endpoint
func GetApiContext(endpoint Endpoint) ApiContext {
	// construct a string identifier for the Endpoint
	isSSH := false
	var id string
	if len(endpoint.SshUserAndHost) > 0 {
		isSSH = true
		// SSH will use a unique connection per path as well as user/host (passed as param)
		id = fmt.Sprintf("%s:%s", endpoint.SshUserAndHost, endpoint.SshPath)
	} else {
		// We'll use the same HTTP context for all
		id = "HTTP"
	}
	ctx, ok := contextCache[id]
	if !ok {
		// Construct new
		if isSSH {
			ctx = NewSshApiContext(endpoint)
		}
		// If not SSH, OR if full SSH server isn't supported, use HTTPS with SSH auth only
		if ctx == nil {
			ctx = NewHttpApiContext(endpoint)
		}
	}

	return ctx
}

func NewSshApiContext(endpoint Endpoint) ApiContext {
	// TODO - full SSH implementation with persistent connection
	return nil
}

// HTTP implementation of context
type HttpApiContext struct {
	endpoint Endpoint
}

func NewHttpApiContext(endpoint Endpoint) ApiContext {
	// pretty simple
	return &HttpApiContext{endpoint}
}

func (self *HttpApiContext) Endpoint() Endpoint {
	return self.endpoint
}
func (*HttpApiContext) Close() error {
	// nothing to do
	return nil
}

func (self *HttpApiContext) Download(oid string) (io.ReadCloser, int64, *WrappedError) {
	req, creds, err := self.newApiRequest("GET", oid)
	if err != nil {
		return nil, 0, Error(err)
	}

	res, obj, wErr := self.doApiRequest(req, creds)
	if wErr != nil {
		return nil, 0, wErr
	}

	req, creds, err = obj.NewRequest(self, "download", "GET")
	if err != nil {
		return nil, 0, Error(err)
	}

	res, wErr = self.doHttpRequest(req, creds)
	if wErr != nil {
		return nil, 0, wErr
	}

	return res.Body, res.ContentLength, nil
}

func (self *HttpApiContext) Upload(oid string, sz int64, file io.Reader, cb CopyCallback) *WrappedError {
	reqObj := &httpObjectResource{
		Oid:  oid,
		Size: sz,
	}

	by, err := json.Marshal(reqObj)
	if err != nil {
		return Error(err)
	}

	req, creds, err := self.newApiRequest("POST", oid)
	if err != nil {
		return Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = &byteCloser{bytes.NewReader(by)}

	res, obj, wErr := self.doApiRequest(req, creds)
	if wErr != nil {
		sendApiEvent(apiEventFail)
		return wErr
	}

	sendApiEvent(apiEventSuccess)

	reader := &CallbackReader{
		C:         cb,
		TotalSize: reqObj.Size,
		Reader:    file,
	}

	if res.StatusCode == 200 {
		// Drain the reader to update any progress bars
		io.Copy(ioutil.Discard, reader)
		return nil
	}

	req, creds, err = obj.NewRequest(self, "upload", "PUT")
	if err != nil {
		return Error(err)
	}

	if len(req.Header.Get("Content-Type")) == 0 {
		req.Header.Set("Content-Type", "application/octet-stream")
	}
	req.Header.Set("Content-Length", strconv.FormatInt(reqObj.Size, 10))
	req.ContentLength = reqObj.Size

	req.Body = ioutil.NopCloser(reader)

	res, wErr = self.doHttpRequest(req, creds)
	if wErr != nil {
		return wErr
	}

	if res.StatusCode > 299 {
		return Errorf(nil, "Invalid status for %s %s: %d", req.Method, req.URL, res.StatusCode)
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	req, creds, err = obj.NewRequest(self, "verify", "POST")
	if err == objectRelationDoesNotExist {
		return nil
	} else if err != nil {
		return Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = ioutil.NopCloser(bytes.NewReader(by))
	res, wErr = self.doHttpRequest(req, creds)

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return wErr

}
