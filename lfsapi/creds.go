package lfsapi

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

// credsConfig supplies configuration options pertaining to the authorization
// process in package lfsapi.
type credsConfig struct {
	// AskPass is a string containing an executable name as well as a
	// program arguments.
	//
	// See: https://git-scm.com/docs/gitcredentials#_requesting_credentials
	// for more.
	AskPass string `os:"GIT_ASKPASS" git:"core.askpass" os:"SSH_ASKPASS"`
	// Helper is a string defining the credential helper that Git should use.
	Helper string `git:"credential.helper"`
	// Cached is a boolean determining whether or not to enable the
	// credential cacher.
	Cached bool `git:"lfs.cachecredentials"`
	// SkipPrompt is a boolean determining whether or not to prompt the user
	// for a password.
	SkipPrompt bool `os:"GIT_TERMINAL_PROMPT"`
}

// getCredentialHelper parses a 'credsConfig' from the git and OS environments,
// returning the appropriate CredentialHelper to authenticate requests with.
//
// It returns an error if any configuration was invalid, or otherwise
// un-useable.
func getCredentialHelper(cfg *config.Configuration) (CredentialHelper, error) {
	ccfg, err := getCredentialConfig(cfg)
	if err != nil {
		return nil, err
	}

	var hs []CredentialHelper
	if len(ccfg.Helper) == 0 && len(ccfg.AskPass) > 0 {
		hs = append(hs, &AskPassCredentialHelper{
			Program: ccfg.AskPass,
		})
	}

	var h CredentialHelper
	h = &commandCredentialHelper{
		SkipPrompt: ccfg.SkipPrompt,
	}

	if ccfg.Cached {
		h = withCredentialCache(h)
	}
	hs = append(hs, h)

	switch len(hs) {
	case 0:
		return nil, nil
	case 1:
		return hs[0], nil
	}
	return CredentialHelpers(hs), nil
}

// getCredentialConfig parses a *credsConfig given the OS and Git
// configurations.
func getCredentialConfig(cfg *config.Configuration) (*credsConfig, error) {
	var what credsConfig

	if err := cfg.Unmarshal(&what); err != nil {
		return nil, err
	}
	return &what, nil
}

// CredentialHelpers is a []CredentialHelper that iterates through each
// credential helper to fill, reject, or approve credentials.
type CredentialHelpers []CredentialHelper

// Fill implements CredentialHelper.Fill by asking each CredentialHelper in
// order to fill the credentials.
//
// If a fill was successful, it is returned immediately, and no other
// `CredentialHelper`s are consulted. If any CredentialHelper returns an error,
// it is returned immediately.
func (h CredentialHelpers) Fill(what Creds) (Creds, error) {
	for _, c := range h {
		creds, err := c.Fill(what)
		if err != nil {
			return nil, err
		}

		if creds != nil {
			return creds, nil
		}
	}

	return nil, nil
}

// Reject implements CredentialHelper.Reject and rejects the given Creds "what"
// amongst all knonw CredentialHelpers. If any `CredentialHelper`s returned a
// non-nil error, no further `CredentialHelper`s are notified, so as to prevent
// inconsistent state.
func (h CredentialHelpers) Reject(what Creds) error {
	for _, c := range h {
		if err := c.Reject(what); err != nil {
			return err
		}
	}

	return nil
}

// Approve implements CredentialHelper.Approve and approves the given Creds
// "what" amongst all known CredentialHelpers. If any `CredentialHelper`s
// returned a non-nil error, no further `CredentialHelper`s are notified, so as
// to prevent inconsistent state.
func (h CredentialHelpers) Approve(what Creds) error {
	for _, c := range h {
		if err := c.Approve(what); err != nil {
			return err
		}
	}

	return nil
}

// AskPassCredentialHelper implements the CredentialHelper type for GIT_ASKPASS
// and 'core.askpass' configuration values.
type AskPassCredentialHelper struct {
	// Program is the executable program's absolute or relative name.
	Program string
}

// Fill implements fill by running the ASKPASS program and returning its output
// as a password encoded in the Creds type given the key "password".
//
// It accepts the password as coming from the program's stdout, as when invoked
// with the given arguments (see (*AskPassCredentialHelper).args() below)./
//
// If there was an error running the command, it is returned instead of a set of
// filled credentials.
func (a *AskPassCredentialHelper) Fill(what Creds) (Creds, error) {
	var user bytes.Buffer
	var pass bytes.Buffer
	var err bytes.Buffer

	u := &url.URL{
		Scheme: what["protocol"],
		Host:   what["host"],
		Path:   what["path"],
	}

	// 'ucmd' will run the GIT_ASKPASS (or core.askpass) command prompting
	// for a username.
	ucmd := exec.Command(a.Program, a.args(fmt.Sprintf("Username for %q", u))...)
	ucmd.Stderr = &err
	ucmd.Stdout = &user

	tracerx.Printf("creds: filling with GIT_ASKPASS: %s", strings.Join(ucmd.Args, " "))
	if err := ucmd.Run(); err != nil {
		return nil, err
	}

	if err.Len() > 0 {
		return nil, errors.New(err.String())
	}

	if username := strings.TrimSpace(user.String()); len(username) > 0 {
		// If a non-empty username was given, add it to the URL via func
		// 'net/url.User()'.
		u.User = url.User(username)
	}

	// Regardless, create 'pcmd' to run the GIT_ASKPASS (or core.askpass)
	// command prompting for a password.
	pcmd := exec.Command(a.Program, a.args(fmt.Sprintf("Password for %q", u))...)
	pcmd.Stderr = &err
	pcmd.Stdout = &pass

	tracerx.Printf("creds: filling with GIT_ASKPASS: %s", strings.Join(pcmd.Args, " "))
	if err := pcmd.Run(); err != nil {
		return nil, err
	}

	if err.Len() > 0 {
		return nil, errors.New(err.String())
	}

	// Finally, now that we have the username and password information,
	// store it in the creds instance that we will return to the caller.
	creds := make(Creds)
	creds["username"] = strings.TrimSpace(user.String())
	creds["password"] = strings.TrimSpace(pass.String())

	return creds, nil
}

// Approve implements CredentialHelper.Approve, and returns nil. The ASKPASS
// credential helper does not implement credential approval.
func (a *AskPassCredentialHelper) Approve(_ Creds) error { return nil }

// Reject implements CredentialHelper.Reject, and returns nil. The ASKPASS
// credential helper does not implement credential rejection.
func (a *AskPassCredentialHelper) Reject(_ Creds) error { return nil }

// args returns the arguments given to the ASKPASS program, if a prompt was
// given.

// See: https://git-scm.com/docs/gitcredentials#_requesting_credentials for
// more.
func (a *AskPassCredentialHelper) args(prompt string) []string {
	if len(prompt) == 0 {
		return nil
	}
	return []string{prompt}
}

type CredentialHelper interface {
	Fill(Creds) (Creds, error)
	Reject(Creds) error
	Approve(Creds) error
}

type Creds map[string]string

func bufferCreds(c Creds) *bytes.Buffer {
	buf := new(bytes.Buffer)

	for k, v := range c {
		buf.Write([]byte(k))
		buf.Write([]byte("="))
		buf.Write([]byte(v))
		buf.Write([]byte("\n"))
	}

	return buf
}

func withCredentialCache(helper CredentialHelper) CredentialHelper {
	return &credentialCacher{
		cmu:    new(sync.Mutex),
		creds:  make(map[string]Creds),
		helper: helper,
	}
}

type credentialCacher struct {
	// cmu guards creds
	cmu    *sync.Mutex
	creds  map[string]Creds
	helper CredentialHelper
}

func credCacheKey(creds Creds) string {
	parts := []string{
		creds["protocol"],
		creds["host"],
		creds["path"],
	}
	return strings.Join(parts, "//")
}

func (c *credentialCacher) Fill(creds Creds) (Creds, error) {
	key := credCacheKey(creds)

	c.cmu.Lock()
	defer c.cmu.Unlock()

	if cache, ok := c.creds[key]; ok {
		tracerx.Printf("creds: git credential cache (%q, %q, %q)",
			creds["protocol"], creds["host"], creds["path"])
		return cache, nil
	}

	creds, err := c.helper.Fill(creds)
	if err == nil && len(creds["username"]) > 0 && len(creds["password"]) > 0 {
		c.creds[key] = creds
	}
	return creds, err
}

func (c *credentialCacher) Reject(creds Creds) error {
	c.cmu.Lock()
	defer c.cmu.Unlock()

	delete(c.creds, credCacheKey(creds))
	return c.helper.Reject(creds)
}

func (c *credentialCacher) Approve(creds Creds) error {
	err := c.helper.Approve(creds)
	if err == nil {
		c.cmu.Lock()
		c.creds[credCacheKey(creds)] = creds
		c.cmu.Unlock()
	}
	return err
}

type commandCredentialHelper struct {
	SkipPrompt bool
}

func (h *commandCredentialHelper) Fill(creds Creds) (Creds, error) {
	tracerx.Printf("creds: git credential fill (%q, %q, %q)",
		creds["protocol"], creds["host"], creds["path"])
	return h.exec("fill", creds)
}

func (h *commandCredentialHelper) Reject(creds Creds) error {
	_, err := h.exec("reject", creds)
	return err
}

func (h *commandCredentialHelper) Approve(creds Creds) error {
	_, err := h.exec("approve", creds)
	return err
}

func (h *commandCredentialHelper) exec(subcommand string, input Creds) (Creds, error) {
	output := new(bytes.Buffer)
	cmd := exec.Command("git", "credential", subcommand)
	cmd.Stdin = bufferCreds(input)
	cmd.Stdout = output
	/*
	   There is a reason we don't hook up stderr here:
	   Git's credential cache daemon helper does not close its stderr, so if this
	   process is the process that fires up the daemon, it will wait forever
	   (until the daemon exits, really) trying to read from stderr.

	   See https://github.com/git-lfs/git-lfs/issues/117 for more details.
	*/

	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if _, ok := err.(*exec.ExitError); ok {
		if h.SkipPrompt {
			return nil, fmt.Errorf("Change the GIT_TERMINAL_PROMPT env var to be prompted to enter your credentials for %s://%s.",
				input["protocol"], input["host"])
		}

		// 'git credential' exits with 128 if the helper doesn't fill the username
		// and password values.
		if subcommand == "fill" && err.Error() == "exit status 128" {
			return nil, nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("'git credential %s' error: %s\n", subcommand, err.Error())
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
