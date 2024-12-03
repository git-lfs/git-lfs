package creds

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

// CredentialHelperWrapper is used to contain the encapsulate the information we need for credential handling during auth.
type CredentialHelperWrapper struct {
	CredentialHelper CredentialHelper
	Input            Creds
	Url              *url.URL
	Creds            Creds
}

// CredentialHelper is an interface used by the lfsapi Client to interact with
// the 'git credential' command: https://git-scm.com/docs/gitcredentials
// Other implementations include ASKPASS support, and an in-memory cache.
type CredentialHelper interface {
	Fill(Creds) (Creds, error)
	Reject(Creds) error
	Approve(Creds) error
}

func (credWrapper *CredentialHelperWrapper) FillCreds() error {
	creds, err := credWrapper.CredentialHelper.Fill(credWrapper.Input)
	if creds == nil || len(creds) < 1 {
		errmsg := tr.Tr.Get("Git credentials for %s not found", credWrapper.Url)
		if err != nil {
			errmsg = fmt.Sprintf("%s:\n%s", errmsg, err.Error())
		} else {
			errmsg = fmt.Sprintf("%s.", errmsg)
		}
		err = errors.New(errmsg)
	}
	credWrapper.Creds = creds
	return err
}

// Creds represents a set of key/value pairs that are passed to 'git credential'
// as input.
type Creds map[string][]string

func (c Creds) IsMultistage() bool {
	return slices.Contains([]string{"1", "true"}, FirstEntryForKey(c, "continue"))
}

func (c Creds) buffer(protectProtocol bool) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	buf.Write([]byte("capability[]=authtype\n"))
	buf.Write([]byte("capability[]=state\n"))
	for k, v := range c {
		for _, item := range v {
			if strings.Contains(item, "\n") {
				return nil, errors.Errorf(tr.Tr.Get("credential value for %s contains newline: %q", k, item))
			}
			if protectProtocol && strings.Contains(item, "\r") {
				return nil, errors.Errorf(tr.Tr.Get("credential value for %s contains carriage return: %q\nIf this is intended, set `credential.protectProtocol=false`", k, item))
			}
			if strings.Contains(item, string(rune(0))) {
				return nil, errors.Errorf(tr.Tr.Get("credential value for %s contains null byte: %q", k, item))
			}

			buf.Write([]byte(k))
			buf.Write([]byte("="))
			buf.Write([]byte(item))
			buf.Write([]byte("\n"))
		}
	}

	return buf, nil
}

type CredentialHelperContext struct {
	netrcCredHelper   *netrcCredentialHelper
	commandCredHelper *commandCredentialHelper
	askpassCredHelper *AskPassCredentialHelper
	cachingCredHelper *credentialCacher

	urlConfig      *config.URLConfig
	wwwAuthHeaders []string
	state          []string
}

func NewCredentialHelperContext(gitEnv config.Environment, osEnv config.Environment) *CredentialHelperContext {
	c := &CredentialHelperContext{urlConfig: config.NewURLConfig(gitEnv)}

	c.netrcCredHelper = newNetrcCredentialHelper(osEnv)

	askpass, ok := osEnv.Get("GIT_ASKPASS")
	if !ok {
		askpass, ok = gitEnv.Get("core.askpass")
	}
	if !ok {
		askpass, _ = osEnv.Get("SSH_ASKPASS")
	}
	if len(askpass) > 0 {
		askpassfile, err := tools.TranslateCygwinPath(askpass)
		if err != nil {
			tracerx.Printf("Error reading askpass helper %q: %v", askpassfile, err)
		}
		if len(askpassfile) > 0 {
			c.askpassCredHelper = &AskPassCredentialHelper{
				Program: askpassfile,
			}
		}
	}

	cacheCreds := gitEnv.Bool("lfs.cachecredentials", true)
	if cacheCreds {
		c.cachingCredHelper = NewCredentialCacher()
	}

	c.commandCredHelper = &commandCredentialHelper{
		SkipPrompt: osEnv.Bool("GIT_TERMINAL_PROMPT", false),
	}

	return c
}

func (ctxt *CredentialHelperContext) SetWWWAuthHeaders(headers []string) {
	ctxt.wwwAuthHeaders = headers
}

func (ctxt *CredentialHelperContext) SetStateFields(fields []string) {
	ctxt.state = fields
}

// getCredentialHelper parses a 'credsConfig' from the git and OS environments,
// returning the appropriate CredentialHelper to authenticate requests with.
//
// It returns an error if any configuration was invalid, or otherwise
// un-useable.
func (ctxt *CredentialHelperContext) GetCredentialHelper(helper CredentialHelper, u *url.URL) CredentialHelperWrapper {
	rawurl := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	input := Creds{"protocol": []string{u.Scheme}, "host": []string{u.Host}}
	if u.User != nil && u.User.Username() != "" {
		input["username"] = []string{u.User.Username()}
	}
	if u.Scheme == "cert" || ctxt.urlConfig.Bool("credential", rawurl, "usehttppath", false) {
		input["path"] = []string{strings.TrimPrefix(u.Path, "/")}
	}
	if len(ctxt.wwwAuthHeaders) != 0 && !ctxt.urlConfig.Bool("credential", rawurl, "skipwwwauth", false) {
		input["wwwauth[]"] = ctxt.wwwAuthHeaders
	}
	if len(ctxt.state) != 0 {
		input["state[]"] = ctxt.state
	}

	if helper != nil {
		return CredentialHelperWrapper{CredentialHelper: helper, Input: input, Url: u}
	}

	helpers := make([]CredentialHelper, 0, 4)
	if ctxt.netrcCredHelper != nil {
		helpers = append(helpers, ctxt.netrcCredHelper)
	}
	if ctxt.cachingCredHelper != nil {
		helpers = append(helpers, ctxt.cachingCredHelper)
	}
	if ctxt.askpassCredHelper != nil {
		helper, _ := ctxt.urlConfig.Get("credential", rawurl, "helper")
		if len(helper) == 0 {
			helpers = append(helpers, ctxt.askpassCredHelper)
		}
	}

	ctxt.commandCredHelper.protectProtocol = ctxt.urlConfig.Bool("credential", rawurl, "protectProtocol", true)

	return CredentialHelperWrapper{CredentialHelper: NewCredentialHelpers(append(helpers, ctxt.commandCredHelper)), Input: input, Url: u}
}

// AskPassCredentialHelper implements the CredentialHelper type for GIT_ASKPASS
// and 'core.askpass' configuration values.
type AskPassCredentialHelper struct {
	// Program is the executable program's absolute or relative name.
	Program string
}

type credValueType int

const (
	credValueTypeUnknown credValueType = iota
	credValueTypeUsername
	credValueTypePassword
)

// Fill implements fill by running the ASKPASS program and returning its output
// as a password encoded in the Creds type given the key "password".
//
// It accepts the password as coming from the program's stdout, as when invoked
// with the given arguments (see (*AskPassCredentialHelper).args() below)./
//
// If there was an error running the command, it is returned instead of a set of
// filled credentials.
//
// The ASKPASS program is only queried if a credential was not already
// provided, i.e. through the git URL
func (a *AskPassCredentialHelper) Fill(what Creds) (Creds, error) {
	u := &url.URL{
		Scheme: FirstEntryForKey(what, "protocol"),
		Host:   FirstEntryForKey(what, "host"),
		Path:   FirstEntryForKey(what, "path"),
	}

	creds := make(Creds)

	username, err := a.getValue(what, credValueTypeUsername, u)
	if err != nil {
		return nil, err
	}
	creds["username"] = []string{username}

	if len(username) > 0 {
		// If a non-empty username was given, add it to the URL via func
		// 'net/url.User()'.
		u.User = url.User(username)
	}

	password, err := a.getValue(what, credValueTypePassword, u)
	if err != nil {
		return nil, err
	}
	creds["password"] = []string{password}

	return creds, nil
}

func (a *AskPassCredentialHelper) getValue(what Creds, valueType credValueType, u *url.URL) (string, error) {
	var valueString string

	switch valueType {
	case credValueTypeUsername:
		valueString = "username"
	case credValueTypePassword:
		valueString = "password"
	default:
		return "", errors.Errorf(tr.Tr.Get("Invalid Credential type queried from AskPass"))
	}

	// Return the existing credential if it was already provided, otherwise
	// query AskPass for it
	if given, ok := what[valueString]; ok && len(given) > 0 {
		return given[0], nil
	}
	return a.getFromProgram(valueType, u)
}

func (a *AskPassCredentialHelper) getFromProgram(valueType credValueType, u *url.URL) (string, error) {
	var (
		value bytes.Buffer
		err   bytes.Buffer

		valueString string
	)

	switch valueType {
	case credValueTypeUsername:
		valueString = "Username"
	case credValueTypePassword:
		valueString = "Password"
	default:
		return "", errors.Errorf(tr.Tr.Get("Invalid Credential type queried from AskPass"))
	}

	// 'cmd' will run the GIT_ASKPASS (or core.askpass) command prompting
	// for the desired valueType (`Username` or `Password`)
	cmd, errVal := subprocess.ExecCommand(a.Program, a.args(fmt.Sprintf("%s for %q", valueString, u))...)
	if errVal != nil {
		tracerx.Printf("creds: failed to find GIT_ASKPASS command: %s", a.Program)
		return "", errVal
	}
	cmd.Stderr = &err
	cmd.Stdout = &value

	tracerx.Printf("creds: filling with GIT_ASKPASS: %s", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		return "", err
	}

	if err.Len() > 0 {
		return "", errors.New(err.String())
	}

	return strings.TrimSpace(value.String()), nil
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
	SkipPrompt      bool
	protectProtocol bool
}

func (h *commandCredentialHelper) Fill(creds Creds) (Creds, error) {
	tracerx.Printf("creds: git credential fill (%q, %q, %q)",
		FirstEntryForKey(creds, "protocol"),
		FirstEntryForKey(creds, "host"),
		FirstEntryForKey(creds, "path"))
	return h.exec("fill", creds)
}

func (h *commandCredentialHelper) Reject(creds Creds) error {
	_, err := h.exec("reject", creds)
	return err
}

func (h *commandCredentialHelper) Approve(creds Creds) error {
	tracerx.Printf("creds: git credential approve (%q, %q, %q)",
		FirstEntryForKey(creds, "protocol"),
		FirstEntryForKey(creds, "host"),
		FirstEntryForKey(creds, "path"))
	_, err := h.exec("approve", creds)
	return err
}

func (h *commandCredentialHelper) exec(subcommand string, input Creds) (Creds, error) {
	output := new(bytes.Buffer)
	cmd, err := subprocess.ExecCommand("git", "credential", subcommand)
	if err != nil {
		return nil, errors.New(tr.Tr.Get("failed to find `git credential %s`: %v", subcommand, err))
	}
	cmd.Stdin, err = input.buffer(h.protectProtocol)
	if err != nil {
		return nil, errors.New(tr.Tr.Get("invalid input to `git credential %s`: %v", subcommand, err))
	}
	cmd.Stdout = output
	/*
	   There is a reason we don't read from stderr here:
	   Git's credential cache daemon helper does not close its stderr, so if this
	   process is the process that fires up the daemon, it will wait forever
	   (until the daemon exits, really) trying to read from stderr.

	   Instead, we simply pass it through to our stderr.

	   See https://github.com/git-lfs/git-lfs/issues/117 for more details.
	*/
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if _, ok := err.(*exec.ExitError); ok {
		if h.SkipPrompt {
			return nil, errors.New(tr.Tr.Get("change the GIT_TERMINAL_PROMPT env var to be prompted to enter your credentials for %s://%s",
				FirstEntryForKey(input, "protocol"), FirstEntryForKey(input, "host")))
		}

		// 'git credential' exits with 128 if the helper doesn't fill the username
		// and password values.
		if subcommand == "fill" && err.Error() == "exit status 128" {
			return nil, nil
		}
	}

	if err != nil {
		return nil, errors.New(tr.Tr.Get("`git credential %s` error: %s", subcommand, err.Error()))
	}

	creds := make(Creds)
	for _, line := range strings.Split(output.String(), "\n") {
		pieces := strings.SplitN(line, "=", 2)
		if len(pieces) < 2 || len(pieces[1]) < 1 {
			continue
		}
		if _, ok := creds[pieces[0]]; ok {
			creds[pieces[0]] = append(creds[pieces[0]], pieces[1])
		} else {
			creds[pieces[0]] = []string{pieces[1]}
		}
	}

	return creds, nil
}

type credentialCacher struct {
	creds map[string]Creds
	mu    sync.Mutex
}

func NewCredentialCacher() *credentialCacher {
	return &credentialCacher{creds: make(map[string]Creds)}
}

func credCacheKey(creds Creds) string {
	parts := []string{
		FirstEntryForKey(creds, "protocol"),
		FirstEntryForKey(creds, "host"),
		FirstEntryForKey(creds, "path"),
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
			FirstEntryForKey(what, "protocol"),
			FirstEntryForKey(what, "host"),
			FirstEntryForKey(what, "path"))
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
		return nil, errors.New(tr.Tr.Get("credential fill errors:\n%s", strings.Join(errs, "\n")))
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

	return errors.New(tr.Tr.Get("no valid credential helpers to reject"))
}

// Approve implements CredentialHelper.Approve and approves the given Creds
// "what" with the first successful CredentialHelper. If an error occurs,
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

	return errors.New(tr.Tr.Get("no valid credential helpers to approve"))
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
	nullCredError = errors.New(tr.Tr.Get("No credential helper configured"))
	NullCreds     = &nullCredentialHelper{}
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

// FirstEntryForKey extracts and returns the first entry for a given key, or
// returns the empty string if no value for that key is available.
func FirstEntryForKey(input Creds, key string) string {
	if val, ok := input[key]; ok && len(val) > 0 {
		return val[0]
	}
	return ""
}
