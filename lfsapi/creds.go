package lfsapi

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

// CredentialHelper is an interface used by the lfsapi Client to interact with
// the 'git credential' command: https://git-scm.com/docs/gitcredentials
// Other implementations include ASKPASS support, and an in-memory cache.
type CredentialHelper interface {
	Fill(Creds) (Creds, error)
	Reject(Creds) error
	Approve(Creds) error
}

// Creds represents a set of key/value pairs that are passed to 'git credential'
// as input.
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

// getCredentialHelper parses a 'credsConfig' from the git and OS environments,
// returning the appropriate CredentialHelper to authenticate requests with.
//
// It returns an error if any configuration was invalid, or otherwise
// un-useable.
func (c *Client) getCredentialHelper(u *url.URL) (CredentialHelper, Creds) {
	rawurl := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	input := Creds{"protocol": u.Scheme, "host": u.Host}
	if u.User != nil && u.User.Username() != "" {
		input["username"] = u.User.Username()
	}
	if c.uc.Bool("credential", rawurl, "usehttppath", false) {
		input["path"] = strings.TrimPrefix(u.Path, "/")
	}

	if c.Credentials != nil {
		return c.Credentials, input
	}

	helpers := make([]CredentialHelper, 0, 3)
	if c.cachingCredHelper != nil {
		helpers = append(helpers, c.cachingCredHelper)
	}
	if c.askpassCredHelper != nil {
		helper, _ := c.uc.Get("credential", rawurl, "helper")
		if len(helper) == 0 {
			helpers = append(helpers, c.askpassCredHelper)
		}
	}

	return NewCredentialHelpers(append(helpers, c.commandCredHelper)), input
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
	tracerx.Printf("creds: git credential approve (%q, %q, %q)",
		creds["protocol"], creds["host"], creds["path"])
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

type credentialCacher struct {
	creds map[string]Creds
	mu    sync.Mutex
}

func newCredentialCacher() *credentialCacher {
	return &credentialCacher{creds: make(map[string]Creds)}
}

func credCacheKey(creds Creds) string {
	parts := []string{
		creds["protocol"],
		creds["host"],
		creds["path"],
	}
	return strings.Join(parts, "//")
}

func (c *credentialCacher) Fill(what Creds) (Creds, error) {
	key := credCacheKey(what)
	c.mu.Lock()
	cached, ok := c.creds[key]
	c.mu.Unlock()

	if ok {
		tracerx.Printf("creds: git credential cache (%q, %q, %q)",
			what["protocol"], what["host"], what["path"])
		return cached, nil
	}

	return nil, credHelperNoOp
}

func (c *credentialCacher) Approve(what Creds) error {
	key := credCacheKey(what)

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.creds[key]; ok {
		return nil
	}

	c.creds[key] = what
	return credHelperNoOp
}

func (c *credentialCacher) Reject(what Creds) error {
	key := credCacheKey(what)
	c.mu.Lock()
	delete(c.creds, key)
	c.mu.Unlock()
	return credHelperNoOp
}

// CredentialHelpers iterates through a slice of CredentialHelper objects
// CredentialHelpers is a []CredentialHelper that iterates through each
// credential helper to fill, reject, or approve credentials. Typically, the
// first success returns immediately. Errors are reported to tracerx, unless
// all credential helpers return errors. Any erroring credential helpers are
// skipped for future calls.
//
// A CredentialHelper can return a credHelperNoOp error, signaling that the
// CredentialHelpers should try the next one.
type CredentialHelpers struct {
	helpers        []CredentialHelper
	skippedHelpers map[int]bool
	mu             sync.Mutex
}

// NewCredentialHelpers initializes a new CredentialHelpers from the given
// slice of CredentialHelper instances.
func NewCredentialHelpers(helpers []CredentialHelper) CredentialHelper {
	return &CredentialHelpers{
		helpers:        helpers,
		skippedHelpers: make(map[int]bool),
	}
}

var credHelperNoOp = errors.New("no-op!")

// Fill implements CredentialHelper.Fill by asking each CredentialHelper in
// order to fill the credentials.
//
// If a fill was successful, it is returned immediately, and no other
// `CredentialHelper`s are consulted. If any CredentialHelper returns an error,
// it is reported to tracerx, and the next one is attempted. If they all error,
// then a collection of all the error messages is returned. Erroring credential
// helpers are added to the skip list, and never attempted again for the
// lifetime of the current Git LFS command.
func (s *CredentialHelpers) Fill(what Creds) (Creds, error) {
	errs := make([]string, 0, len(s.helpers))
	for i, h := range s.helpers {
		if s.skipped(i) {
			continue
		}

		creds, err := h.Fill(what)
		if err != nil {
			if err != credHelperNoOp {
				s.skip(i)
				tracerx.Printf("credential fill error: %s", err)
				errs = append(errs, err.Error())
			}
			continue
		}

		if creds != nil {
			return creds, nil
		}
	}

	if len(errs) > 0 {
		return nil, errors.New("credential fill errors:\n" + strings.Join(errs, "\n"))
	}

	return nil, nil
}

// Reject implements CredentialHelper.Reject and rejects the given Creds "what"
// with the first successful attempt.
func (s *CredentialHelpers) Reject(what Creds) error {
	for i, h := range s.helpers {
		if s.skipped(i) {
			continue
		}

		if err := h.Reject(what); err != credHelperNoOp {
			return err
		}
	}

	return errors.New("no valid credential helpers to reject")
}

// Approve implements CredentialHelper.Approve and approves the given Creds
// "what" with the first successful CredentialHelper. If an error occurrs,
// it calls Reject() with the same Creds and returns the error immediately. This
// ensures a caching credential helper removes the cache, since the Erroring
// CredentialHelper never successfully saved it.
func (s *CredentialHelpers) Approve(what Creds) error {
	skipped := make(map[int]bool)
	for i, h := range s.helpers {
		if s.skipped(i) {
			skipped[i] = true
			continue
		}

		if err := h.Approve(what); err != credHelperNoOp {
			if err != nil && i > 0 { // clear any cached approvals
				for j := 0; j < i; j++ {
					if !skipped[j] {
						s.helpers[j].Reject(what)
					}
				}
			}
			return err
		}
	}

	return errors.New("no valid credential helpers to approve")
}

func (s *CredentialHelpers) skip(i int) {
	s.mu.Lock()
	s.skippedHelpers[i] = true
	s.mu.Unlock()
}

func (s *CredentialHelpers) skipped(i int) bool {
	s.mu.Lock()
	skipped := s.skippedHelpers[i]
	s.mu.Unlock()
	return skipped
}

type nullCredentialHelper struct{}

var (
	nullCredError = errors.New("No credential helper configured")
	nullCreds     = &nullCredentialHelper{}
)

func (h *nullCredentialHelper) Fill(input Creds) (Creds, error) {
	return nil, nullCredError
}

func (h *nullCredentialHelper) Approve(creds Creds) error {
	return nil
}

func (h *nullCredentialHelper) Reject(creds Creds) error {
	return nil
}
