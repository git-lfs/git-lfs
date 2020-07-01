package lfsapi

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/creds"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfshttp"
	"github.com/rubyist/tracerx"
)

var (
	defaultEndpointFinder = NewEndpointFinder(nil)
)

// DoWithAuth sends an HTTP request to get an HTTP response. It attempts to add
// authentication from netrc or git's credential helpers if necessary,
// supporting basic and ntlm authentication.
func (c *Client) DoWithAuth(remote string, access creds.Access, req *http.Request) (*http.Response, error) {
	res, err := c.doWithAuth(remote, access, req, nil)

	if errors.IsAuthError(err) {
		if len(req.Header.Get("Authorization")) == 0 {
			// This case represents a rejected request that
			// should have been authenticated but wasn't. Do
			// not count this against our redirection
			// maximum.
			newAccess := c.Endpoints.AccessFor(access.URL())
			tracerx.Printf("api: http response indicates %q authentication. Resubmitting...", newAccess.Mode())
			return c.DoWithAuth(remote, newAccess, req)
		}
	}

	return res, err
}

// DoWithAuthNoRetry sends an HTTP request to get an HTTP response. It works in
// the same way as DoWithAuth, but will not retry the request if it fails with
// an authorization error.
func (c *Client) DoWithAuthNoRetry(remote string, access creds.Access, req *http.Request) (*http.Response, error) {
	return c.doWithAuth(remote, access, req, nil)
}

// DoAPIRequestWithAuth sends an HTTP request to get an HTTP response similarly
// to DoWithAuth, but using the LFS API endpoint for the provided remote and
// operation to determine the access mode.
func (c *Client) DoAPIRequestWithAuth(remote string, req *http.Request) (*http.Response, error) {
	operation := getReqOperation(req)
	apiEndpoint := c.Endpoints.Endpoint(operation, remote)
	access := c.Endpoints.AccessFor(apiEndpoint.Url)
	return c.DoWithAuth(remote, access, req)
}

func (c *Client) doWithAuth(remote string, access creds.Access, req *http.Request, via []*http.Request) (*http.Response, error) {
	req.Header = c.client.ExtraHeadersFor(req)

	credWrapper, err := c.getCreds(remote, access, req)
	if err != nil {
		return nil, err
	}

	res, err := c.doWithCreds(req, credWrapper, access, via)
	if err != nil {
		if errors.IsAuthError(err) {
			newAccess := access.Upgrade(getAuthAccess(res))
			if newAccess.Mode() != access.Mode() {
				c.Endpoints.SetAccess(newAccess)
			}

			if credWrapper.Creds != nil {
				req.Header.Del("Authorization")
				credWrapper.CredentialHelper.Reject(credWrapper.Creds)
			}
		}
	}

	if res != nil && res.StatusCode < 300 && res.StatusCode > 199 {
		credWrapper.CredentialHelper.Approve(credWrapper.Creds)
	}

	return res, err
}

func (c *Client) doWithCreds(req *http.Request, credWrapper creds.CredentialHelperWrapper, access creds.Access, via []*http.Request) (*http.Response, error) {
	if access.Mode() == creds.NTLMAccess {
		return c.doWithNTLM(req, credWrapper)
	} else if access.Mode() == creds.NegotiateAccess {
		return c.doWithNegotiate(req, credWrapper)
	}

	req.Header.Set("User-Agent", lfshttp.UserAgent)

	client, err := c.client.HttpClient(req.URL, access.Mode())
	if err != nil {
		return nil, err
	}

	redirectedReq, res, err := c.client.DoWithRedirect(client, req, "", via)
	if err != nil || res != nil {
		return res, err
	}

	if redirectedReq == nil {
		return res, errors.New("failed to redirect request")
	}

	return c.doWithAuth("", access, redirectedReq, via)
}

// getCreds fills the authorization header for the given request if possible,
// from the following sources:
//
// 1. NTLM access is handled elsewhere.
// 2. Existing Authorization or ?token query tells LFS that the request is ready.
// 3. Netrc based on the hostname.
// 4. URL authentication on the Endpoint URL or the Git Remote URL.
// 5. Git Credential Helper, potentially prompting the user.
//
// There are three URLs in play, that make this a little confusing.
//
// 1. The request URL, which should be something like "https://git.com/repo.git/info/lfs/objects/batch"
// 2. The LFS API URL, which should be something like "https://git.com/repo.git/info/lfs"
//    This URL used for the "lfs.URL.access" git config key, which determines
//    what kind of auth the LFS server expects. Could be BasicAccess,
//    NTLMAccess, NegotiateAccess, or NoneAccess, in which the Git Credential
//    Helper step is skipped. We do not want to prompt the user for a password
//    to fetch public repository data.
// 3. The Git Remote URL, which should be something like "https://git.com/repo.git"
//    This URL is used for the Git Credential Helper. This way existing https
//    Git remote credentials can be re-used for LFS.
func (c *Client) getCreds(remote string, access creds.Access, req *http.Request) (creds.CredentialHelperWrapper, error) {
	ef := c.Endpoints
	if ef == nil {
		ef = defaultEndpointFinder
	}

	operation := getReqOperation(req)
	apiEndpoint := ef.Endpoint(operation, remote)

	if access.Mode() != creds.NTLMAccess && access.Mode() != creds.NegotiateAccess {
		if requestHasAuth(req) || access.Mode() == creds.NoneAccess {
			return creds.CredentialHelperWrapper{CredentialHelper: creds.NullCreds, Input: nil, Url: nil, Creds: nil}, nil
		}

		credsURL, err := getCredURLForAPI(ef, operation, remote, apiEndpoint, req)
		if err != nil {
			return creds.CredentialHelperWrapper{CredentialHelper: creds.NullCreds, Input: nil, Url: nil, Creds: nil}, errors.Wrap(err, "creds")
		}

		if credsURL == nil {
			return creds.CredentialHelperWrapper{CredentialHelper: creds.NullCreds, Input: nil, Url: nil, Creds: nil}, nil
		}

		credWrapper := c.getGitCredsWrapper(ef, req, credsURL)
		err = credWrapper.FillCreds()
		if err == nil {
			tracerx.Printf("Filled credentials for %s", credsURL)
			setRequestAuth(req, credWrapper.Creds["username"], credWrapper.Creds["password"])
		}
		return credWrapper, err
	}

	// NTLM and Negotiate only

	credsURL, err := url.Parse(apiEndpoint.Url)
	if err != nil {
		return creds.CredentialHelperWrapper{CredentialHelper: creds.NullCreds, Input: nil, Url: nil, Creds: nil}, errors.Wrap(err, "creds")
	}

	// NTLM uses creds to create the session
	credWrapper := c.getGitCredsWrapper(ef, req, credsURL)
	return credWrapper, err
}

func (c *Client) getGitCredsWrapper(ef EndpointFinder, req *http.Request, u *url.URL) creds.CredentialHelperWrapper {
	return c.credContext.GetCredentialHelper(c.Credentials, u)
}

func getCredURLForAPI(ef EndpointFinder, operation, remote string, apiEndpoint lfshttp.Endpoint, req *http.Request) (*url.URL, error) {
	apiURL, err := url.Parse(apiEndpoint.Url)
	if err != nil {
		return nil, err
	}

	// if the LFS request doesn't match the current LFS url, don't bother
	// attempting to set the Authorization header from the LFS or Git remote URLs.
	if req.URL.Scheme != apiURL.Scheme ||
		req.URL.Host != apiURL.Host {
		return req.URL, nil
	}

	if setRequestAuthFromURL(req, apiURL) {
		return nil, nil
	}

	if len(remote) > 0 {
		if u := ef.GitRemoteURL(remote, operation == "upload"); u != "" {
			schemedUrl, _ := fixSchemelessURL(u)

			gitRemoteURL, err := url.Parse(schemedUrl)
			if err != nil {
				return nil, err
			}

			if gitRemoteURL.Scheme == apiURL.Scheme &&
				gitRemoteURL.Host == apiURL.Host {

				if setRequestAuthFromURL(req, gitRemoteURL) {
					return nil, nil
				}

				return gitRemoteURL, nil
			}
		}
	}

	return apiURL, nil
}

// fixSchemelessURL prepends an empty scheme "//" if none was found in
// the URL and replaces the first colon with a slash in order to satisfy RFC
// 3986 §3.3, and `net/url.Parse()`.
//
// It returns a string parse-able with `net/url.Parse()` and a boolean whether
// or not an empty scheme was added.
func fixSchemelessURL(u string) (string, bool) {
	if hasScheme(u) {
		return u, false
	}

	colon := strings.Index(u, ":")
	slash := strings.Index(u, "/")

	if colon >= 0 && (slash < 0 || colon < slash) {
		// First path segment has a colon, assumed that it's a
		// scheme-less URL. Append an empty scheme on top to
		// satisfy RFC 3986 §3.3, and `net/url.Parse()`.
		//
		// In addition, replace the first colon with a slash since
		// otherwise the colon looks like it's introducing a port
		// number.
		return fmt.Sprintf("//%s", strings.Replace(u, ":", "/", 1)), true
	}
	return u, true
}

var (
	// supportedSchemes is the list of URL schemes the `lfsapi` package
	// supports.
	supportedSchemes = []string{"ssh", "http", "https"}
)

// hasScheme returns whether or not a given string (taken to represent a RFC
// 3986 URL) has a scheme that is supported by the `lfsapi` package.
func hasScheme(what string) bool {
	for _, scheme := range supportedSchemes {
		if strings.HasPrefix(what, fmt.Sprintf("%s://", scheme)) {
			return true
		}
	}

	return false
}

func requestHasAuth(req *http.Request) bool {
	// The "Authorization" string constant is safe, since we assume that all
	// request headers have been canonicalized.
	if len(req.Header.Get("Authorization")) > 0 {
		return true
	}

	return len(req.URL.Query().Get("token")) > 0
}

func setRequestAuthFromURL(req *http.Request, u *url.URL) bool {
	if u.User == nil {
		return false
	}

	if pass, ok := u.User.Password(); ok {
		fmt.Fprintln(os.Stderr, "warning: current Git remote contains credentials")
		setRequestAuth(req, u.User.Username(), pass)
		return true
	}

	return false
}

func setRequestAuth(req *http.Request, user, pass string) {
	// better not be NTLM!
	if len(user) == 0 && len(pass) == 0 {
		return
	}

	token := fmt.Sprintf("%s:%s", user, pass)
	auth := "Basic " + strings.TrimSpace(base64.StdEncoding.EncodeToString([]byte(token)))
	req.Header.Set("Authorization", auth)
}

func getReqOperation(req *http.Request) string {
	operation := "download"
	if req.Method == "POST" || req.Method == "PUT" {
		operation = "upload"
	}
	return operation
}

var (
	authenticateHeaders = []string{"Lfs-Authenticate", "Www-Authenticate"}
)

func getAuthAccess(res *http.Response) creds.AccessMode {
	for _, headerName := range authenticateHeaders {
		for _, auth := range res.Header[headerName] {
			pieces := strings.SplitN(strings.ToLower(auth), " ", 2)
			if len(pieces) == 0 {
				continue
			}

			switch creds.AccessMode(pieces[0]) {
			case creds.NegotiateAccess, creds.NTLMAccess:
				return creds.AccessMode(pieces[0])
			}
		}
	}

	return creds.BasicAccess
}
