package lfs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

const (
	mediaType = "application/vnd.git-lfs+json; charset=utf-8"
)

var (
	lfsMediaTypeRE             = regexp.MustCompile(`\Aapplication/vnd\.git\-lfs\+json(;|\z)`)
	jsonMediaTypeRE            = regexp.MustCompile(`\Aapplication/json(;|\z)`)
	objectRelationDoesNotExist = errors.New("relation does not exist")
	hiddenHeaders              = map[string]bool{
		"Authorization": true,
	}

	defaultErrors = map[int]string{
		400: "Client error: %s",
		401: "Authorization error: %s\nCheck that you have proper access to the repository",
		403: "Authorization error: %s\nCheck that you have proper access to the repository",
		404: "Repository or object not found: %s\nCheck that it exists and that you have proper access to it",
		500: "Server error: %s",
	}
)

type objectError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *objectError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

type objectResource struct {
	Oid     string                   `json:"oid,omitempty"`
	Size    int64                    `json:"size"`
	Actions map[string]*linkRelation `json:"actions,omitempty"`
	Links   map[string]*linkRelation `json:"_links,omitempty"`
	Error   *objectError             `json:"error,omitempty"`
}

func (o *objectResource) NewRequest(relation, method string) (*http.Request, Creds, error) {
	rel, ok := o.Rel(relation)
	if !ok {
		return nil, nil, objectRelationDoesNotExist
	}

	req, creds, err := newClientRequest(method, rel.Href, rel.Header)
	if err != nil {
		return nil, nil, err
	}

	return req, creds, nil
}

func (o *objectResource) Rel(name string) (*linkRelation, bool) {
	var rel *linkRelation
	var ok bool

	if o.Actions != nil {
		rel, ok = o.Actions[name]
	} else {
		rel, ok = o.Links[name]
	}

	return rel, ok
}

type linkRelation struct {
	Href   string            `json:"href"`
	Header map[string]string `json:"header,omitempty"`
}

type ClientError struct {
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url,omitempty"`
	RequestId        string `json:"request_id,omitempty"`
}

func (e *ClientError) Error() string {
	msg := e.Message
	if len(e.DocumentationUrl) > 0 {
		msg += "\nDocs: " + e.DocumentationUrl
	}
	if len(e.RequestId) > 0 {
		msg += "\nRequest ID: " + e.RequestId
	}
	return msg
}

func Download(oid string) (io.ReadCloser, int64, *WrappedError) {
	req, creds, err := newApiRequest("GET", oid)
	if err != nil {
		return nil, 0, Error(err)
	}

	res, obj, wErr := doApiRequest(req, creds)
	if wErr != nil {
		return nil, 0, wErr
	}
	LogTransfer("lfs.api.download", res)

	req, creds, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, 0, Error(err)
	}

	res, wErr = doHttpRequest(req, creds)
	if wErr != nil {
		return nil, 0, wErr
	}
	LogTransfer("lfs.data.download", res)

	return res.Body, res.ContentLength, nil
}

type byteCloser struct {
	*bytes.Reader
}

func DownloadCheck(oid string) (*objectResource, *WrappedError) {
	req, creds, err := newApiRequest("GET", oid)
	if err != nil {
		return nil, Error(err)
	}

	res, obj, wErr := doApiRequest(req, creds)
	if wErr != nil {
		return nil, wErr
	}
	LogTransfer("lfs.api.download", res)

	_, _, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, Error(err)
	}

	return obj, nil
}

func DownloadObject(obj *objectResource) (io.ReadCloser, int64, *WrappedError) {
	req, creds, err := obj.NewRequest("download", "GET")
	if err != nil {
		return nil, 0, Error(err)
	}

	res, wErr := doHttpRequest(req, creds)
	if wErr != nil {
		return nil, 0, wErr
	}
	LogTransfer("lfs.data.download", res)

	return res.Body, res.ContentLength, nil
}

func (b *byteCloser) Close() error {
	return nil
}

func Batch(objects []*objectResource, operation string) ([]*objectResource, *WrappedError) {
	if len(objects) == 0 {
		return nil, nil
	}

	o := map[string]interface{}{"objects": objects, "operation": operation}

	by, err := json.Marshal(o)
	if err != nil {
		return nil, Error(err)
	}

	req, creds, err := newBatchApiRequest()
	if err != nil {
		return nil, Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = &byteCloser{bytes.NewReader(by)}

	tracerx.Printf("api: batch %d files", len(objects))
	res, objs, wErr := doApiBatchRequest(req, creds)
	if wErr != nil {
		if res == nil {
			return nil, wErr
		}

		switch res.StatusCode {
		case 401:
			Config.SetPrivateAccess()
			tracerx.Printf("api: batch not authorized, submitting with auth")
			return Batch(objects, operation)
		case 404, 410:
			tracerx.Printf("api: batch not implemented: %d", res.StatusCode)
			return nil, Error(newNotImplError())
		}

		tracerx.Printf("api error: %s", wErr)
	}
	LogTransfer("lfs.api.batch", res)

	if res.StatusCode != 200 {
		return nil, Errorf(nil, "Invalid status for %s %s: %d", req.Method, req.URL, res.StatusCode)
	}

	return objs, nil
}

func UploadCheck(oidPath string) (*objectResource, *WrappedError) {
	oid := filepath.Base(oidPath)

	stat, err := os.Stat(oidPath)
	if err != nil {
		return nil, Error(err)
	}

	reqObj := &objectResource{
		Oid:  oid,
		Size: stat.Size(),
	}

	by, err := json.Marshal(reqObj)
	if err != nil {
		return nil, Error(err)
	}

	req, creds, err := newApiRequest("POST", oid)
	if err != nil {
		return nil, Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = &byteCloser{bytes.NewReader(by)}

	tracerx.Printf("api: uploading (%s)", oid)
	res, obj, wErr := doApiRequest(req, creds)
	if wErr != nil {
		return nil, wErr
	}
	LogTransfer("lfs.api.upload", res)

	if res.StatusCode == 200 {
		return nil, nil
	}

	if obj.Oid == "" {
		obj.Oid = oid
	}
	if obj.Size == 0 {
		obj.Size = reqObj.Size
	}

	return obj, nil
}

func UploadObject(o *objectResource, cb CopyCallback) *WrappedError {
	path, err := LocalMediaPath(o.Oid)
	if err != nil {
		return Error(err)
	}

	file, err := os.Open(path)
	if err != nil {
		return Error(err)
	}
	defer file.Close()

	reader := &CallbackReader{
		C:         cb,
		TotalSize: o.Size,
		Reader:    file,
	}

	req, creds, err := o.NewRequest("upload", "PUT")
	if err != nil {
		return Error(err)
	}

	if len(req.Header.Get("Content-Type")) == 0 {
		req.Header.Set("Content-Type", "application/octet-stream")
	}
	if req.Header.Get("Transfer-Encoding") == "chunked" {
		req.TransferEncoding = []string{"chunked"}
	} else {
		req.Header.Set("Content-Length", strconv.FormatInt(o.Size, 10))
	}
	req.ContentLength = o.Size

	req.Body = ioutil.NopCloser(reader)

	res, wErr := doHttpRequest(req, creds)
	if wErr != nil {
		return wErr
	}
	LogTransfer("lfs.data.upload", res)

	if res.StatusCode > 299 {
		return Errorf(nil, "Invalid status for %s %s: %d", req.Method, req.URL, res.StatusCode)
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	req, creds, err = o.NewRequest("verify", "POST")
	if err == objectRelationDoesNotExist {
		return nil
	} else if err != nil {
		return Error(err)
	}

	by, err := json.Marshal(o)
	if err != nil {
		return Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = ioutil.NopCloser(bytes.NewReader(by))
	res, wErr = doHttpRequest(req, creds)
	if wErr != nil {
		return wErr
	}

	LogTransfer("lfs.data.verify", res)
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return wErr
}

func doHttpRequest(req *http.Request, creds Creds) (*http.Response, *WrappedError) {
	res, err := Config.HttpClient().Do(req)
	if res == nil {
		res = &http.Response{
			StatusCode: 0,
			Header:     make(http.Header),
			Request:    req,
			Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		}
	}

	var wErr *WrappedError

	if err != nil {
		wErr = Errorf(err, "Error for %s %s", res.Request.Method, res.Request.URL)
	} else {
		saveCredentials(creds, res)
		wErr = handleResponse(res)
	}

	if wErr != nil {
		if res != nil {
			setErrorResponseContext(wErr, res)
		} else {
			setErrorRequestContext(wErr, req)
		}
	}

	return res, wErr
}

func doApiRequestWithRedirects(req *http.Request, creds Creds, via []*http.Request) (*http.Response, *WrappedError) {
	res, wErr := doHttpRequest(req, creds)
	if wErr != nil {
		return res, wErr
	}

	if res.StatusCode == 307 {
		redirectTo := res.Header.Get("Location")
		locurl, err := url.Parse(redirectTo)
		if err == nil && !locurl.IsAbs() {
			locurl = req.URL.ResolveReference(locurl)
			redirectTo = locurl.String()
		}

		redirectedReq, redirectedCreds, err := newClientRequest(req.Method, redirectTo, nil)
		if err != nil {
			return res, Errorf(err, err.Error())
		}

		via = append(via, req)

		// Avoid seeking and re-wrapping the countingReadCloser, just get the "real" body
		realBody := req.Body
		if wrappedBody, ok := req.Body.(*countingReadCloser); ok {
			realBody = wrappedBody.ReadCloser
		}

		seeker, ok := realBody.(io.Seeker)
		if !ok {
			return res, Errorf(nil, "Request body needs to be an io.Seeker to handle redirects.")
		}

		if _, err := seeker.Seek(0, 0); err != nil {
			return res, Error(err)
		}
		redirectedReq.Body = realBody
		redirectedReq.ContentLength = req.ContentLength

		if err = checkRedirect(redirectedReq, via); err != nil {
			return res, Errorf(err, err.Error())
		}

		return doApiRequestWithRedirects(redirectedReq, redirectedCreds, via)
	}

	return res, nil
}

func doApiRequest(req *http.Request, creds Creds) (*http.Response, *objectResource, *WrappedError) {
	via := make([]*http.Request, 0, 4)
	res, wErr := doApiRequestWithRedirects(req, creds, via)
	if wErr != nil {
		return res, nil, wErr
	}

	obj := &objectResource{}
	wErr = decodeApiResponse(res, obj)

	if wErr != nil {
		setErrorResponseContext(wErr, res)
		return nil, nil, wErr
	}

	return res, obj, nil
}

// doApiBatchRequest runs the request to the batch API. If the API returns a 401,
// the repo will be marked as having private access and the request will be
// re-run. When the repo is marked as having private access, credentials will
// be retrieved.
func doApiBatchRequest(req *http.Request, creds Creds) (*http.Response, []*objectResource, *WrappedError) {
	via := make([]*http.Request, 0, 4)
	res, wErr := doApiRequestWithRedirects(req, creds, via)

	if wErr != nil {
		return res, nil, wErr
	}

	var objs map[string][]*objectResource
	wErr = decodeApiResponse(res, &objs)

	if wErr != nil {
		setErrorResponseContext(wErr, res)
	}

	return res, objs["objects"], wErr
}

func handleResponse(res *http.Response) *WrappedError {
	if res.StatusCode < 400 {
		return nil
	}

	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	cliErr := &ClientError{}
	wErr := decodeApiResponse(res, cliErr)
	if wErr == nil {
		if len(cliErr.Message) == 0 {
			wErr = defaultError(res)
		} else {
			wErr = Error(cliErr)
		}
	}

	wErr.Panic = res.StatusCode > 499 && res.StatusCode != 501 && res.StatusCode != 509
	return wErr
}

func decodeApiResponse(res *http.Response, obj interface{}) *WrappedError {
	ctype := res.Header.Get("Content-Type")
	if !(lfsMediaTypeRE.MatchString(ctype) || jsonMediaTypeRE.MatchString(ctype)) {
		return nil
	}

	err := json.NewDecoder(res.Body).Decode(obj)
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	if err != nil {
		return Errorf(err, "Unable to parse HTTP response for %s %s", res.Request.Method, res.Request.URL)
	}

	return nil
}

func defaultError(res *http.Response) *WrappedError {
	var msgFmt string

	if f, ok := defaultErrors[res.StatusCode]; ok {
		msgFmt = f
	} else if res.StatusCode < 500 {
		msgFmt = defaultErrors[400] + fmt.Sprintf(" from HTTP %d", res.StatusCode)
	} else {
		msgFmt = defaultErrors[500] + fmt.Sprintf(" from HTTP %d", res.StatusCode)
	}

	return Error(fmt.Errorf(msgFmt, res.Request.URL))
}

func saveCredentials(creds Creds, res *http.Response) {
	if creds == nil {
		return
	}

	if res.StatusCode < 300 {
		execCreds(creds, "approve")
	} else if res.StatusCode == 401 {
		execCreds(creds, "reject")
	}
}

func newApiRequest(method, oid string) (*http.Request, Creds, error) {
	endpoint := Config.Endpoint()
	objectOid := oid
	operation := "download"
	if method == "POST" {
		if oid != "batch" {
			objectOid = ""
			operation = "upload"
		}
	}

	res, err := sshAuthenticate(endpoint, operation, oid)
	if err != nil {
		tracerx.Printf("ssh: attempted with %s.  Error: %s",
			endpoint.SshUserAndHost, err.Error(),
		)
	}

	if len(res.Href) > 0 {
		endpoint.Url = res.Href
	}

	u, err := ObjectUrl(endpoint, objectOid)
	if err != nil {
		return nil, nil, err
	}

	req, creds, err := newClientRequest(method, u.String(), res.Header)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", mediaType)
	return req, creds, nil
}

func newClientRequest(method, rawurl string, header map[string]string) (*http.Request, Creds, error) {
	req, err := http.NewRequest(method, rawurl, nil)
	if err != nil {
		return nil, nil, err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	req.Header.Set("User-Agent", UserAgent)
	creds, err := getCreds(req)
	if err != nil {
		return nil, nil, err
	}

	return req, creds, nil
}

func newBatchApiRequest() (*http.Request, Creds, error) {
	endpoint := Config.Endpoint()

	res, err := sshAuthenticate(endpoint, "download", "")
	if err != nil {
		tracerx.Printf("ssh: attempted with %s.  Error: %s",
			endpoint.SshUserAndHost, err.Error(),
		)
	}

	if len(res.Href) > 0 {
		endpoint.Url = res.Href
	}

	u, err := ObjectUrl(endpoint, "batch")
	if err != nil {
		return nil, nil, err
	}

	req, creds, err := newBatchClientRequest("POST", u.String())
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", mediaType)
	if res.Header != nil {
		for key, value := range res.Header {
			req.Header.Set(key, value)
		}
	}

	return req, creds, nil
}

func newBatchClientRequest(method, rawurl string) (*http.Request, Creds, error) {
	req, err := http.NewRequest(method, rawurl, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	// Get the creds if we're private
	if Config.PrivateAccess() {
		// The PrivateAccess() check can be pushed down and this block simplified
		// once everything goes through the batch endpoint.
		creds, err := getCreds(req)
		if err != nil {
			return nil, nil, err
		}

		return req, creds, nil
	}

	return req, nil, nil
}

func getCreds(req *http.Request) (Creds, error) {
	if len(req.Header.Get("Authorization")) > 0 {
		return nil, nil
	}

	apiUrl, err := Config.ObjectUrl("")
	if err != nil {
		return nil, err
	}

	if req.URL.Scheme != apiUrl.Scheme ||
		req.URL.Host != apiUrl.Host {
		return nil, nil
	}

	if setRequestAuthFromUrl(req, apiUrl) {
		return nil, nil
	}

	credsUrl := apiUrl
	if len(Config.CurrentRemote) > 0 {
		if u, ok := Config.GitConfig("remote." + Config.CurrentRemote + ".url"); ok {
			gitRemoteUrl, err := url.Parse(u)
			if err != nil {
				return nil, err
			}

			if gitRemoteUrl.Scheme == apiUrl.Scheme &&
				gitRemoteUrl.Host == apiUrl.Host {

				if setRequestAuthFromUrl(req, gitRemoteUrl) {
					return nil, nil
				}

				credsUrl = gitRemoteUrl
			}
		}
	}

	creds, err := credentials(credsUrl)
	if err != nil {
		return nil, err
	}

	setRequestAuth(req, creds["username"], creds["password"])
	return creds, nil
}

func setRequestAuthFromUrl(req *http.Request, u *url.URL) bool {
	if u.User != nil {
		if pass, ok := u.User.Password(); ok {
			fmt.Fprintln(os.Stderr, "warning: current Git remote contains credentials")
			setRequestAuth(req, u.User.Username(), pass)
			return true
		}
	}

	return false
}

func setRequestAuth(req *http.Request, user, pass string) {
	token := fmt.Sprintf("%s:%s", user, pass)
	auth := "Basic " + base64.URLEncoding.EncodeToString([]byte(token))
	req.Header.Set("Authorization", auth)
}

func setErrorResponseContext(err *WrappedError, res *http.Response) {
	err.Set("Status", res.Status)
	setErrorHeaderContext(err, "Request", res.Header)
	setErrorRequestContext(err, res.Request)
}

func setErrorRequestContext(err *WrappedError, req *http.Request) {
	err.Set("Endpoint", Config.Endpoint().Url)
	err.Set("URL", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
	setErrorHeaderContext(err, "Response", req.Header)
}

func setErrorHeaderContext(err *WrappedError, prefix string, head http.Header) {
	for key, _ := range head {
		contextKey := fmt.Sprintf("%s:%s", prefix, key)
		if _, skip := hiddenHeaders[key]; skip {
			err.Set(contextKey, "--")
		} else {
			err.Set(contextKey, head.Get(key))
		}
	}
}

type notImplError struct {
	error
}

func (e notImplError) NotImplemented() bool {
	return true
}

func newNotImplError() error {
	return notImplError{errors.New("Not Implemented")}
}

func isNotImplError(err *WrappedError) bool {
	type notimplerror interface {
		NotImplemented() bool
	}
	if e, ok := err.Err.(notimplerror); ok {
		return e.NotImplemented()
	}
	return false
}
