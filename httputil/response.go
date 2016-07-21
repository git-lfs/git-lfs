package httputil

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/github/git-lfs/auth"
	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
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

// DecodeResponse attempts to decode the contents of the response as a JSON object
func DecodeResponse(res *http.Response, obj interface{}) error {
	ctype := res.Header.Get("Content-Type")
	if !(lfsMediaTypeRE.MatchString(ctype) || jsonMediaTypeRE.MatchString(ctype)) {
		return nil
	}

	err := json.NewDecoder(res.Body).Decode(obj)
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	if err != nil {
		return errutil.Errorf(err, "Unable to parse HTTP response for %s", TraceHttpReq(res.Request))
	}

	return nil
}

// GetDefaultError returns the default text for standard error codes (blank if none)
func GetDefaultError(code int) string {
	if s, ok := defaultErrors[code]; ok {
		return s
	}
	return ""
}

// Check the response from a HTTP request for problems
func handleResponse(res *http.Response, creds auth.Creds) error {
	auth.SaveCredentials(config.Config, creds, res)

	if res.StatusCode < 400 {
		return nil
	}

	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	cliErr := &ClientError{}
	err := DecodeResponse(res, cliErr)
	if err == nil {
		if len(cliErr.Message) == 0 {
			err = defaultError(res)
		} else {
			err = errutil.Error(cliErr)
		}
	}

	if res.StatusCode == 401 {
		return errutil.NewAuthError(err)
	}

	if res.StatusCode > 499 && res.StatusCode != 501 && res.StatusCode != 509 {
		return errutil.NewFatalError(err)
	}

	return err
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

	return errutil.Error(fmt.Errorf(msgFmt, res.Request.URL))
}

func SetErrorResponseContext(err error, res *http.Response) {
	errutil.ErrorSetContext(err, "Status", res.Status)
	setErrorHeaderContext(err, "Request", res.Header)
	setErrorRequestContext(err, res.Request)
}

func setErrorRequestContext(err error, req *http.Request) {
	errutil.ErrorSetContext(err, "Endpoint", config.Config.Endpoint(auth.GetOperationForRequest(req)).Url)
	errutil.ErrorSetContext(err, "URL", TraceHttpReq(req))
	setErrorHeaderContext(err, "Response", req.Header)
}

func setErrorHeaderContext(err error, prefix string, head http.Header) {
	for key, _ := range head {
		contextKey := fmt.Sprintf("%s:%s", prefix, key)
		if _, skip := hiddenHeaders[key]; skip {
			errutil.ErrorSetContext(err, contextKey, "--")
		} else {
			errutil.ErrorSetContext(err, contextKey, head.Get(key))
		}
	}
}
