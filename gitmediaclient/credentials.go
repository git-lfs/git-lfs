package gitmediaclient

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

func credentials(u *url.URL) (Creds, error) {
	creds := Creds{"protocol": u.Scheme, "host": u.Host}
	cmd, err := execCreds(creds, "fill")
	if err != nil {
		return nil, err
	}
	return cmd.Credentials(), nil
}

func execCreds(input Creds, subCommand string) (*CredentialCmd, error) {
	cmd := NewCommand(input, subCommand)
	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if err != nil {
		return cmd, fmt.Errorf("'git credential %s' error: %s\n%s", cmd.SubCommand, err.Error(), cmd.StderrString())
	}

	return cmd, nil
}

type CredentialCmd struct {
	output     *bytes.Buffer
	err        *bytes.Buffer
	SubCommand string
	*exec.Cmd
}

func NewCommand(input Creds, subCommand string) *CredentialCmd {
	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)
	cmd := exec.Command("git", "credential", subCommand)
	cmd.Stdin = input.Buffer()

	if commandHasOutput(subCommand) {
		cmd.Stdout = buf1
		cmd.Stderr = buf2
	}

	return &CredentialCmd{buf1, buf2, subCommand, cmd}
}

func (c *CredentialCmd) StderrString() string {
	return c.err.String()
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

// commandHasOutput returns true if the command that's being run
// produces output. Of the three current subcommands `fill`, `approve`,
// and `reject`, only `fill` produces output. There is a bug in the way
// the git credential helpers launch the daemon if it is not already running
// such that the stderr of the grandchild does not appear to be getting closed,
// causing the git media client to not receive EOF on the pipe and wait forever.
func commandHasOutput(command string) bool {
	return command == "fill"
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
