package lfsapi

import (
	"net/http"

	"github.com/git-lfs/git-lfs/creds"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

func (c *Client) doWithNegotiate(req *http.Request, credWrapper creds.CredentialHelperWrapper) (*http.Response, error) {
	// There are two possibilities here if we're using Negotiate
	// authentication.  One is that we're using Kerberos, which we try
	// first.  The other is that we're using NTLM, which we try second with
	// single sign-on credentials.  Finally, if that also fails, we fall
	// back to prompting for credentials with NTLM and trying that.
	res, err := c.doWithAccess(req, "", nil, creds.NegotiateAccess)
	if err == nil || errors.IsAuthError(err) {
		if res.StatusCode != 401 {
			return res, nil
		}
	}

	// If we received an error, fall back to NTLM.  That will be the case if
	// we don't have any cached credentials, which on Unix will look like a
	// failed attempt to read a file in /tmp, not a standard auth error.

	tracerx.Printf("attempting NTLM as fallback")
	res, err = c.doWithAccess(req, "", nil, creds.NTLMAccess)
	if err != nil && !errors.IsAuthError(err) {
		return res, err
	}

	if res.StatusCode != 401 {
		return res, nil
	}

	return c.ntlmReAuth(req, credWrapper, true)
}
