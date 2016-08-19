package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

// getCreds gets the credentials for a HTTP request and sets the given
// request's Authorization header with them using Basic Authentication.
// 1. Check the URL for authentication. Ex: http://user:pass@example.com
// 2. Check netrc for authentication.
// 3. Check the Git remote URL for authentication IF it's the same scheme and
//    host of the URL.
// 4. Ask 'git credential' to fill in the password from one of the above URLs.
//
// This prefers the Git remote URL for checking credentials so that users only
// have to enter their passwords once for Git and Git LFS. It uses the same
// URL path that Git does, in case 'useHttpPath' is enabled in the Git config.
func GetCreds(cfg *config.Configuration, req *http.Request) (Creds, error) {
	if skipCredsCheck(cfg, req) {
		return nil, nil
	}

	credsUrl, err := getCredURLForAPI(cfg, req)
	if err != nil {
		return nil, errors.Wrap(err, "creds")
	}

	if credsUrl == nil {
		return nil, nil
	}

	if setCredURLFromNetrc(cfg, req) {
		return nil, nil
	}

	return fillCredentials(cfg, req, credsUrl)
}

func getCredURLForAPI(cfg *config.Configuration, req *http.Request) (*url.URL, error) {
	operation := GetOperationForRequest(req)
	apiUrl, err := url.Parse(cfg.Endpoint(operation).Url)
	if err != nil {
		return nil, err
	}

	// if the LFS request doesn't match the current LFS url, don't bother
	// attempting to set the Authorization header from the LFS or Git remote URLs.
	if req.URL.Scheme != apiUrl.Scheme ||
		req.URL.Host != apiUrl.Host {
		return req.URL, nil
	}

	if setRequestAuthFromUrl(cfg, req, apiUrl) {
		return nil, nil
	}

	credsUrl := apiUrl
	if len(cfg.CurrentRemote) > 0 {
		if u := cfg.GitRemoteUrl(cfg.CurrentRemote, operation == "upload"); u != "" {
			gitRemoteUrl, err := url.Parse(u)
			if err != nil {
				return nil, err
			}

			if gitRemoteUrl.Scheme == apiUrl.Scheme &&
				gitRemoteUrl.Host == apiUrl.Host {

				if setRequestAuthFromUrl(cfg, req, gitRemoteUrl) {
					return nil, nil
				}

				credsUrl = gitRemoteUrl
			}
		}
	}
	return credsUrl, nil
}

func setCredURLFromNetrc(cfg *config.Configuration, req *http.Request) bool {
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

	machine, err := cfg.FindNetrcHost(host)
	if err != nil {
		tracerx.Printf("netrc: error finding match for %q: %s", hostname, err)
		return false
	}

	if machine == nil {
		return false
	}

	setRequestAuth(cfg, req, machine.Login, machine.Password)
	return true
}

func skipCredsCheck(cfg *config.Configuration, req *http.Request) bool {
	if cfg.NtlmAccess(GetOperationForRequest(req)) {
		return false
	}

	if len(req.Header.Get("Authorization")) > 0 {
		return true
	}

	q := req.URL.Query()
	return len(q["token"]) > 0
}

func fillCredentials(cfg *config.Configuration, req *http.Request, u *url.URL) (Creds, error) {
	path := strings.TrimPrefix(u.Path, "/")
	input := Creds{"protocol": u.Scheme, "host": u.Host, "path": path}
	if u.User != nil && u.User.Username() != "" {
		input["username"] = u.User.Username()
	}

	creds, err := execCreds(cfg, input, "fill")
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
	setRequestAuth(cfg, req, creds["username"], creds["password"])

	return creds, err
}

func SaveCredentials(cfg *config.Configuration, creds Creds, res *http.Response) {
	if creds == nil {
		return
	}

	switch res.StatusCode {
	case 401, 403:
		execCreds(cfg, creds, "reject")
	default:
		if res.StatusCode < 300 {
			execCreds(cfg, creds, "approve")
		}
	}
}

type Creds map[string]string

func (c Creds) Buffer() *bytes.Buffer {
	buf := new(bytes.Buffer)

	for k, v := range c {
		buf.Write([]byte(k))
		buf.Write([]byte("="))
		buf.Write([]byte(v))
		buf.Write([]byte("\n"))
	}

	return buf
}

// Credentials function which will be called whenever credentials are requested
type CredentialFunc func(*config.Configuration, Creds, string) (Creds, error)

func execCredsCommand(cfg *config.Configuration, input Creds, subCommand string) (Creds, error) {
	output := new(bytes.Buffer)
	cmd := exec.Command("git", "credential", subCommand)
	cmd.Stdin = input.Buffer()
	cmd.Stdout = output
	/*
		There is a reason we don't hook up stderr here:
		Git's credential cache daemon helper does not close its stderr, so if this
		process is the process that fires up the daemon, it will wait forever
		(until the daemon exits, really) trying to read from stderr.

		See https://github.com/github/git-lfs/issues/117 for more details.
	*/

	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if _, ok := err.(*exec.ExitError); ok {
		if !cfg.Os.Bool("GIT_TERMINAL_PROMPT", true) {
			return nil, fmt.Errorf("Change the GIT_TERMINAL_PROMPT env var to be prompted to enter your credentials for %s://%s.",
				input["protocol"], input["host"])
		}

		// 'git credential' exits with 128 if the helper doesn't fill the username
		// and password values.
		if subCommand == "fill" && err.Error() == "exit status 128" {
			return nil, nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("'git credential %s' error: %s\n", subCommand, err.Error())
	}

	creds := make(Creds)
	for _, line := range strings.Split(output.String(), "\n") {
		pieces := strings.SplitN(line, "=", 2)
		if len(pieces) < 2 || len(pieces[1]) < 1 {
			continue
		}
		creds[pieces[0]] = pieces[1]
	}

	return creds, nil
}

func setRequestAuthFromUrl(cfg *config.Configuration, req *http.Request, u *url.URL) bool {
	if !cfg.NtlmAccess(GetOperationForRequest(req)) && u.User != nil {
		if pass, ok := u.User.Password(); ok {
			fmt.Fprintln(os.Stderr, "warning: current Git remote contains credentials")
			setRequestAuth(cfg, req, u.User.Username(), pass)
			return true
		}
	}

	return false
}

func setRequestAuth(cfg *config.Configuration, req *http.Request, user, pass string) {
	if cfg.NtlmAccess(GetOperationForRequest(req)) {
		return
	}

	if len(user) == 0 && len(pass) == 0 {
		return
	}

	token := fmt.Sprintf("%s:%s", user, pass)
	auth := "Basic " + strings.TrimSpace(base64.StdEncoding.EncodeToString([]byte(token)))
	req.Header.Set("Authorization", auth)
}

var execCreds CredentialFunc = execCredsCommand

// GetCredentialsFunc returns the current credentials function
func GetCredentialsFunc() CredentialFunc {
	return execCreds
}

// SetCredentialsFunc overrides the default credentials function (which is to call git)
// Returns the previous credentials func
func SetCredentialsFunc(f CredentialFunc) CredentialFunc {
	oldf := execCreds
	execCreds = f
	return oldf
}

// GetOperationForRequest determines the operation type for a http.Request
func GetOperationForRequest(req *http.Request) string {
	operation := "download"
	if req.Method == "POST" || req.Method == "PUT" {
		operation = "upload"
	}
	return operation
}
