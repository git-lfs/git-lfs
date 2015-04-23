package lfs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/rubyist/tracerx"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

const (
	mediaType = "application/vnd.git-lfs+json; charset-utf-8"
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

type objectResource struct {
	Oid   string                   `json:"oid,omitempty"`
	Size  int64                    `json:"size,omitempty"`
	Links map[string]*linkRelation `json:"_links,omitempty"`
}

func (o *objectResource) NewRequest(relation, method string) (*http.Request, Creds, error) {
	rel, ok := o.Rel(relation)
	if !ok {
		return nil, nil, objectRelationDoesNotExist
	}

	req, creds, err := newClientRequest(method, rel.Href)
	if err != nil {
		return nil, nil, err
	}

	for h, v := range rel.Header {
		req.Header.Set(h, v)
	}

	return req, creds, nil
}

func (o *objectResource) Rel(name string) (*linkRelation, bool) {
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

func Download(oid string) (io.ReadCloser, int64, *WrappedError) {
	req, creds, err := newApiRequest("GET", oid)
	if err != nil {
		return nil, 0, Error(err)
	}

	res, obj, wErr := doApiRequest(req, creds)
	if wErr != nil {
		return nil, 0, wErr
	}

	req, creds, err = obj.NewRequest("download", "GET")
	if err != nil {
		return nil, 0, Error(err)
	}

	res, wErr = doHttpRequest(req, creds)
	if wErr != nil {
		return nil, 0, wErr
	}

	return res.Body, res.ContentLength, nil
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

	reqObj := &objectResource{
		Oid:  oid,
		Size: stat.Size(),
	}

	by, err := json.Marshal(reqObj)
	if err != nil {
		return Error(err)
	}

	req, creds, err := newApiRequest("POST", oid)
	if err != nil {
		return Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = &byteCloser{bytes.NewReader(by)}

	tracerx.Printf("api: uploading %s (%s)", filename, oid)
	res, obj, wErr := doApiRequest(req, creds)
	if wErr != nil {
		return wErr
	}

	if res.StatusCode == 200 {
		return nil
	}

	req, creds, err = obj.NewRequest("upload", "PUT")
	if err != nil {
		return Error(err)
	}

	if len(req.Header.Get("Content-Type")) == 0 {
		req.Header.Set("Content-Type", "application/octet-stream")
	}
	req.Header.Set("Content-Length", strconv.FormatInt(reqObj.Size, 10))
	req.ContentLength = reqObj.Size

	reader := &CallbackReader{
		C:         cb,
		TotalSize: reqObj.Size,
		Reader:    file,
	}

	bar := pb.New64(reqObj.Size)
	bar.SetUnits(pb.U_BYTES)
	bar.Start()

	req.Body = ioutil.NopCloser(bar.NewProxyReader(reader))

	res, wErr = doHttpRequest(req, creds)
	bar.Finish()
	if wErr != nil {
		return wErr
	}

	if res.StatusCode > 299 {
		return Errorf(nil, "Invalid status for %s %s: %d", req.Method, req.URL, res.StatusCode)
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	req, creds, err = obj.NewRequest("verify", "POST")
	if err == objectRelationDoesNotExist {
		return nil
	} else if err != nil {
		return Error(err)
	}

	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Content-Length", strconv.Itoa(len(by)))
	req.ContentLength = int64(len(by))
	req.Body = ioutil.NopCloser(bytes.NewReader(by))
	res, wErr = doHttpRequest(req, creds)

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return wErr
}

func doHttpRequest(req *http.Request, creds Creds) (*http.Response, *WrappedError) {
	res, err := DoHTTP(Config, req)

	var wErr *WrappedError

	if err != nil {
		wErr = Errorf(err, "Error for %s %s", res.Request.Method, res.Request.URL)
	} else {
		if creds != nil {
			saveCredentials(creds, res)
		}

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
		redirectedReq, redirectedCreds, err := newClientRequest(req.Method, res.Header.Get("Location"))
		if err != nil {
			return res, Errorf(err, err.Error())
		}

		via = append(via, req)
		if seeker, ok := req.Body.(io.Seeker); ok {
			_, err := seeker.Seek(0, 0)
			if err != nil {
				return res, Error(err)
			}
			redirectedReq.Body = req.Body
			redirectedReq.ContentLength = req.ContentLength
		} else {
			return res, Errorf(nil, "Request body needs to be an io.Seeker to handle redirects.")
		}

		if err = checkRedirect(redirectedReq, via); err != nil {
			return res, Errorf(err, err.Error())
		}

		return doApiRequestWithRedirects(redirectedReq, redirectedCreds, via)
	}

	return res, wErr
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
	}

	return res, obj, wErr
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
		objectOid = ""
		operation = "upload"
	}

	u, err := ObjectUrl(endpoint, objectOid)
	if err != nil {
		return nil, nil, err
	}

	req, creds, err := newClientRequest(method, u.String())
	if err == nil {
		req.Header.Set("Accept", mediaType)
		if err := mergeSshHeader(req.Header, endpoint, operation, oid); err != nil {
			tracerx.Printf("ssh: attempted with %s.  Error: %s",
				endpoint.SshUserAndHost, err.Error(),
			)
		}
	}
	return req, creds, err
}

func newClientRequest(method, rawurl string) (*http.Request, Creds, error) {
	req, err := http.NewRequest(method, rawurl, nil)
	if err != nil {
		return req, nil, err
	}

	req.Header.Set("User-Agent", UserAgent)
	creds, err := getCreds(req)
	return req, creds, err
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

func setErrorRequestContext(err *WrappedError, req *http.Request) {
	err.Set("Endpoint", Config.Endpoint().Url)
	err.Set("URL", fmt.Sprintf("%s %s", req.Method, req.URL.String()))
	setErrorHeaderContext(err, "Response", req.Header)
}

func setErrorResponseContext(err *WrappedError, res *http.Response) {
	err.Set("Status", res.Status)
	setErrorHeaderContext(err, "Request", res.Header)
	setErrorRequestContext(err, res.Request)
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
