package lfshttp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/tr"
)

type httpError interface {
	Error() string
	HTTPResponse() *http.Response
}

func IsHTTP(err error) (*http.Response, bool) {
	if httpErr, ok := err.(httpError); ok {
		return httpErr.HTTPResponse(), true
	}
	return nil, false
}

type ClientError struct {
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url,omitempty"`
	RequestId        string `json:"request_id,omitempty"`
	response         *http.Response
}

func (e *ClientError) HTTPResponse() *http.Response {
	return e.response
}

func (e *ClientError) Error() string {
	return e.Message
}

func (c *Client) handleResponse(res *http.Response) error {
	if res.StatusCode < 400 {
		return nil
	}

	cliErr := &ClientError{response: res}
	err := DecodeJSON(res, cliErr)
	if IsDecodeTypeError(err) {
		err = nil
	}

	if err == nil {
		if len(cliErr.Message) == 0 {
			err = defaultError(res)
		} else {
			err = cliErr
		}
	}

	if res.StatusCode == 401 {
		return errors.NewAuthError(err)
	}

	if res.StatusCode == 422 {
		return errors.NewUnprocessableEntityError(err)
	}

	if res.StatusCode == 429 {
		// The Retry-After header could be set, check to see if it exists.
		h := res.Header.Get("Retry-After")
		retLaterErr := errors.NewRetriableLaterError(err, h)
		if retLaterErr != nil {
			return retLaterErr
		}
	}

	if res.StatusCode > 499 && res.StatusCode != 501 && res.StatusCode != 507 && res.StatusCode != 509 {
		return errors.NewFatalError(err)
	}

	return err
}

type statusCodeError struct {
	response *http.Response
}

func NewStatusCodeError(res *http.Response) error {
	return &statusCodeError{response: res}
}

func (e *statusCodeError) Error() string {
	req := e.response.Request
	return tr.Tr.Get("Invalid HTTP status for %s %s: %d",
		req.Method,
		strings.SplitN(req.URL.String(), "?", 2)[0],
		e.response.StatusCode,
	)
}

func (e *statusCodeError) HTTPResponse() *http.Response {
	return e.response
}

func defaultError(res *http.Response) error {
	var msgFmt string

	defaultErrors := map[int]string{
		400: tr.Tr.Get("Client error: %%s"),
		401: tr.Tr.Get("Authorization error: %%s\nCheck that you have proper access to the repository"),
		403: tr.Tr.Get("Authorization error: %%s\nCheck that you have proper access to the repository"),
		404: tr.Tr.Get("Repository or object not found: %%s\nCheck that it exists and that you have proper access to it"),
		422: tr.Tr.Get("Unprocessable entity: %%s"),
		429: tr.Tr.Get("Rate limit exceeded: %%s"),
		500: tr.Tr.Get("Server error: %%s"),
		501: tr.Tr.Get("Not Implemented: %%s"),
		507: tr.Tr.Get("Insufficient server storage: %%s"),
		509: tr.Tr.Get("Bandwidth limit exceeded: %%s"),
	}
	if f, ok := defaultErrors[res.StatusCode]; ok {
		msgFmt = f
	} else if res.StatusCode < 500 {
		msgFmt = tr.Tr.Get("Client error %%s from HTTP %d", res.StatusCode)
	} else {
		msgFmt = tr.Tr.Get("Server error %%s from HTTP %d", res.StatusCode)
	}

	return errors.Errorf(fmt.Sprintf(msgFmt), res.Request.URL)
}
