package lfs

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

// getCreds gets the credentials for the given request's URL, and sets its
// Authorization header with them using Basic Authentication. This is like
// getCredsForAPI(), but skips checking the LFS url or git remote.
func getCreds(req *http.Request) (Creds, error) {
	if len(req.Header.Get("Authorization")) > 0 {
		return nil, nil
	}

	creds, err := credentials(req.URL)
	if err != nil {
		return nil, err
	}

	setRequestAuth(req, creds["username"], creds["password"])
	return creds, nil
}

// getCredsForAPI gets the credentials for LFS API requests and sets the given
// request's Authorization header with them using Basic Authentication.
// 1. Check the LFS URL for authentication. Ex: http://user:pass@example.com
// 2. Check the Git remote URL for authentication IF it's the same scheme and
//    host of the LFS URL.
// 3. Ask 'git credential' to fill in the password from one of the above URLs.
func getCredsForAPI(req *http.Request) (Creds, error) {
	if len(req.Header.Get("Authorization")) > 0 {
		return nil, nil
	}

	credsUrl, err := getCredURLForAPI(req)
	if err != nil || credsUrl == nil {
		return nil, err
	}

	creds, err := credentials(credsUrl)
	if err != nil {
		return nil, err
	}

	setRequestAuth(req, creds["username"], creds["password"])
	return creds, nil
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

func credentials(u *url.URL) (Creds, error) {
	path := strings.TrimPrefix(u.Path, "/")
	creds := Creds{"protocol": u.Scheme, "host": u.Host, "path": path}
	cmd, err := execCreds(creds, "fill")
	if err != nil {
		return nil, err
	}
	return cmd.Credentials(), nil
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

type credentialCmd struct {
	output     *bytes.Buffer
	SubCommand string
	*exec.Cmd
}

func newCredentialCommand(input Creds, subCommand string) *credentialCmd {
	buf1 := new(bytes.Buffer)
	cmd := exec.Command("git", "credential", subCommand)

	cmd.Stdin = input.Buffer()
	cmd.Stdout = buf1
	/*
		There is a reason we don't hook up stderr here:
		Git's credential cache daemon helper does not close its stderr, so if this
		process is the process that fires up the daemon, it will wait forever
		(until the daemon exits, really) trying to read from stderr.

		See https://github.com/github/git-lfs/issues/117 for more details.
	*/

	return &credentialCmd{buf1, subCommand, cmd}
}

func (c *credentialCmd) Credentials() Creds {
	creds := make(Creds)
	output := c.output.String()

	for _, line := range strings.Split(output, "\n") {
		pieces := strings.SplitN(line, "=", 2)
		if len(pieces) < 2 {
			continue
		}
		creds[pieces[0]] = pieces[1]
	}

	return creds
}

type credentialFetcher interface {
	Credentials() Creds
}

type credentialFunc func(Creds, string) (credentialFetcher, error)

var execCreds credentialFunc

func init() {
	execCreds = func(input Creds, subCommand string) (credentialFetcher, error) {
		cmd := newCredentialCommand(input, subCommand)
		err := cmd.Start()
		if err == nil {
			err = cmd.Wait()
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ProcessState.Success() == false && !Config.GetenvBool("GIT_TERMINAL_PROMPT", true) {
				return nil, fmt.Errorf("Change the GIT_TERMINAL_PROMPT env var to be prompted to enter your credentials for %s://%s.",
					input["protocol"], input["host"])
			}
		}

		if err != nil {
			return cmd, fmt.Errorf("'git credential %s' error: %s\n", cmd.SubCommand, err.Error())
		}

		return cmd, nil
	}
}
