package lfsapi

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/git-lfs/git-lfs/errors"
)

func (c *Client) doWithNTLM(req *http.Request, credHelper CredentialHelper, creds Creds, credsURL *url.URL) (*http.Response, error) {
	res, err := c.Do(req)
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
	body, err := rewoundRequestBody(req)
	if err != nil {
		return nil, err
	}
	req.Body = body

	chRes, challengeMsg, err := c.ntlmNegotiate(req, ntlmNegotiateMessage)
	if err != nil {
		return chRes, err
	}

	body, err = rewoundRequestBody(req)
	if err != nil {
		return nil, err
	}
	req.Body = body

	res, err := c.ntlmChallenge(req, challengeMsg, creds)
	if err != nil {
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

func (c *Client) ntlmNegotiate(req *http.Request, message string) (*http.Response, []byte, error) {
	req.Header.Add("Authorization", message)
	res, err := c.Do(req)
	if err != nil && !errors.IsAuthError(err) {
		return res, nil, err
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	by, err := parseChallengeResponse(res)
	return res, by, err
}

func (c *Client) ntlmChallenge(req *http.Request, challengeBytes []byte, creds Creds) (*http.Response, error) {
	challenge, err := ntlm.ParseChallengeMessage(challengeBytes)
	if err != nil {
		return nil, err
	}

	session, err := c.ntlmClientSession(creds)
	if err != nil {
		return nil, err
	}

	session.ProcessChallengeMessage(challenge)
	authenticate, err := session.GenerateAuthenticateMessage()
	if err != nil {
		return nil, err
	}

	authMsg := base64.StdEncoding.EncodeToString(authenticate.Bytes())
	req.Header.Set("Authorization", "NTLM "+authMsg)
	return c.Do(req)
}

func (c *Client) ntlmClientSession(creds Creds) (ntlm.ClientSession, error) {
	c.ntlmMu.Lock()
	defer c.ntlmMu.Unlock()

	splits := strings.Split(creds["username"], "\\")
	if len(splits) != 2 {
		return nil, fmt.Errorf("Your user name must be of the form DOMAIN\\user. It is currently %s", creds["username"])
	}

	domain := strings.ToUpper(splits[0])
	username := splits[1]

	if c.ntlmSessions == nil {
		c.ntlmSessions = make(map[string]ntlm.ClientSession)
	}

	if ses, ok := c.ntlmSessions[domain]; ok {
		return ses, nil
	}

	session, err := ntlm.CreateClientSession(ntlm.Version2, ntlm.ConnectionOrientedMode)
	if err != nil {
		return nil, err
	}

	session.SetUserInfo(username, creds["password"], strings.ToUpper(splits[0]))
	c.ntlmSessions[domain] = session
	return session, nil
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

const ntlmNegotiateMessage = "NTLM TlRMTVNTUAABAAAAB7IIogwADAAzAAAACwALACgAAAAKAAAoAAAAD1dJTExISS1NQUlOTk9SVEhBTUVSSUNB"
