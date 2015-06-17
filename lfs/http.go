package lfs

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

type transferStats struct {
	HeaderSize int
	BodySize   int
	Start      time.Time
	Stop       time.Time
}

type transfer struct {
	requestStats  *transferStats
	responseStats *transferStats
}

var (
	// TODO should use some locks
	transfers           = make(map[*http.Response]*transfer)
	transferBuckets     = make(map[string][]*http.Response)
	transfersLock       sync.Mutex
	transferBucketsLock sync.Mutex
)

func LogTransfer(key string, res *http.Response) {
	if Config.isLoggingStats {
		transferBucketsLock.Lock()
		transferBuckets[key] = append(transferBuckets[key], res)
		transferBucketsLock.Unlock()
	}
}

type HttpClient struct {
	*http.Client
}

func (c *HttpClient) Do(req *http.Request) (*http.Response, error) {
	traceHttpRequest(req)

	crc := countingRequest(req)
	if req.Body != nil {
		// Only set the body if we have a body, but create the countingRequest
		// anyway to make using zeroed stats easier.
		req.Body = crc
	}

	start := time.Now()
	res, err := c.Client.Do(req)
	if err != nil {
		return res, err
	}

	traceHttpResponse(res)

	cresp := countingResponse(res)
	res.Body = cresp

	if Config.isLoggingStats {
		reqHeaderSize := 0
		resHeaderSize := 0

		if dump, err := httputil.DumpRequest(req, false); err == nil {
			reqHeaderSize = len(dump)
		}

		if dump, err := httputil.DumpResponse(res, false); err == nil {
			resHeaderSize = len(dump)
		}

		reqstats := &transferStats{HeaderSize: reqHeaderSize, BodySize: crc.Count}

		// Response body size cannot be figured until it is read. Do not rely on a Content-Length
		// header because it may not exist or be -1 in the case of chunked responses.
		resstats := &transferStats{HeaderSize: resHeaderSize, Start: start}
		t := &transfer{requestStats: reqstats, responseStats: resstats}
		transfersLock.Lock()
		transfers[res] = t
		transfersLock.Unlock()
	}

	return res, err
}

func DoHTTP(req *http.Request) (*http.Response, error) {
	res, err := Config.HttpClient().Do(req)
	if res == nil {
		res = &http.Response{StatusCode: 0, Header: make(http.Header), Request: req}
	}
	return res, err
}

func (c *Configuration) HttpClient() *HttpClient {
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
	if sslVerify == "false" || len(Config.Getenv("GIT_SSL_NO_VERIFY")) > 0 {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	c.httpClient = &HttpClient{
		&http.Client{Transport: tr, CheckRedirect: checkRedirect},
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

func traceHttpRequest(req *http.Request) {
	tracerx.Printf("HTTP: %s %s", req.Method, req.URL.String())

	if Config.isTracingHttp == false {
		return
	}

	dump, err := httputil.DumpRequest(req, false)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(dump))
	for scanner.Scan() {
		fmt.Fprintf(os.Stderr, "> %s\n", scanner.Text())
	}
}

func traceHttpResponse(res *http.Response) {
	if res == nil {
		return
	}

	tracerx.Printf("HTTP: %d", res.StatusCode)

	if Config.isTracingHttp == false {
		return
	}

	dump, err := httputil.DumpResponse(res, false)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(dump))
	for scanner.Scan() {
		fmt.Fprintf(os.Stderr, "< %s\n", scanner.Text())
	}
}

func countingRequest(req *http.Request) *countingReadCloser {
	return &countingReadCloser{request: req, ReadCloser: req.Body}
}

func countingResponse(res *http.Response) *countingReadCloser {
	return &countingReadCloser{response: res, ReadCloser: res.Body}
}

type countingReadCloser struct {
	Count    int
	request  *http.Request
	response *http.Response
	io.ReadCloser
}

func (c *countingReadCloser) Read(b []byte) (int, error) {
	n, err := c.ReadCloser.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	}

	c.Count += n

	if Config.isTracingHttp {
		contentType := ""
		if c.response != nil { // Response, only print certain kinds of data
			contentType = strings.ToLower(strings.SplitN(c.response.Header.Get("Content-Type"), ";", 2)[0])
		} else {
			contentType = strings.ToLower(strings.SplitN(c.request.Header.Get("Content-Type"), ";", 2)[0])
		}

		for _, tracedType := range tracedTypes {
			if strings.Contains(contentType, tracedType) {
				fmt.Fprint(os.Stderr, string(b[0:n]))
			}
		}
	}

	if err == io.EOF && Config.isLoggingStats {
		// This transfer is done, we're checking it this way so we can also
		// catch transfers where the caller forgets to Close() the Body.
		if c.response != nil {
			transfersLock.Lock()
			if transfer, ok := transfers[c.response]; ok {
				transfer.responseStats.BodySize = c.Count
				transfer.responseStats.Stop = time.Now()
			}
			transfersLock.Unlock()
		}
	}
	return n, err
}

// LogHttpStats is intended to be called after all HTTP operations for the
// commmand have finished. It dumps k/v logs, one line per transfer into
// a log file with the current timestamp.
func LogHttpStats() {
	if !Config.isLoggingStats {
		return
	}

	file, err := statsLogFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error logging http stats: %s\n", err)
		return
	}

	fmt.Fprintf(file, "concurrent=%d batch=%v time=%d version=%s\n", Config.ConcurrentTransfers(), Config.BatchTransfer(), time.Now().Unix(), Version)

	for key, responses := range transferBuckets {
		for _, response := range responses {
			stats := transfers[response]
			fmt.Fprintf(file, "key=%s reqheader=%d reqbody=%d resheader=%d resbody=%d restime=%d status=%d url=%s\n",
				key,
				stats.requestStats.HeaderSize,
				stats.requestStats.BodySize,
				stats.responseStats.HeaderSize,
				stats.responseStats.BodySize,
				stats.responseStats.Stop.Sub(stats.responseStats.Start).Nanoseconds(),
				response.StatusCode,
				response.Request.URL)
		}
	}

	fmt.Fprintf(os.Stderr, "HTTP Stats logged to file %s\n", file.Name())
}

func statsLogFile() (*os.File, error) {
	logBase := filepath.Join(LocalLogDir, "http")
	if err := os.MkdirAll(logBase, 0755); err != nil {
		return nil, err
	}

	logFile := fmt.Sprintf("http-%d.log", time.Now().Unix())
	return os.Create(filepath.Join(logBase, logFile))
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

	res, obj, wErr := self.doApiRequest(req, creds)
	if wErr != nil {
		return nil, wErr
	}
	LogTransfer("lfs.api.download", res)

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
	LogTransfer("lfs.data.download", res)

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
		if res != nil {
			switch res.StatusCode {
			case 404, 410:
				tracerx.Printf("api: batch not implemented: %d", res.StatusCode)
				sendApiEvent(apiEventFail)
				return nil, Error(newNotImplError())
			}
		}
		sendApiEvent(apiEventFail)
		return nil, wErr
	}
	LogTransfer("lfs.api.batch", res)

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
	LogTransfer("lfs.api.upload", res)

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
	LogTransfer("lfs.data.upload", res)

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
	LogTransfer("lfs.data.verify", res)

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return wErr
}

func (self *HttpApiContext) doHttpRequest(req *http.Request, creds Creds) (*http.Response, *WrappedError) {
	res, err := DoHTTP(req)

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
		return res, wErr
	}

	if res.StatusCode == 307 {
		redirectedReq, redirectedCreds, err := self.newClientRequest(req.Method, res.Header.Get("Location"))
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

		return self.doApiRequestWithRedirects(redirectedReq, redirectedCreds, via)
	}

	return res, nil
}

func (self *HttpApiContext) doApiRequest(req *http.Request, creds Creds) (*http.Response, *ObjectResource, *WrappedError) {

	via := make([]*http.Request, 0, 4)
	res, wErr := self.doApiRequestWithRedirects(req, creds, via)
	if wErr != nil {
		return res, nil, wErr
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
