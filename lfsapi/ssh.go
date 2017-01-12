package lfsapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rubyist/tracerx"
)

func (c *Client) resolveSSHEndpoint(e Endpoint, method string) (sshAuthResponse, error) {
	res := sshAuthResponse{}
	if len(e.SshUserAndHost) == 0 {
		return res, nil
	}

	operation := "upload"
	switch method {
	case "GET", "HEAD":
		operation = "download"
	}

	tracerx.Printf("ssh: %s git-lfs-authenticate %s %s",
		e.SshUserAndHost, e.SshPath, operation)

	exe, args := sshGetExeAndArgs(c.osEnv, e)
	args = append(args,
		fmt.Sprintf("git-lfs-authenticate %s %s", e.SshPath, operation))

	cmd := exec.Command(exe, args...)

	// Save stdout and stderr in separate buffers
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	// Execute command
	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	// Processing result
	if err != nil {
		res.Message = strings.TrimSpace(errbuf.String())
	} else {
		err = json.Unmarshal(outbuf.Bytes(), &res)
	}

	return res, err
}

type sshAuthResponse struct {
	Message   string            `json:"-"`
	Href      string            `json:"href"`
	Header    map[string]string `json:"header"`
	ExpiresAt string            `json:"expires_at"`
}

// Return the executable name for ssh on this machine and the base args
// Base args includes port settings, user/host, everything pre the command to execute
func sshGetExeAndArgs(osEnv Env, e Endpoint) (exe string, baseargs []string) {
	isPlink := false
	isTortoise := false

	ssh, _ := osEnv.Get("GIT_SSH")
	sshCmd, _ := osEnv.Get("GIT_SSH_COMMAND")
	cmdArgs := strings.Fields(sshCmd)
	if len(cmdArgs) > 0 {
		ssh = cmdArgs[0]
		cmdArgs = cmdArgs[1:]
	}

	if ssh == "" {
		ssh = "ssh"
	} else {
		basessh := filepath.Base(ssh)
		// Strip extension for easier comparison
		if ext := filepath.Ext(basessh); len(ext) > 0 {
			basessh = basessh[:len(basessh)-len(ext)]
		}
		isPlink = strings.EqualFold(basessh, "plink")
		isTortoise = strings.EqualFold(basessh, "tortoiseplink")
	}

	args := make([]string, 0, 4+len(cmdArgs))
	if len(cmdArgs) > 0 {
		args = append(args, cmdArgs...)
	}

	if isTortoise {
		// TortoisePlink requires the -batch argument to behave like ssh/plink
		args = append(args, "-batch")
	}

	if len(e.SshPort) > 0 {
		if isPlink || isTortoise {
			args = append(args, "-P")
		} else {
			args = append(args, "-p")
		}
		args = append(args, e.SshPort)
	}
	args = append(args, e.SshUserAndHost)

	return ssh, args
}
