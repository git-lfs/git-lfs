package gitmediaclient

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

func credentials(u *url.URL) (map[string]string, error) {
	credInput := fmt.Sprintf("protocol=%s\nhost=%s\n", u.Scheme, u.Host)
	cmd, err := execCreds(credInput, "fill")
	if err != nil {
		return nil, err
	}
	return cmd.Credentials(), nil
}

func execCreds(input, subCommand string) (*CredentialCmd, error) {
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

func NewCommand(input, subCommand string) *CredentialCmd {
	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)
	cmd := exec.Command("git", "credential", subCommand)
	cmd.Stdin = bytes.NewBufferString(input)
	cmd.Stdout = buf1
	cmd.Stderr = buf2
	return &CredentialCmd{buf1, buf2, subCommand, cmd}
}

func (c *CredentialCmd) StderrString() string {
	return c.err.String()
}

func (c *CredentialCmd) StdoutString() string {
	return c.output.String()
}

func (c *CredentialCmd) Credentials() map[string]string {
	creds := make(map[string]string)

	for _, line := range strings.Split(c.StdoutString(), "\n") {
		pieces := strings.SplitN(line, "=", 2)
		if len(pieces) < 2 {
			continue
		}
		creds[pieces[0]] = pieces[1]
	}

	return creds
}
