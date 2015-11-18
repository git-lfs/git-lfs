package lfs

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

// getCreds gets the credentials for the given request's URL, and sets its
// Authorization header with them using Basic Authentication. This is like
// getCredsForAPI(), but skips checking the LFS url or git remote.
func getCreds(req *http.Request) (Creds, error) {
	if skipCredsCheck(req) {
		return nil, nil
	}

	return fillCredentials(req, req.URL)
}

// getCredsForAPI gets the credentials for LFS API requests and sets the given
// request's Authorization header with them using Basic Authentication.
// 1. Check the LFS URL for authentication. Ex: http://user:pass@example.com
// 2. Check netrc for authentication.
// 3. Check the Git remote URL for authentication IF it's the same scheme and
//    host of the LFS URL.
// 4. Ask 'git credential' to fill in the password from one of the above URLs.
//
// This prefers the Git remote URL for checking credentials so that users only
// have to enter their passwords once for Git and Git LFS. It uses the same
// URL path that Git does, in case 'useHttpPath' is enabled in the Git config.
func getCredsForAPI(req *http.Request) (Creds, error) {
	if skipCredsCheck(req) {
		return nil, nil
	}

	credsUrl, err := getCredURLForAPI(req)
	if err != nil {
		return nil, Error(err)
	}

	if credsUrl == nil {
		return nil, nil
	}

	if setCredURLFromNetrc(req) {
		return nil, nil
	}

	return fillCredentials(req, credsUrl)
}

func getCredURLForAPI(req *http.Request) (*url.URL, error) {
	apiUrl, err := Config.ObjectUrl("")
	if err != nil {
		return nil, err
	}

	// if the LFS request doesn't match the current LFS url, don't bother
	// attempting to set the Authorization header from the LFS or Git remote URLs.
	if req.URL.Scheme != apiUrl.Scheme ||
		req.URL.Host != apiUrl.Host {
		return req.URL, nil
	}

	if setRequestAuthFromUrl(req, apiUrl) {
		return nil, nil
	}

	credsUrl := apiUrl
	if len(Config.CurrentRemote) > 0 {
		if u, ok := Config.GitConfig("remote." + Config.CurrentRemote + ".url"); ok {
			gitRemoteUrl, err := url.Parse(u)
			if err != nil {
				return nil, err
			}

			if gitRemoteUrl.Scheme == apiUrl.Scheme &&
				gitRemoteUrl.Host == apiUrl.Host {

				if setRequestAuthFromUrl(req, gitRemoteUrl) {
					return nil, nil
				}

				credsUrl = gitRemoteUrl
			}
		}
	}
	return credsUrl, nil
}

func setCredURLFromNetrc(req *http.Request) bool {
	host, _, err := net.SplitHostPort(req.URL.Host)
	if err != nil {
		return false
	}

	machine, err := Config.FindNetrcHost(host)
	if err != nil || machine == nil {
		return false
	}

	setRequestAuth(req, machine.Login, machine.Password)
	return true
}

func skipCredsCheck(req *http.Request) bool {
	if Config.NtlmAccess() {
		return false
	}

	if len(req.Header.Get("Authorization")) > 0 {
		return true
	}

	q := req.URL.Query()
	return len(q["token"]) > 0
}

func fillCredentials(req *http.Request, u *url.URL) (Creds, error) {
	path := strings.TrimPrefix(u.Path, "/")
	input := Creds{"protocol": u.Scheme, "host": u.Host, "path": path}
	if u.User != nil && u.User.Username() != "" {
		input["username"] = u.User.Username()
	}

	creds, err := execCreds(input, "fill")
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

func saveCredentials(creds Creds, res *http.Response) {
	if creds == nil {
		return
	}

	switch res.StatusCode {
	case 401, 403:
		execCreds(creds, "reject")
	default:
		if res.StatusCode < 300 {
			execCreds(creds, "approve")
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

type credentialFunc func(Creds, string) (Creds, error)

func execCredsCommand(input Creds, subCommand string) (Creds, error) {
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
		if !Config.GetenvBool("GIT_TERMINAL_PROMPT", true) {
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

var execCreds credentialFunc = execCredsCommand
