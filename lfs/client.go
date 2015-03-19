package lfs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
)

const (
	mediaType = "application/vnd.git-lfs+json; charset-utf-8"
)

var (
	mediaTypeRE                = regexp.MustCompile(`\Aapplication/vnd\.git\-lfs\+json(;|\z)`)
	objectRelationDoesNotExist = errors.New("relation does not exist")
	hiddenHeaders              = map[string]bool{
		"Authorization": true,
	}

	// 401 and 403 print the same default error message
	defaultErrors = map[int]string{
		400: "Client error: %s",
		401: "Authorization error: %s\nCheck that you have proper access to the repository",
		404: "Repository or object not found: %s\nCheck that it exists and that you have proper access to it",
		500: "Server error: %s",
	}
)

type objectResource struct {
	Oid   string                   `json:"oid,omitempty"`
	Size  int64                    `json:"size,omitempty"`
	Links map[string]*linkRelation `json:"_links,omitempty"`
}

func (o *objectResource) NewRequest(relation, method string) (*http.Request, error) {
	rel, ok := o.Rel(relation)
	if !ok {
		return nil, objectRelationDoesNotExist
	}

	req, err := http.NewRequest(method, rel.Href, nil)
	if err != nil {
		return nil, err
	}

	for h, v := range rel.Header {
		req.Header.Set(h, v)
	}

	return req, nil
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
	return nil, 0, nil
}

func Upload(oid, filename string, cb CopyCallback) *WrappedError {
	return nil
}

func doApiRequest(req *http.Request, creds Creds) (*http.Response, *WrappedError) {
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

func handleResponse(res *http.Response) *WrappedError {
	if res.StatusCode < 400 {
		return nil
	}

	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	var wErr *WrappedError

	if mediaTypeRE.MatchString(res.Header.Get("Content-Type")) {
		cliErr := &ClientError{}
		err := json.NewDecoder(res.Body).Decode(cliErr)
		if err != nil {
			return Errorf(err, "Unable to parse HTTP response for %s %s", res.Request.Method, res.Request.URL)
		}

		wErr = Error(cliErr)
	} else {
		wErr = defaultError(res)
	}

	wErr.Panic = res.StatusCode > 499 && res.StatusCode != 501 && res.StatusCode != 509
	return wErr
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
	u, err := Config.ObjectUrl(oid)
	if err != nil {
		return nil, nil, err
	}

	req, creds, err := newClientRequest(method, u.String())
	if err == nil {
		req.Header.Set("Accept", mediaType)
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
	err.Set("Endpoint", Config.Endpoint())
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

func init() {
	defaultErrors[403] = defaultErrors[401]
}
