// +build !windows

package lfsapi

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
)

func (c *Client) ntlmAuthenticateRequest(req *http.Request, creds *ntmlCredentials) (*http.Response, error) {
	negotiate, err := base64.StdEncoding.DecodeString(ntlmNegotiateMessage)
	if err != nil {
		return nil, err
	}

	chRes, challengeMsg, err := c.ntlmSendType1Message(req, negotiate)
	if err != nil {
		return chRes, err
	}

	challenge, err := ntlm.ParseChallengeMessage(challengeMsg)
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

	return c.ntlmSendType3Message(req, authenticate.Bytes())
}

func (c *Client) ntlmClientSession(creds *ntmlCredentials) (ntlm.ClientSession, error) {
	c.ntlmMu.Lock()
	defer c.ntlmMu.Unlock()

	if creds == nil {
		return nil, fmt.Errorf("Your user name must be of the form DOMAIN\\user. Single-sign-on is not supported.")
	}

	if c.ntlmSessions == nil {
		c.ntlmSessions = make(map[string]ntlm.ClientSession)
	}

	if ses, ok := c.ntlmSessions[creds.domain]; ok {
		return ses, nil
	}

	session, err := ntlm.CreateClientSession(ntlm.Version2, ntlm.ConnectionOrientedMode)
	if err != nil {
		return nil, err
	}

	session.SetUserInfo(creds.username, creds.password, creds.domain)
	c.ntlmSessions[creds.domain] = session
	return session, nil
}

const ntlmNegotiateMessage = "TlRMTVNTUAABAAAAB7IIogwADAAzAAAACwALACgAAAAKAAAoAAAAD1dJTExISS1NQUlOTk9SVEhBTUVSSUNB"
