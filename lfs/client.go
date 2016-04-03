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
	"strings"

	"github.com/github/git-lfs/git"
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

type ObjectError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ObjectError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

type ObjectResource struct {
	Oid     string                   `json:"oid,omitempty"`
	Size    int64                    `json:"size"`
	Actions map[string]*linkRelation `json:"actions,omitempty"`
	Links   map[string]*linkRelation `json:"_links,omitempty"`
	Error   *ObjectError             `json:"error,omitempty"`
}

func (o *ObjectResource) NewRequest(relation, method string) (*http.Request, error) {
	rel, ok := o.Rel(relation)
	if !ok {
		if relation == "download" {
			return nil, errors.New("Object not found on the server.")
		}
		return nil, fmt.Errorf("No %q action for this object.", relation)

	}

	req, err := newClientRequest(method, rel.Href, rel.Header)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (o *ObjectResource) Rel(name string) (*linkRelation, bool) {
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

// Download will attempt to download the object with the given oid. The batched
// API will be used, but if the server does not implement the batch operations
// it will fall back to the legacy API.
func Download(oid string, size int64) (io.ReadCloser, int64, error) {
	if !Config.BatchTransfer() {
		return DownloadLegacy(oid)
	}

	objects := []*ObjectResource{
		&ObjectResource{Oid: oid, Size: size},
	}

	objs, err := downloadBatch(objects)
	if err != nil {
		if IsNotImplementedError(err) {
			git.Config.SetLocal("", "lfs.batch", "false")
			return DownloadLegacy(oid)
		}
		return nil, 0, err
	}

	if len(objs) != 1 { // Expecting to find one object
		return nil, 0, Error(fmt.Errorf("Object not found: %s", oid))
	}

	return DownloadObject(objs[0])
}

// DownloadLegacy attempts to download the object for the given oid using the
// legacy API.
func DownloadLegacy(oid string) (io.ReadCloser, int64, error) {
	req, err := newApiRequest("GET", oid)
	if err != nil {
		return nil, 0, Error(err)
	}

	res, obj, err := doLegacyApiRequest(req)
	if err != nil {
		return nil, 0, err
	}
	LogTransfer("lfs.api.download", res)
	req, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, 0, Error(err)
	}

	res, err = doStorageRequest(req)
	if err != nil {
		return nil, 0, err
	}
	LogTransfer("lfs.data.download", res)

	return res.Body, res.ContentLength, nil
}

type byteCloser struct {
	*bytes.Reader
}

func DownloadCheck(oid string) (*ObjectResource, error) {
	req, err := newApiRequest("GET", oid)
	if err != nil {
		return nil, Error(err)
	}

	res, obj, err := doLegacyApiRequest(req)
	if err != nil {
		return nil, err
	}
	LogTransfer("lfs.api.download", res)

	_, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, Error(err)
	}

	return obj, nil
}

func DownloadObject(obj *ObjectResource) (io.ReadCloser, int64, error) {
	req, err := obj.NewRequest("download", "GET")
	if err != nil {
		return nil, 0, Error(err)
	}

	res, err := doStorageRequest(req)
	if err != nil {
		return nil, 0, newRetriableError(err)
	}
	LogTransfer("lfs.data.download", res)

	return res.Body, res.ContentLength, nil
}

func (b *byteCloser) Close() error {
	return nil
}

func downloadBatch(objects []*ObjectResource) ([]*ObjectResource, error) {
	return Batch(objects, "download", nil)
}

func Batch(objects []*ObjectResource, operation string, metadata *TransferMetadata) ([]*ObjectResource, error) {
	if len(objects) == 0 {
		return nil, nil
	}

	o := map[string]interface{}{"objects": objects, "operation": operation}
	if metadata != nil {
		m := map[string]interface{}{}
		ref := normalizeRef(metadata.ref)
		if len(ref) > 0 {
			m["ref"] = ref
		}
		if len(m) > 0 {
			o["meta"] = m
		}
	}

	by, err := json.Marshal(o)
	if err != nil {
		return nil, Error(err)
	}

	req, err := newBatchApiRequest(operation)
	if err != nil {
		return nil, Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = &byteCloser{bytes.NewReader(by)}

	tracerx.Printf("api: batch %d files", len(objects))

	res, objs, err := doApiBatchRequest(req)

	if err != nil {

		if res == nil {
			return nil, newRetriableError(err)
		}

		if res.StatusCode == 0 {
			return nil, newRetriableError(err)
		}

		if IsAuthError(err) {
			setAuthType(req, res)
			return Batch(objects, operation, metadata)
		}

		switch res.StatusCode {
		case 404, 410:
			tracerx.Printf("api: batch not implemented: %d", res.StatusCode)
			return nil, newNotImplementedError(nil)
		}

		tracerx.Printf("api error: %s", err)
		return nil, Error(err)
	}
	LogTransfer("lfs.api.batch", res)

	if res.StatusCode != 200 {
		return nil, Error(fmt.Errorf("Invalid status for %s: %d", traceHttpReq(req), res.StatusCode))
	}

	return objs, nil
}

func UploadCheck(oidPath string) (*ObjectResource, error) {
	oid := filepath.Base(oidPath)

	stat, err := os.Stat(oidPath)
	if err != nil {
		return nil, Error(err)
	}

	reqObj := &ObjectResource{
		Oid:  oid,
		Size: stat.Size(),
	}

	by, err := json.Marshal(reqObj)
	if err != nil {
		return nil, Error(err)
	}

	req, err := newApiRequest("POST", oid)
	if err != nil {
		return nil, Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = &byteCloser{bytes.NewReader(by)}

	tracerx.Printf("api: uploading (%s)", oid)
	res, obj, err := doLegacyApiRequest(req)

	if err != nil {
		if IsAuthError(err) {
			setAuthType(req, res)
			return UploadCheck(oidPath)
		}

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

func UploadObject(o *ObjectResource, cb CopyCallback) error {
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

	req, err := o.NewRequest("upload", "PUT")
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

	res, err := doStorageRequest(req)
	if err != nil {
		return newRetriableError(err)
	}
	LogTransfer("lfs.data.upload", res)

	// A status code of 403 likely means that an authentication token for the
	// upload has expired. This can be safely retried.
	if res.StatusCode == 403 {
		return newRetriableError(err)
	}

	if res.StatusCode > 299 {
		return Errorf(nil, "Invalid status for %s: %d", traceHttpReq(req), res.StatusCode)
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	if _, ok := o.Rel("verify"); !ok {
		return nil
	}

	req, err = o.NewRequest("verify", "POST")
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
	res, err = doAPIRequest(req, true)
	if err != nil {
		return err
	}

	LogTransfer("lfs.data.verify", res)
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return err
}

// doLegacyApiRequest runs the request to the LFS legacy API.
func doLegacyApiRequest(req *http.Request) (*http.Response, *ObjectResource, error) {
	via := make([]*http.Request, 0, 4)
	res, err := doApiRequestWithRedirects(req, via, true)
	if err != nil {
		return res, nil, err
	}

	obj := &ObjectResource{}
	err = decodeApiResponse(res, obj)

	if err != nil {
		setErrorResponseContext(err, res)
		return nil, nil, err
	}

	return res, obj, nil
}

// getOperationForHttpRequest determines the operation type for a http.Request
func getOperationForHttpRequest(req *http.Request) string {
	operation := "download"
	if req.Method == "POST" || req.Method == "PUT" {
		operation = "upload"
	}
	return operation
}

// doApiBatchRequest runs the request to the LFS batch API. If the API returns a
// 401, the repo will be marked as having private access and the request will be
// re-run. When the repo is marked as having private access, credentials will
// be retrieved.
func doApiBatchRequest(req *http.Request) (*http.Response, []*ObjectResource, error) {
	res, err := doAPIRequest(req, Config.PrivateAccess(getOperationForHttpRequest(req)))

	if err != nil {
		if res != nil && res.StatusCode == 401 {
			return res, nil, newAuthError(err)
		}
		return res, nil, err
	}

	var objs map[string][]*ObjectResource
	err = decodeApiResponse(res, &objs)

	if err != nil {
		setErrorResponseContext(err, res)
	}

	return res, objs["objects"], err
}

// doStorageREquest runs the request to the storage API from a link provided by
// the "actions" or "_links" properties an LFS API response.
func doStorageRequest(req *http.Request) (*http.Response, error) {
	creds, err := getCreds(req)
	if err != nil {
		return nil, err
	}

	return doHttpRequest(req, creds)
}

// doAPIRequest runs the request to the LFS API, without parsing the response
// body. If the API returns a 401, the repo will be marked as having private
// access and the request will be re-run. When the repo is marked as having
// private access, credentials will be retrieved.
func doAPIRequest(req *http.Request, useCreds bool) (*http.Response, error) {
	via := make([]*http.Request, 0, 4)
	return doApiRequestWithRedirects(req, via, useCreds)
}

// doHttpRequest runs the given HTTP request. LFS or Storage API requests should
// use doApiBatchRequest() or doStorageRequest() instead.
func doHttpRequest(req *http.Request, creds Creds) (*http.Response, error) {
	var (
		res *http.Response
		err error
	)

	if Config.NtlmAccess(getOperationForHttpRequest(req)) {
		res, err = DoNTLMRequest(req, true)
	} else {
		res, err = Config.HttpClient(req.Host).Do(req)
	}

	if res == nil {
		res = &http.Response{
			StatusCode: 0,
			Header:     make(http.Header),
			Request:    req,
			Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		}
	}

	if err != nil {
		if IsAuthError(err) {
			setAuthType(req, res)
			doHttpRequest(req, creds)
		} else {
			err = Error(err)
		}
	} else {
		err = handleResponse(res, creds)
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

func doApiRequestWithRedirects(req *http.Request, via []*http.Request, useCreds bool) (*http.Response, error) {
	var creds Creds
	if useCreds {
		c, err := getCredsForAPI(req)
		if err != nil {
			return nil, err
		}
		creds = c
	}

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

		redirectedReq, err := newClientRequest(req.Method, redirectTo, nil)
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

		return doApiRequestWithRedirects(redirectedReq, via, useCreds)
	}

	return res, nil
}

func handleResponse(res *http.Response, creds Creds) error {
	saveCredentials(creds, res)

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

	if res.StatusCode == 401 {
		return newAuthError(err)
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
		return Errorf(err, "Unable to parse HTTP response for %s", traceHttpReq(res.Request))
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

func newApiRequest(method, oid string) (*http.Request, error) {
	objectOid := oid
	operation := "download"
	if method == "POST" {
		if oid != "batch" {
			objectOid = ""
			operation = "upload"
		}
	}
	endpoint := Config.Endpoint(operation)

	res, err := sshAuthenticate(endpoint, operation, oid)
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

	req, err := newClientRequest(method, u.String(), res.Header)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", mediaType)
	return req, nil
}

func newClientRequest(method, rawurl string, header map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, rawurl, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	req.Header.Set("User-Agent", UserAgent)

	return req, nil
}

func newBatchApiRequest(operation string) (*http.Request, error) {
	endpoint := Config.Endpoint(operation)

	res, err := sshAuthenticate(endpoint, operation, "")
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

	req, err := newBatchClientRequest("POST", u.String())
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", mediaType)
	if res.Header != nil {
		for key, value := range res.Header {
			req.Header.Set(key, value)
		}
	}

	return req, nil
}

func newBatchClientRequest(method, rawurl string) (*http.Request, error) {
	req, err := http.NewRequest(method, rawurl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	return req, nil
}

func setRequestAuthFromUrl(req *http.Request, u *url.URL) bool {
	if !Config.NtlmAccess(getOperationForHttpRequest(req)) && u.User != nil {
		if pass, ok := u.User.Password(); ok {
			fmt.Fprintln(os.Stderr, "warning: current Git remote contains credentials")
			setRequestAuth(req, u.User.Username(), pass)
			return true
		}
	}

	return false
}

func setAuthType(req *http.Request, res *http.Response) {
	authType := getAuthType(res)
	operation := getOperationForHttpRequest(req)
	Config.SetAccess(operation, authType)
	tracerx.Printf("api: http response indicates %q authentication. Resubmitting...", authType)
}

func getAuthType(res *http.Response) string {
	auth := res.Header.Get("Www-Authenticate")
	if len(auth) < 1 {
		auth = res.Header.Get("Lfs-Authenticate")
	}

	if strings.HasPrefix(strings.ToLower(auth), "ntlm") {
		return "ntlm"
	}

	return "basic"
}

func setRequestAuth(req *http.Request, user, pass string) {
	if Config.NtlmAccess(getOperationForHttpRequest(req)) {
		return
	}

	if len(user) == 0 && len(pass) == 0 {
		return
	}

	token := fmt.Sprintf("%s:%s", user, pass)
	auth := "Basic " + strings.TrimSpace(base64.StdEncoding.EncodeToString([]byte(token)))
	req.Header.Set("Authorization", auth)
}

func setErrorResponseContext(err error, res *http.Response) {
	ErrorSetContext(err, "Status", res.Status)
	setErrorHeaderContext(err, "Request", res.Header)
	setErrorRequestContext(err, res.Request)
}

func setErrorRequestContext(err error, req *http.Request) {
	ErrorSetContext(err, "Endpoint", Config.Endpoint(getOperationForHttpRequest(req)).Url)
	ErrorSetContext(err, "URL", traceHttpReq(req))
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

func normalizeRef(ref string) string {
	if len(ref) == 0 || ref == "HEAD" {
		return ""
	}
	if strings.HasPrefix("refs/heads/", ref) {
		return ref
	}
	return "refs/heads/" + ref
}
