package lfsapi

import (
	"fmt"
	"net/http"

	"github.com/git-lfs/git-lfs/errors"
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

func (c *Client) handleResponse(res *http.Response) error {
	if res.StatusCode < 400 {
		return nil
	}

	cliErr := &ClientError{}
	err := DecodeJSON(res, cliErr)
	if IsDecodeTypeError(err) {
		err = nil
	}

	if err == nil {
		if len(cliErr.Message) == 0 {
			err = defaultError(res)
		} else {
			err = errors.Wrap(cliErr, "http")
		}
	}

	if res.StatusCode == 401 {
		return errors.NewAuthError(err)
	}

	if res.StatusCode > 499 && res.StatusCode != 501 && res.StatusCode != 507 && res.StatusCode != 509 {
		return errors.NewFatalError(err)
	}

	return err
}

var (
	defaultErrors = map[int]string{
		400: "Client error: %s",
		401: "Authorization error: %s\nCheck that you have proper access to the repository",
		403: "Authorization error: %s\nCheck that you have proper access to the repository",
		404: "Repository or object not found: %s\nCheck that it exists and that you have proper access to it",
		429: "Rate limit exceeded: %s",
		500: "Server error: %s",
		501: "Not Implemented: %s",
		507: "Insufficient server storage: %s",
		509: "Bandwidth limit exceeded: %s",
	}
)

func defaultError(res *http.Response) error {
	var msgFmt string

	if f, ok := defaultErrors[res.StatusCode]; ok {
		msgFmt = f
	} else if res.StatusCode < 500 {
		msgFmt = defaultErrors[400] + fmt.Sprintf(" from HTTP %d", res.StatusCode)
	} else {
		msgFmt = defaultErrors[500] + fmt.Sprintf(" from HTTP %d", res.StatusCode)
	}

	return errors.Errorf(msgFmt, res.Request.URL)
}
