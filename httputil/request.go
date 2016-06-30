package httputil

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/github/git-lfs/auth"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"

	"github.com/rubyist/tracerx"
)

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

// Internal http request management
func doHttpRequest(req *http.Request, creds auth.Creds) (*http.Response, error) {
	var (
		res *http.Response
		err error
	)

	if config.Config.NtlmAccess(auth.GetOperationForRequest(req)) {
		res, err = doNTLMRequest(req, true)
	} else {
		res, err = NewHttpClient(config.Config, req.Host).Do(req)
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
		if errutil.IsAuthError(err) {
			authType := GetAuthType(res)
			oldAuthType := SetAuthType(req, authType)
			if authType != oldAuthType {
				tracerx.Printf("api: http response indicates %q authentication. (old was %q) Resubmitting...", authType, oldAuthType)
				doHttpRequest(req, creds)
			} else {
				err = errutil.Error(err)
			}
		} else {
			err = errutil.Error(err)
		}
	} else {
		err = handleResponse(res, creds)
	}

	if err != nil {
		if res != nil {
			SetErrorResponseContext(err, res)
		} else {
			setErrorRequestContext(err, req)
		}
	}

	return res, err
}

// DoHttpRequest performs a single HTTP request
func DoHttpRequest(req *http.Request, useCreds bool) (*http.Response, error) {
	var creds auth.Creds
	if useCreds {
		c, err := auth.GetCreds(req)
		if err != nil {
			return nil, err
		}
		creds = c
	}

	return doHttpRequest(req, creds)
}

// DoHttpRequestWithRedirects runs a HTTP request and responds to redirects
func DoHttpRequestWithRedirects(req *http.Request, via []*http.Request, useCreds bool) (*http.Response, error) {
	var creds auth.Creds
	if useCreds {
		c, err := auth.GetCreds(req)
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

		redirectedReq, err := NewHttpRequest(req.Method, redirectTo, nil)
		if err != nil {
			return res, errutil.Errorf(err, err.Error())
		}

		via = append(via, req)

		// Avoid seeking and re-wrapping the CountingReadCloser, just get the "real" body
		realBody := req.Body
		if wrappedBody, ok := req.Body.(*CountingReadCloser); ok {
			realBody = wrappedBody.ReadCloser
		}

		seeker, ok := realBody.(io.Seeker)
		if !ok {
			return res, errutil.Errorf(nil, "Request body needs to be an io.Seeker to handle redirects.")
		}

		if _, err := seeker.Seek(0, 0); err != nil {
			return res, errutil.Error(err)
		}
		redirectedReq.Body = realBody
		redirectedReq.ContentLength = req.ContentLength

		if err = CheckRedirect(redirectedReq, via); err != nil {
			return res, errutil.Errorf(err, err.Error())
		}

		return DoHttpRequestWithRedirects(redirectedReq, via, useCreds)
	}

	return res, nil
}

// NewHttpRequest creates a template request, with the given headers & UserAgent supplied
func NewHttpRequest(method, rawurl string, header map[string]string) (*http.Request, error) {
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

func SetAuthType(req *http.Request, authType string) string {
	operation := auth.GetOperationForRequest(req)
	oldAuthType := config.Config.Access(operation)
	config.Config.SetAccess(operation, authType)
	return oldAuthType
}

func GetAuthType(res *http.Response) string {
	auth := res.Header.Get("Www-Authenticate")
	if len(auth) < 1 {
		auth = res.Header.Get("Lfs-Authenticate")
	}

	if strings.HasPrefix(strings.ToLower(auth), "ntlm") {
		return "ntlm"
	}

	return "basic"
}
