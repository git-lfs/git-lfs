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
	"github.com/rubyist/tracerx"
)

var (
	defaultCredentialHelper = &CommandCredentialHelper{}
	defaultNetrcFinder      = &noFinder{}
	defaultEndpointFinder   = NewEndpointFinder(nil)
)

func (c *Client) DoWithAuth(remote string, req *http.Request) (*http.Response, error) {
	credHelper := c.Credentials
	if credHelper == nil {
		credHelper = defaultCredentialHelper
	}

	netrcFinder := c.Netrc
	if netrcFinder == nil {
		netrcFinder = defaultNetrcFinder
	}

	ef := c.Endpoints
	if ef == nil {
		ef = defaultEndpointFinder
	}

	apiEndpoint, access, creds, credsURL, err := getCreds(credHelper, netrcFinder, ef, remote, req)
	if err != nil {
		return nil, err
	}

	res, err := c.doWithCreds(req, creds, credsURL, access)
	if err != nil {
		if errors.IsAuthError(err) {
			newAccess := getAuthAccess(res)
			if newAccess != access {
				c.Endpoints.SetAccess(apiEndpoint.Url, newAccess)
			}

			if access == NoneAccess || creds != nil {
				tracerx.Printf("api: http response indicates %q authentication. Resubmitting...", newAccess)
				req.Header.Del("Authorization")
				if creds != nil {
					credHelper.Reject(creds)
				}
				return c.DoWithAuth(remote, req)
			}
		}

		err = errors.Wrap(err, "http")
	}

	if res == nil {
		return nil, err
	}

	switch res.StatusCode {
	case 401, 403:
		credHelper.Reject(creds)
	default:
		if res.StatusCode < 300 && res.StatusCode > 199 {
			credHelper.Approve(creds)
		}
	}

	return res, err
}

func (c *Client) doWithCreds(req *http.Request, creds Creds, credsURL *url.URL, access Access) (*http.Response, error) {
	if access == NTLMAccess {
		return c.doWithNTLM(req, creds, credsURL)
	}
	return c.Do(req)
}

func getCreds(credHelper CredentialHelper, netrcFinder NetrcFinder, ef EndpointFinder, remote string, req *http.Request) (Endpoint, Access, Creds, *url.URL, error) {
	operation := getReqOperation(req)
	apiEndpoint := ef.Endpoint(operation, remote)
	access := ef.AccessFor(apiEndpoint.Url)
	if access == NTLMAccess || requestHasAuth(req) || setAuthFromNetrc(netrcFinder, req) || access == NoneAccess {
		return apiEndpoint, access, nil, nil, nil
	}

	credsURL, err := getCredURLForAPI(ef, operation, remote, apiEndpoint, req)
	if err != nil {
		return apiEndpoint, access, nil, nil, errors.Wrap(err, "creds")
	}

	if credsURL == nil {
		return apiEndpoint, access, nil, nil, nil
	}

	creds, err := fillCredentials(credHelper, ef, req, credsURL)
	return apiEndpoint, access, creds, credsURL, err
}

func fillCredentials(credHelper CredentialHelper, ef EndpointFinder, req *http.Request, u *url.URL) (Creds, error) {
	path := strings.TrimPrefix(u.Path, "/")
	input := Creds{"protocol": u.Scheme, "host": u.Host, "path": path}
	if u.User != nil && u.User.Username() != "" {
		input["username"] = u.User.Username()
	}

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

	if err != nil {
		return nil, err
	}

	tracerx.Printf("Filled credentials for %s", u)
	setRequestAuth(req, creds["username"], creds["password"])

	return creds, err
}

func setAuthFromNetrc(netrcFinder NetrcFinder, req *http.Request) bool {
	hostname := req.URL.Host
	var host string

	if strings.Contains(hostname, ":") {
		var err error
		host, _, err = net.SplitHostPort(hostname)
		if err != nil {
			tracerx.Printf("netrc: error parsing %q: %s", hostname, err)
			return false
		}
	} else {
		host = hostname
	}

	if machine := netrcFinder.FindMachine(host); machine != nil {
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
			gitRemoteURL, err := url.Parse(u)
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

func requestHasAuth(req *http.Request) bool {
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
