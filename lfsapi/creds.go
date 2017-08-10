package lfsapi

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/config"
	"github.com/rubyist/tracerx"
)

// credsConfig supplies configuration options pertaining to the authorization
// process in package lfsapi.
type credsConfig struct {
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

	var h CredentialHelper
	h = &commandCredentialHelper{
		SkipPrompt: ccfg.SkipPrompt,
	}

	if ccfg.Cached {
		h = withCredentialCache(h)
	}

	return h, nil
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
