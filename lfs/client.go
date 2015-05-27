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
	"os"
	"path/filepath"
	"regexp"

	"github.com/rubyist/tracerx"
)

// SJS MOVE LATER move all this to http specific file
// not moved for now to make merging easier
const (
	mediaType = "application/vnd.git-lfs+json; charset=utf-8"
)

// The apiEvent* statuses (and the apiEvent channel) are used by
// UploadQueue to know when it is OK to process uploads concurrently.
const (
	apiEventFail = iota
	apiEventSuccess
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

	apiEvent = make(chan int)
)

type httpObjectResource struct {
	Oid   string                   `json:"oid,omitempty"`
	Size  int64                    `json:"size,omitempty"`
	Links map[string]*linkRelation `json:"_links,omitempty"`
}

func (o *httpObjectResource) NewRequest(ctx *HttpApiContext, relation, method string) (*http.Request, Creds, error) {
	rel, ok := o.Rel(relation)
	if !ok {
		return nil, nil, objectRelationDoesNotExist
	}

	req, creds, err := ctx.newClientRequest(method, rel.Href)
	if err != nil {
		return nil, nil, err
	}

	for h, v := range rel.Header {
		req.Header.Set(h, v)
	}

	return req, creds, nil
}

func (o *httpObjectResource) Rel(name string) (*linkRelation, bool) {
	if o.Links == nil {
		return nil, false
	}

	rel, ok := o.Links[name]
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

// SJS MOVE LATER: end section that needs relocating to http-specific source file

func Download(oid string) (io.ReadCloser, int64, *WrappedError) {

	ctx := GetApiContext(Config.Endpoint())
	return ctx.Download(oid)
}

type byteCloser struct {
	*bytes.Reader
}

func (b *byteCloser) Close() error {
	return nil
}

func Upload(oidPath, filename string, cb CopyCallback) *WrappedError {
	oid := filepath.Base(oidPath)
	file, err := os.Open(oidPath)
	if err != nil {
		return Error(err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return Error(err)
	}

	tracerx.Printf("api: uploading %s (%s)", filename, oid)
	ctx := GetApiContext(Config.Endpoint())
	return ctx.Upload(oid, stat.Size(), file, cb)

}

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
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

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
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

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
func (self *HttpApiContext) doApiRequest(req *http.Request, creds Creds) (*http.Response, *httpObjectResource, *WrappedError) {

	via := make([]*http.Request, 0, 4)
	res, wErr := self.doApiRequestWithRedirects(req, creds, via)
	if wErr != nil {
		return nil, nil, wErr
	}

	obj := &httpObjectResource{}
	wErr = self.decodeApiResponse(res, obj)

	if wErr != nil {
		self.setErrorResponseContext(wErr, res)
		return nil, nil, wErr
	}

	return res, obj, nil
}

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
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

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
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

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
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

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
func (self *HttpApiContext) newApiRequest(method, oid string) (*http.Request, Creds, error) {
	endpoint := self.Endpoint()
	objectOid := oid
	operation := "download"
	if method == "POST" {
		objectOid = ""
		operation = "upload"
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

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
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

func getCreds(req *http.Request) (Creds, error) {
	if len(req.Header.Get("Authorization")) > 0 {
		return nil, nil
	}

	apiUrl, err := Config.ObjectUrl("")
	if err != nil {
		return nil, err
	}

	if req.URL.Scheme == apiUrl.Scheme &&
		req.URL.Host == apiUrl.Host {
		creds, err := credentials(req.URL)
		if err != nil {
			return nil, err
		}

		token := fmt.Sprintf("%s:%s", creds["username"], creds["password"])
		auth := "Basic " + base64.URLEncoding.EncodeToString([]byte(token))
		req.Header.Set("Authorization", auth)
		return creds, nil
	}

	return nil, nil
}

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
func (self *HttpApiContext) setErrorResponseContext(err *WrappedError, res *http.Response) {
	err.Set("Status", res.Status)
	self.setErrorHeaderContext(err, "Request", res.Header)
	self.setErrorRequestContext(err, res.Request)
}

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
func (self *HttpApiContext) setErrorRequestContext(err *WrappedError, req *http.Request) {
	err.Set("Endpoint", Config.Endpoint().Url)
	err.Set("URL", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
	self.setErrorHeaderContext(err, "Response", req.Header)
}

// SJS MOVE LATER: In the interests of easier merging, this method left in this source file
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

func sendApiEvent(event int) {
	select {
	case apiEvent <- event:
	default:
	}
}
