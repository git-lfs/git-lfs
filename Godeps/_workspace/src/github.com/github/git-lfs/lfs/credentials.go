package lfs

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

type credentialFetcher interface {
	Credentials() Creds
}

type credentialFunc func(Creds, string) (credentialFetcher, error)

var execCreds credentialFunc

func credentials(u *url.URL) (Creds, error) {
	creds := Creds{"protocol": u.Scheme, "host": u.Host}
	cmd, err := execCreds(creds, "fill")
	if err != nil {
		return nil, err
	}
	return cmd.Credentials(), nil
}

type CredentialCmd struct {
	output     *bytes.Buffer
	SubCommand string
	*exec.Cmd
}

func NewCommand(input Creds, subCommand string) *CredentialCmd {
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

	return &CredentialCmd{buf1, subCommand, cmd}
}

func (c *CredentialCmd) StdoutString() string {
	return c.output.String()
}

func (c *CredentialCmd) Credentials() Creds {
	creds := make(Creds)

	for _, line := range strings.Split(c.StdoutString(), "\n") {
		pieces := strings.SplitN(line, "=", 2)
		if len(pieces) < 2 {
			continue
		}
		creds[pieces[0]] = pieces[1]
	}

	return creds
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

func init() {
	execCreds = func(input Creds, subCommand string) (credentialFetcher, error) {
		cmd := NewCommand(input, subCommand)
		err := cmd.Start()
		if err == nil {
			err = cmd.Wait()
		}

		if err != nil {
			return cmd, fmt.Errorf("'git credential %s' error: %s\n", cmd.SubCommand, err.Error())
		}

		return cmd, nil
	}
}
