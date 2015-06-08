package lfs

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

func DoHTTP(c *Configuration, req *http.Request) (*http.Response, error) {
	traceHttpRequest(c, req)
	res, err := c.HttpClient().Do(req)
	if res == nil {
		res = &http.Response{StatusCode: 0, Header: make(http.Header), Request: req}
	}
	traceHttpResponse(c, res)
	return res, err
}

func (c *Configuration) HttpClient() *http.Client {
	if c.httpClient != nil {
		return c.httpClient
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	sslVerify, _ := c.GitConfig("http.sslverify")
	if sslVerify == "false" || len(os.Getenv("GIT_SSL_NO_VERIFY")) > 0 {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	c.httpClient = &http.Client{
		Transport:     tr,
		CheckRedirect: checkRedirect,
	}

	return c.httpClient
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 3 {
		return errors.New("stopped after 3 redirects")
	}

	oldest := via[0]
	for key, _ := range oldest.Header {
		if key == "Authorization" {
			if req.URL.Scheme != oldest.URL.Scheme || req.URL.Host != oldest.URL.Host {
				continue
			}
		}
		req.Header.Set(key, oldest.Header.Get(key))
	}

	tracerx.Printf("api: redirect %s %s to %s", oldest.Method, oldest.URL, req.URL)

	return nil
}

var tracedTypes = []string{"json", "text", "xml", "html"}

func traceHttpRequest(c *Configuration, req *http.Request) {
	tracerx.Printf("HTTP: %s %s", req.Method, req.URL.String())

	if c.isTracingHttp == false {
		return
	}

	if req.Body != nil {
		req.Body = newCountedRequest(req)
	}

	fmt.Fprintf(os.Stderr, "> %s %s %s\n", req.Method, req.URL.RequestURI(), req.Proto)
	for key, _ := range req.Header {
		fmt.Fprintf(os.Stderr, "> %s: %s\n", key, req.Header.Get(key))
	}
}

func traceHttpResponse(c *Configuration, res *http.Response) {
	if res == nil {
		return
	}

	tracerx.Printf("HTTP: %d", res.StatusCode)

	if c.isTracingHttp == false {
		return
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "< %s %s\n", res.Proto, res.Status)
	for key, _ := range res.Header {
		fmt.Fprintf(os.Stderr, "< %s: %s\n", key, res.Header.Get(key))
	}

	traceBody := false
	ctype := strings.ToLower(strings.SplitN(res.Header.Get("Content-Type"), ";", 2)[0])
	for _, tracedType := range tracedTypes {
		if strings.Contains(ctype, tracedType) {
			traceBody = true
		}
	}

	res.Body = newCountedResponse(res)
	if traceBody {
		res.Body = newTracedBody(res.Body)
	}

	fmt.Fprintf(os.Stderr, "\n")
}

const (
	countingUpload = iota
	countingDownload
)

type countingBody struct {
	Direction int
	Size      int64
	io.ReadCloser
}

func (r *countingBody) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.Size += int64(n)
	return n, err
}

func (r *countingBody) Close() error {
	if r.Direction == countingUpload {
		fmt.Fprintf(os.Stderr, "* uploaded %d bytes\n", r.Size)
	} else {
		fmt.Fprintf(os.Stderr, "* downloaded %d bytes\n", r.Size)
	}
	return r.ReadCloser.Close()
}

func newCountedResponse(res *http.Response) *countingBody {
	return &countingBody{countingDownload, 0, res.Body}
}

func newCountedRequest(req *http.Request) *countingBody {
	return &countingBody{countingUpload, 0, req.Body}
}

type tracedBody struct {
	io.ReadCloser
}

func (r *tracedBody) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	fmt.Fprintf(os.Stderr, "%s\n", string(p[0:n]))
	return n, err
}

func (r *tracedBody) Close() error {
	return r.ReadCloser.Close()
}

func newTracedBody(body io.ReadCloser) *tracedBody {
	return &tracedBody{body}
}

// HTTP specific API contect implementation
// HTTP implementation of ApiContext
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

func (self *HttpApiContext) DownloadCheck(oid string) (*ObjectResource, *WrappedError) {
	req, creds, err := self.newApiRequest("GET", oid)
	if err != nil {
		return nil, Error(err)
	}

	_, obj, wErr := self.doApiRequest(req, creds)
	if wErr != nil {
		return nil, wErr
	}

	if !obj.CanDownload() {
		return nil, Error(objectRelationDoesNotExist)
	}

	return obj, nil
}

func (self *HttpApiContext) DownloadObject(obj *ObjectResource) (io.ReadCloser, int64, *WrappedError) {
	req, creds, err := obj.NewRequest(self, "download", "GET")
	if err != nil {
		return nil, 0, Error(err)
	}

	res, wErr := self.doHttpRequest(req, creds)
	if wErr != nil {
		return nil, 0, wErr
	}

	return res.Body, res.ContentLength, nil
}

func (self *HttpApiContext) Batch(objects []*ObjectResource) ([]*ObjectResource, *WrappedError) {
	if len(objects) == 0 {
		return nil, nil
	}

	o := map[string][]*ObjectResource{"objects": objects}

	by, err := json.Marshal(o)
	if err != nil {
		return nil, Error(err)
	}

	req, creds, err := self.newApiRequest("POST", "batch")
	if err != nil {
		return nil, Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = &byteCloser{bytes.NewReader(by)}

	tracerx.Printf("api: batch %d files", len(objects))
	res, objs, wErr := self.doApiBatchRequest(req, creds)
	if wErr != nil {
		sendApiEvent(apiEventFail)
		return nil, wErr
	}

	sendApiEvent(apiEventSuccess)
	if res.StatusCode != 200 {
		return nil, Errorf(nil, "Invalid status for %s %s: %d", req.Method, req.URL, res.StatusCode)
	}

	return objs, nil
}

func (self *HttpApiContext) UploadCheck(oid string, sz int64) (*ObjectResource, *WrappedError) {

	reqObj := &ObjectResource{
		Oid:  oid,
		Size: sz,
	}

	by, err := json.Marshal(reqObj)
	if err != nil {
		sendApiEvent(apiEventFail)
		return nil, Error(err)
	}

	req, creds, err := self.newApiRequest("POST", oid)
	if err != nil {
		sendApiEvent(apiEventFail)
		return nil, Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = &byteCloser{bytes.NewReader(by)}

	tracerx.Printf("api: uploading (%s)", oid)
	res, obj, wErr := self.doApiRequest(req, creds)
	if wErr != nil {
		sendApiEvent(apiEventFail)
		return nil, wErr
	}

	sendApiEvent(apiEventSuccess)

	if res.StatusCode == 200 {
		return nil, nil
	}

	return obj, nil
}

func (self *HttpApiContext) UploadObject(o *ObjectResource, reader io.Reader) *WrappedError {
	req, creds, err := o.NewRequest(self, "upload", "PUT")
	if err != nil {
		return Error(err)
	}

	if len(req.Header.Get("Content-Type")) == 0 {
		req.Header.Set("Content-Type", "application/octet-stream")
	}
	req.Header.Set("Content-Length", strconv.FormatInt(o.Size, 10))
	req.ContentLength = o.Size

	req.Body = ioutil.NopCloser(reader)

	res, wErr := self.doHttpRequest(req, creds)
	if wErr != nil {
		return wErr
	}

	if res.StatusCode > 299 {
		return Errorf(nil, "Invalid status for %s %s: %d", req.Method, req.URL, res.StatusCode)
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	req, creds, err = o.NewRequest(self, "verify", "POST")
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
	res, wErr = self.doHttpRequest(req, creds)

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return wErr
}

func (self *HttpApiContext) doHttpRequest(req *http.Request, creds Creds) (*http.Response, *WrappedError) {
	res, err := DoHTTP(Config, req)

	var wErr *WrappedError

	if err != nil {
		wErr = Errorf(err, "Error for %s %s", res.Request.Method, res.Request.URL)
	} else {
		if creds != nil {
			saveCredentials(creds, res)
		}

		wErr = self.handleResponse(res)
	}

	if wErr != nil {
		if res != nil {
			self.setErrorResponseContext(wErr, res)
		} else {
			self.setErrorRequestContext(wErr, req)
		}
	}

	return res, wErr
}

func (self *HttpApiContext) doApiRequestWithRedirects(req *http.Request, creds Creds, via []*http.Request) (*http.Response, *WrappedError) {
	res, wErr := self.doHttpRequest(req, creds)
	if wErr != nil {
		return nil, wErr
	}

	if res.StatusCode == 307 {
		redirectedReq, redirectedCreds, err := self.newClientRequest(req.Method, res.Header.Get("Location"))
		if err != nil {
			return nil, Errorf(err, err.Error())
		}

		via = append(via, req)
		seeker, ok := req.Body.(io.Seeker)
		if !ok {
			return nil, Errorf(nil, "Request body needs to be an io.Seeker to handle redirects.")
		}

		if _, err := seeker.Seek(0, 0); err != nil {
			return nil, Error(err)
		}
		redirectedReq.Body = req.Body
		redirectedReq.ContentLength = req.ContentLength

		if err = checkRedirect(redirectedReq, via); err != nil {
			return nil, Errorf(err, err.Error())
		}

		return self.doApiRequestWithRedirects(redirectedReq, redirectedCreds, via)
	}

	return res, nil
}

func (self *HttpApiContext) doApiRequest(req *http.Request, creds Creds) (*http.Response, *ObjectResource, *WrappedError) {

	via := make([]*http.Request, 0, 4)
	res, wErr := self.doApiRequestWithRedirects(req, creds, via)
	if wErr != nil {
		return nil, nil, wErr
	}

	obj := &ObjectResource{}
	wErr = self.decodeApiResponse(res, obj)

	if wErr != nil {
		self.setErrorResponseContext(wErr, res)
		return nil, nil, wErr
	}

	return res, obj, nil
}

func (self *HttpApiContext) doApiBatchRequest(req *http.Request, creds Creds) (*http.Response, []*ObjectResource, *WrappedError) {
	via := make([]*http.Request, 0, 4)
	res, wErr := self.doApiRequestWithRedirects(req, creds, via)
	if wErr != nil {
		return res, nil, wErr
	}

	var objs map[string][]*ObjectResource
	wErr = self.decodeApiResponse(res, &objs)

	if wErr != nil {
		self.setErrorResponseContext(wErr, res)
	}

	return res, objs["objects"], wErr
}

func (self *HttpApiContext) handleResponse(res *http.Response) *WrappedError {
	if res.StatusCode < 400 {
		return nil
	}

	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	cliErr := &ClientError{}
	wErr := self.decodeApiResponse(res, cliErr)
	if wErr == nil {
		if len(cliErr.Message) == 0 {
			wErr = self.defaultError(res)
		} else {
			wErr = Error(cliErr)
		}
	}

	wErr.Panic = res.StatusCode > 499 && res.StatusCode != 501 && res.StatusCode != 509
	return wErr
}

func (self *HttpApiContext) decodeApiResponse(res *http.Response, obj interface{}) *WrappedError {
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

func (self *HttpApiContext) defaultError(res *http.Response) *WrappedError {
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

func (self *HttpApiContext) newApiRequest(method, oid string) (*http.Request, Creds, error) {
	endpoint := self.Endpoint()
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

	req, creds, err := self.newClientRequest(method, u.String())
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

func (self *HttpApiContext) newClientRequest(method, rawurl string) (*http.Request, Creds, error) {
	req, err := http.NewRequest(method, rawurl, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("User-Agent", UserAgent)
	creds, err := getCreds(req)
	if err != nil {
		return nil, nil, err
	}

	return req, creds, nil
}

func (self *HttpApiContext) setErrorResponseContext(err *WrappedError, res *http.Response) {
	err.Set("Status", res.Status)
	self.setErrorHeaderContext(err, "Request", res.Header)
	self.setErrorRequestContext(err, res.Request)
}

func (self *HttpApiContext) setErrorRequestContext(err *WrappedError, req *http.Request) {
	err.Set("Endpoint", Config.Endpoint().Url)
	err.Set("URL", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
	self.setErrorHeaderContext(err, "Response", req.Header)
}

func (self *HttpApiContext) setErrorHeaderContext(err *WrappedError, prefix string, head http.Header) {
	for key, _ := range head {
		contextKey := fmt.Sprintf("%s:%s", prefix, key)
		if _, skip := hiddenHeaders[key]; skip {
			err.Set(contextKey, "--")
		} else {
			err.Set(contextKey, head.Get(key))
		}
	}
}
