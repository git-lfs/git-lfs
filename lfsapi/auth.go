package lfsapi

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/go-netrc/netrc"
	"github.com/rubyist/tracerx"
)

var (
	defaultCredentialHelper = &commandCredentialHelper{}
	defaultNetrcFinder      = &noFinder{}
	defaultEndpointFinder   = NewEndpointFinder(nil)
)

// DoWithAuth sends an HTTP request to get an HTTP response. It attempts to add
// authentication from netrc or git's credential helpers if necessary,
// supporting basic and ntlm authentication.
func (c *Client) DoWithAuth(remote string, req *http.Request) (*http.Response, error) {
	return c.doWithAuth(remote, req, nil)
}

func (c *Client) doWithAuth(remote string, req *http.Request, via []*http.Request) (*http.Response, error) {
	req.Header = c.extraHeadersFor(req)

	apiEndpoint, access, credHelper, credsURL, creds, err := c.getCreds(remote, req)
	if err != nil {
		return nil, err
	}

	res, err := c.doWithCreds(req, credHelper, creds, credsURL, access, via)
	if err != nil {
		if errors.IsAuthError(err) {
			newAccess := getAuthAccess(res)
			if newAccess != access {
				c.Endpoints.SetAccess(apiEndpoint.Url, newAccess)
			}

			if creds != nil || (access == NoneAccess && len(req.Header.Get("Authorization")) == 0) {
				tracerx.Printf("api: http response indicates %q authentication. Resubmitting...", newAccess)
				if creds != nil {
					req.Header.Del("Authorization")
					credHelper.Reject(creds)
				}

				// This case represents a rejected request that
				// should have been authenticated but wasn't. Do
				// not count this against our redirection
				// maximum, so do not recur through doWithAuth
				// and instead call DoWithAuth.
				return c.DoWithAuth(remote, req)
			}
		}
	}

	if res != nil && res.StatusCode < 300 && res.StatusCode > 199 {
		credHelper.Approve(creds)
	}

	return res, err
}

func (c *Client) doWithCreds(req *http.Request, credHelper CredentialHelper, creds Creds, credsURL *url.URL, access Access, via []*http.Request) (*http.Response, error) {
	if access == NTLMAccess {
		return c.doWithNTLM(req, credHelper, creds, credsURL)
	}
	return c.do(req, "", via)
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
//    what kind of auth the LFS server expects. Could be BasicAccess, NTLMAccess,
//    or NoneAccess, in which the Git Credential Helper step is skipped. We do
//    not want to prompt the user for a password to fetch public repository data.
// 3. The Git Remote URL, which should be something like "https://git.com/repo.git"
//    This URL is used for the Git Credential Helper. This way existing https
//    Git remote credentials can be re-used for LFS.
func (c *Client) getCreds(remote string, req *http.Request) (Endpoint, Access, CredentialHelper, *url.URL, Creds, error) {
	ef := c.Endpoints
	if ef == nil {
		ef = defaultEndpointFinder
	}

	netrcFinder := c.Netrc
	if netrcFinder == nil {
		netrcFinder = defaultNetrcFinder
	}

	operation := getReqOperation(req)
	apiEndpoint := ef.Endpoint(operation, remote)
	access := ef.AccessFor(apiEndpoint.Url)

	if access != NTLMAccess {
		if requestHasAuth(req) || setAuthFromNetrc(netrcFinder, req) || access == NoneAccess {
			return apiEndpoint, access, nullCreds, nil, nil, nil
		}

		credsURL, err := getCredURLForAPI(ef, operation, remote, apiEndpoint, req)
		if err != nil {
			return apiEndpoint, access, nullCreds, nil, nil, errors.Wrap(err, "creds")
		}

		if credsURL == nil {
			return apiEndpoint, access, nullCreds, nil, nil, nil
		}

		credHelper, creds, err := c.getGitCreds(ef, req, credsURL)
		if err == nil {
			tracerx.Printf("Filled credentials for %s", credsURL)
			setRequestAuth(req, creds["username"], creds["password"])
		}
		return apiEndpoint, access, credHelper, credsURL, creds, err
	}

	// NTLM ONLY

	credsURL, err := url.Parse(apiEndpoint.Url)
	if err != nil {
		return apiEndpoint, access, nullCreds, nil, nil, errors.Wrap(err, "creds")
	}

	if netrcMachine := getAuthFromNetrc(netrcFinder, req); netrcMachine != nil {
		creds := Creds{
			"protocol": credsURL.Scheme,
			"host":     credsURL.Host,
			"username": netrcMachine.Login,
			"password": netrcMachine.Password,
			"source":   "netrc",
		}

		return apiEndpoint, access, nullCreds, credsURL, creds, nil
	}

	// NTLM uses creds to create the session
	credHelper, creds, err := c.getGitCreds(ef, req, credsURL)
	return apiEndpoint, access, credHelper, credsURL, creds, err
}

func (c *Client) getGitCreds(ef EndpointFinder, req *http.Request, u *url.URL) (CredentialHelper, Creds, error) {
	credHelper, input := c.getCredentialHelper(u)
	creds, err := credHelper.Fill(input)
	if creds == nil || len(creds) < 1 {
		errmsg := fmt.Sprintf("Git credentials for %s not found", u)
		if err != nil {
			errmsg = errmsg + ":\n" + err.Error()
		} else {
			errmsg = errmsg + "."
		}
		err = errors.New(errmsg)
	}

	return credHelper, creds, err
}

func getAuthFromNetrc(netrcFinder NetrcFinder, req *http.Request) *netrc.Machine {
	hostname := req.URL.Host
	var host string

	if strings.Contains(hostname, ":") {
		var err error
		host, _, err = net.SplitHostPort(hostname)
		if err != nil {
			tracerx.Printf("netrc: error parsing %q: %s", hostname, err)
			return nil
		}
	} else {
		host = hostname
	}

	return netrcFinder.FindMachine(host)
}

func setAuthFromNetrc(netrcFinder NetrcFinder, req *http.Request) bool {
	if machine := getAuthFromNetrc(netrcFinder, req); machine != nil {
		setRequestAuth(req, machine.Login, machine.Password)
		return true
	}

	return false
}

func getCredURLForAPI(ef EndpointFinder, operation, remote string, apiEndpoint Endpoint, req *http.Request) (*url.URL, error) {
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
			schemedUrl, _ := prependEmptySchemeIfAbsent(u)

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

// prependEmptySchemeIfAbsent prepends an empty scheme "//" if none was found in
// the URL in order to satisfy RFC 3986 §3.3, and `net/url.Parse()`.
//
// It returns a string parse-able with `net/url.Parse()` and a boolean whether
// or not an empty scheme was added.
func prependEmptySchemeIfAbsent(u string) (string, bool) {
	if hasScheme(u) {
		return u, false
	}

	colon := strings.Index(u, ":")
	slash := strings.Index(u, "/")

	if colon >= 0 && (slash < 0 || colon < slash) {
		// First path segment has a colon, assumed that it's a
		// scheme-less URL. Append an empty scheme on top to
		// satisfy RFC 3986 §3.3, and `net/url.Parse()`.
		return fmt.Sprintf("//%s", u), true
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

func getAuthAccess(res *http.Response) Access {
	for _, headerName := range authenticateHeaders {
		for _, auth := range res.Header[headerName] {
			pieces := strings.SplitN(strings.ToLower(auth), " ", 2)
			if len(pieces) == 0 {
				continue
			}

			switch Access(pieces[0]) {
			case NegotiateAccess, NTLMAccess:
				// When server sends Www-Authentication: Negotiate, it supports both Kerberos and NTLM.
				// Since git-lfs current does not support Kerberos, we will return NTLM in this case.
				return NTLMAccess
			}
		}
	}

	return BasicAccess
}
