package lfsapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

type SSHResolver interface {
	Resolve(Endpoint, string) (sshAuthResponse, error)
}

func withSSHCache(ssh SSHResolver) SSHResolver {
	return &sshCache{
		endpoints: make(map[string]*sshAuthResponse),
		ssh:       ssh,
	}
}

type sshCache struct {
	endpoints map[string]*sshAuthResponse
	ssh       SSHResolver
}

func (c *sshCache) Resolve(e Endpoint, method string) (sshAuthResponse, error) {
	if len(e.SshUserAndHost) == 0 {
		return sshAuthResponse{}, nil
	}

	key := strings.Join([]string{e.SshUserAndHost, e.SshPort, e.SshPath, method}, "//")
	if res, ok := c.endpoints[key]; ok {
		if _, expired := res.IsExpiredWithin(5 * time.Second); !expired {
			tracerx.Printf("ssh cache: %s git-lfs-authenticate %s %s",
				e.SshUserAndHost, e.SshPath, endpointOperation(e, method))
			return *res, nil
		} else {
			tracerx.Printf("ssh cache expired: %s git-lfs-authenticate %s %s",
				e.SshUserAndHost, e.SshPath, endpointOperation(e, method))
		}
	}

	res, err := c.ssh.Resolve(e, method)
	if err == nil {
		c.endpoints[key] = &res
	}
	return res, err
}

type sshAuthResponse struct {
	Message   string            `json:"-"`
	Href      string            `json:"href"`
	Header    map[string]string `json:"header"`
	ExpiresAt time.Time         `json:"expires_at"`
	ExpiresIn int64             `json:"expires_in"`
}

func (r *sshAuthResponse) IsExpiredWithin(d time.Duration) (time.Time, bool) {
	return tools.IsExpiredAtOrIn(d, r.ExpiresAt, time.Duration(r.ExpiresIn)*time.Second)
}

type sshAuthClient struct {
	os Env
}

func (c *sshAuthClient) Resolve(e Endpoint, method string) (sshAuthResponse, error) {
	res := sshAuthResponse{}
	if len(e.SshUserAndHost) == 0 {
		return res, nil
	}

	exe, args := sshGetLFSExeAndArgs(c.os, e, method)
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

func sshGetLFSExeAndArgs(osEnv Env, e Endpoint, method string) (string, []string) {
	operation := endpointOperation(e, method)
	tracerx.Printf("ssh: %s git-lfs-authenticate %s %s",
		e.SshUserAndHost, e.SshPath, operation)

	exe, args := sshGetExeAndArgs(osEnv, e)
	return exe, append(args,
		fmt.Sprintf("git-lfs-authenticate %s %s", e.SshPath, operation))
}

// Return the executable name for ssh on this machine and the base args
// Base args includes port settings, user/host, everything pre the command to execute
func sshGetExeAndArgs(osEnv Env, e Endpoint) (exe string, baseargs []string) {
	isPlink := false
	isTortoise := false

	ssh, _ := osEnv.Get("GIT_SSH")
	sshCmd, _ := osEnv.Get("GIT_SSH_COMMAND")
	cmdArgs := tools.QuotedFields(sshCmd)
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
