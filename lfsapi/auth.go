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

func getCreds(credHelper CredentialHelper, netrcFinder NetrcFinder, ef EndpointFinder, remote string, req *http.Request) (Creds, error) {
	if skipCreds(ef, req) {
		return nil, nil
	}

	operation := getReqOperation(req)
	apiEndpoint := ef.Endpoint(operation, remote)
	credsUrl, err := getCredURLForAPI(ef, operation, remote, apiEndpoint, req)
	if err != nil {
		return nil, errors.Wrap(err, "creds")
	}

	if credsUrl == nil {
		return nil, nil
	}

	if setAuthFromNetrc(netrcFinder, req) {
		return nil, nil
	}

	return fillCredentials(credHelper, ef, req, credsUrl)
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

func getCredURLForAPI(ef EndpointFinder, operation, remote string, e Endpoint, req *http.Request) (*url.URL, error) {
	apiUrl, err := url.Parse(e.Url)
	if err != nil {
		return nil, err
	}

	// if the LFS request doesn't match the current LFS url, don't bother
	// attempting to set the Authorization header from the LFS or Git remote URLs.
	if req.URL.Scheme != apiUrl.Scheme ||
		req.URL.Host != apiUrl.Host {
		return req.URL, nil
	}

	if setRequestAuthFromUrl(ef, req, e, apiUrl) {
		return nil, nil
	}

	credsUrl := apiUrl
	if len(remote) > 0 {
		if u := ef.GitRemoteURL(remote, operation == "upload"); u != "" {
			gitRemoteUrl, err := url.Parse(u)
			if err != nil {
				return nil, err
			}

			if gitRemoteUrl.Scheme == apiUrl.Scheme &&
				gitRemoteUrl.Host == apiUrl.Host {

				if setRequestAuthFromUrl(ef, req, e, gitRemoteUrl) {
					return nil, nil
				}

				credsUrl = gitRemoteUrl
			}
		}
	}

	return credsUrl, nil
}

func skipCreds(ef EndpointFinder, req *http.Request) bool {
	if ef.AccessFor(req.URL.String()) == NTLMAccess {
		return false
	}

	if len(req.Header.Get("Authorization")) > 0 {
		return true
	}

	return len(req.URL.Query().Get("token")) > 0
}

func setRequestAuthFromUrl(ef EndpointFinder, req *http.Request, apiEndpoint Endpoint, u *url.URL) bool {
	if ef.AccessFor(apiEndpoint.Url) == NTLMAccess || u.User == nil {
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
