package lfsapi

import (
	"net/http"

	"github.com/git-lfs/git-lfs/creds"
)

func (c *Client) doWithNegotiate(req *http.Request, credWrapper creds.CredentialHelperWrapper) (*http.Response, error) {
	// There are two possibilities here if we're using Negotiate
	// authentication.  One is that we're using Kerberos, which we try
	// first.  The other is that we're using NTLM, which we no longer
	// support.  Fail in that case.
	return c.doWithAccess(req, "", nil, creds.NegotiateAccess)
}
