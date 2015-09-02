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
	lfsMediaTypeRE  = regexp.MustCompile(`\Aapplication/vnd\.git\-lfs\+json(;|\z)`)
	jsonMediaTypeRE = regexp.MustCompile(`\Aapplication/json(;|\z)`)
	hiddenHeaders   = map[string]bool{
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
		return nil, nil, errors.New("relation does not exist")
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

func Download(oid string) (io.ReadCloser, int64, error) {
	req, creds, err := newApiRequest("GET", oid)
	if err != nil {
		return nil, 0, Error(err)
	}

	res, obj, err := doApiRequest(req, creds)
	if err != nil {
		return nil, 0, err
	}
	LogTransfer("lfs.api.download", res)

	req, creds, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, 0, Error(err)
	}

	res, err = doHttpRequest(req, creds)
	if err != nil {
		return nil, 0, err
	}
	LogTransfer("lfs.data.download", res)

	return res.Body, res.ContentLength, nil
}

type byteCloser struct {
	*bytes.Reader
}

func DownloadCheck(oid string) (*objectResource, error) {
	req, creds, err := newApiRequest("GET", oid)
	if err != nil {
		return nil, Error(err)
	}

	res, obj, err := doApiRequest(req, creds)
	if err != nil {
		return nil, err
	}
	LogTransfer("lfs.api.download", res)

	_, _, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, Error(err)
	}

	return obj, nil
}

func DownloadObject(obj *objectResource) (io.ReadCloser, int64, error) {
	req, creds, err := obj.NewRequest("download", "GET")
	if err != nil {
		return nil, 0, Error(err)
	}

	res, err := doHttpRequest(req, creds)
	if err != nil {
		return nil, 0, newRetriableError(err)
	}
	LogTransfer("lfs.data.download", res)

	return res.Body, res.ContentLength, nil
}

func (b *byteCloser) Close() error {
	return nil
}

func Batch(objects []*objectResource, operation string) ([]*objectResource, error) {
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
	res, objs, err := doApiBatchRequest(req, creds)
	if err != nil {
		if res == nil {
			return nil, err
		}

		switch res.StatusCode {
		case 401:
			Config.SetPrivateAccess()
			tracerx.Printf("api: batch not authorized, submitting with auth")
			return Batch(objects, operation)
		case 404, 410:
			tracerx.Printf("api: batch not implemented: %d", res.StatusCode)
			return nil, newNotImplementedError(nil)
		}

		tracerx.Printf("api error: %s", err)
		return nil, newRetriableError(err)
	}
	LogTransfer("lfs.api.batch", res)

	if res.StatusCode != 200 {
		return nil, Error(fmt.Errorf("Invalid status for %s %s: %d", req.Method, req.URL, res.StatusCode))
	}

	return objs, nil
}

func UploadCheck(oidPath string) (*objectResource, error) {
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
	res, obj, err := doApiRequest(req, creds)
	if err != nil {
		return nil, newRetriableError(err)
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

func UploadObject(o *objectResource, cb CopyCallback) error {
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

	res, err := doHttpRequest(req, creds)
	if err != nil {
		return newRetriableError(err)
	}
	LogTransfer("lfs.data.upload", res)

	if res.StatusCode > 299 {
		return Errorf(nil, "Invalid status for %s %s: %d", req.Method, req.URL, res.StatusCode)
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	if _, ok := o.Rel("verify"); !ok {
		return nil
	}

	req, creds, err = o.NewRequest("verify", "POST")
	if err != nil {
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
	res, err = doHttpRequest(req, creds)
	if err != nil {
		return err
	}

	LogTransfer("lfs.data.verify", res)
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return err
}

func doHttpRequest(req *http.Request, creds Creds) (*http.Response, error) {
	res, err := Config.HttpClient().Do(req)
	if res == nil {
		res = &http.Response{
			StatusCode: 0,
			Header:     make(http.Header),
			Request:    req,
			Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		}
	}

	if err != nil {
		err = Errorf(err, "Error for %s %s", res.Request.Method, res.Request.URL)
	} else {
		saveCredentials(creds, res)
		err = handleResponse(res)
	}

	if err != nil {
		if res != nil {
			setErrorResponseContext(err, res)
		} else {
			setErrorRequestContext(err, req)
		}
	}

	return res, err
}

func doApiRequestWithRedirects(req *http.Request, creds Creds, via []*http.Request) (*http.Response, error) {
	res, err := doHttpRequest(req, creds)
	if err != nil {
		return res, err
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

func doApiRequest(req *http.Request, creds Creds) (*http.Response, *objectResource, error) {
	via := make([]*http.Request, 0, 4)
	res, err := doApiRequestWithRedirects(req, creds, via)
	if err != nil {
		return res, nil, err
	}

	obj := &objectResource{}
	err = decodeApiResponse(res, obj)

	if err != nil {
		setErrorResponseContext(err, res)
		return nil, nil, err
	}

	return res, obj, nil
}

// doApiBatchRequest runs the request to the batch API. If the API returns a 401,
// the repo will be marked as having private access and the request will be
// re-run. When the repo is marked as having private access, credentials will
// be retrieved.
func doApiBatchRequest(req *http.Request, creds Creds) (*http.Response, []*objectResource, error) {
	via := make([]*http.Request, 0, 4)
	res, err := doApiRequestWithRedirects(req, creds, via)

	if err != nil {
		return res, nil, err
	}

	var objs map[string][]*objectResource
	err = decodeApiResponse(res, &objs)

	if err != nil {
		setErrorResponseContext(err, res)
	}

	return res, objs["objects"], err
}

func handleResponse(res *http.Response) error {
	if res.StatusCode < 400 {
		return nil
	}

	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	cliErr := &ClientError{}
	err := decodeApiResponse(res, cliErr)
	if err == nil {
		if len(cliErr.Message) == 0 {
			err = defaultError(res)
		} else {
			err = Error(cliErr)
		}
	}

	if res.StatusCode > 499 && res.StatusCode != 501 && res.StatusCode != 509 {
		return newFatalError(err)
	}

	return err
}

func decodeApiResponse(res *http.Response, obj interface{}) error {
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

func defaultError(res *http.Response) error {
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

func setErrorResponseContext(err error, res *http.Response) {
	ErrorSetContext(err, "Status", res.Status)
	setErrorHeaderContext(err, "Request", res.Header)
	setErrorRequestContext(err, res.Request)
}

func setErrorRequestContext(err error, req *http.Request) {
	ErrorSetContext(err, "Endpoint", Config.Endpoint().Url)
	ErrorSetContext(err, "URL", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
	setErrorHeaderContext(err, "Response", req.Header)
}

func setErrorHeaderContext(err error, prefix string, head http.Header) {
	for key, _ := range head {
		contextKey := fmt.Sprintf("%s:%s", prefix, key)
		if _, skip := hiddenHeaders[key]; skip {
			ErrorSetContext(err, contextKey, "--")
		} else {
			ErrorSetContext(err, contextKey, head.Get(key))
		}
	}
}
