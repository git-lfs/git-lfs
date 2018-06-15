// +build windows

package lfsapi

import (
	"net/http"

	"github.com/alexbrainman/sspi"
	"github.com/alexbrainman/sspi/ntlm"
)

func (c *Client) ntlmAuthenticateRequest(req *http.Request, creds *ntmlCredentials) (*http.Response, error) {
	var sspiCreds *sspi.Credentials
	var err error
	if creds == nil {
		sspiCreds, err = ntlm.AcquireCurrentUserCredentials()
	} else {
		sspiCreds, err = ntlm.AcquireUserCredentials(creds.domain, creds.username, creds.password)
	}

	if err != nil {
		return nil, err
	}
	defer sspiCreds.Release()

	secctx, negotiate, err := ntlm.NewClientContext(sspiCreds)
	if err != nil {
		return nil, err
	}
	defer secctx.Release()

	chRes, challengeMsg, err := c.ntlmSendType1Message(req, negotiate)
	if err != nil {
		return chRes, err
	}

	authenticateMsg, err := secctx.Update(challengeMsg)
	if err != nil {
		return nil, err
	}

	return c.ntlmSendType3Message(req, authenticateMsg)
}
