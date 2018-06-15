package lfsapi

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
)

type ntmlCredentials struct {
	domain   string
	username string
	password string
}

func (c *Client) doWithNTLM(req *http.Request, credHelper CredentialHelper, creds Creds, credsURL *url.URL) (*http.Response, error) {
	res, err := c.do(req, "", nil)
	if err != nil && !errors.IsAuthError(err) {
		return res, err
	}

	if res.StatusCode != 401 {
		return res, nil
	}

	return c.ntlmReAuth(req, credHelper, creds, true)
}

// If the status is 401 then we need to re-authenticate
func (c *Client) ntlmReAuth(req *http.Request, credHelper CredentialHelper, creds Creds, retry bool) (*http.Response, error) {
	ntmlCreds, err := ntlmGetCredentials(creds)
	if err != nil {
		return nil, err
	}

	res, err := c.ntlmAuthenticateRequest(req, ntmlCreds)
	if err != nil && !errors.IsAuthError(err) {
		return res, err
	}

	switch res.StatusCode {
	case 401:
		credHelper.Reject(creds)
		if retry {
			return c.ntlmReAuth(req, credHelper, creds, false)
		}
	case 403:
		credHelper.Reject(creds)
	default:
		if res.StatusCode < 300 && res.StatusCode > 199 {
			credHelper.Approve(creds)
		}
	}

	return res, nil
}

func (c *Client) ntlmSendType1Message(req *http.Request, message []byte) (*http.Response, []byte, error) {
	res, err := c.ntlmSendMessage(req, message)
	if err != nil && !errors.IsAuthError(err) {
		return res, nil, err
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	by, err := parseChallengeResponse(res)
	return res, by, err
}

func (c *Client) ntlmSendType3Message(req *http.Request, authenticate []byte) (*http.Response, error) {
	return c.ntlmSendMessage(req, authenticate)
}

func (c *Client) ntlmSendMessage(req *http.Request, message []byte) (*http.Response, error) {
	body, err := rewoundRequestBody(req)
	if err != nil {
		return nil, err
	}
	req.Body = body

	msg := base64.StdEncoding.EncodeToString(message)
	req.Header.Set("Authorization", "NTLM "+msg)
	return c.do(req, "", nil)
}

func parseChallengeResponse(res *http.Response) ([]byte, error) {
	header := res.Header.Get("Www-Authenticate")
	if len(header) < 6 {
		return nil, fmt.Errorf("Invalid NTLM challenge response: %q", header)
	}

	//parse out the "NTLM " at the beginning of the response
	challenge := header[5:]
	val, err := base64.StdEncoding.DecodeString(challenge)

	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func rewoundRequestBody(req *http.Request) (io.ReadCloser, error) {
	if req.Body == nil {
		return nil, nil
	}

	body, ok := req.Body.(ReadSeekCloser)
	if !ok {
		return nil, fmt.Errorf("Request body must implement io.ReadCloser and io.Seeker. Got: %T", body)
	}

	_, err := body.Seek(0, io.SeekStart)
	return body, err
}

func ntlmGetCredentials(creds Creds) (*ntmlCredentials, error) {
	username := creds["username"]
	password := creds["password"]

	if username == "" && password == "" {
		return nil, nil
	}

	splits := strings.Split(username, "\\")
	if len(splits) != 2 {
		return nil, fmt.Errorf("Your user name must be of the form DOMAIN\\user. It is currently '%s'", username)
	}

	domain := strings.ToUpper(splits[0])
	username = splits[1]

	return &ntmlCredentials{domain: domain, username: username, password: password}, nil
}
