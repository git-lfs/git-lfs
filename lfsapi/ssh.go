package lfsapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/config"
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
	ExpiresIn int               `json:"expires_in"`

	createdAt time.Time
}

func (r *sshAuthResponse) IsExpiredWithin(d time.Duration) (time.Time, bool) {
	return tools.IsExpiredAtOrIn(r.createdAt, d, r.ExpiresAt, time.Duration(r.ExpiresIn)*time.Second)
}

type sshAuthClient struct {
	os config.Environment
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

	now := time.Now()

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
		res.createdAt = now
	}

	return res, err
}

func sshGetLFSExeAndArgs(osEnv config.Environment, e Endpoint, method string) (string, []string) {
	exe, args := sshGetExeAndArgs(osEnv, e)
	operation := endpointOperation(e, method)
	args = append(args, fmt.Sprintf("git-lfs-authenticate %s %s", e.SshPath, operation))
	tracerx.Printf("run_command: %s %s", exe, strings.Join(args, " "))
	return exe, args
}

// Return the executable name for ssh on this machine and the base args
// Base args includes port settings, user/host, everything pre the command to execute
func sshGetExeAndArgs(osEnv config.Environment, e Endpoint) (exe string, baseargs []string) {
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
		ssh = defaultSSHCmd
	}

	basessh := filepath.Base(ssh)

	if basessh != defaultSSHCmd {
		// Strip extension for easier comparison
		if ext := filepath.Ext(basessh); len(ext) > 0 {
			basessh = basessh[:len(basessh)-len(ext)]
		}
		isPlink = strings.EqualFold(basessh, "plink")
		isTortoise = strings.EqualFold(basessh, "tortoiseplink")
	}

	args := make([]string, 0, 5+len(cmdArgs))
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

	if sep, ok := sshSeparators[basessh]; ok {
		// inserts a separator between cli -options and host/cmd commands
		// example: $ ssh -p 12345 -- user@host.com git-lfs-authenticate ...
		args = append(args, sep, e.SshUserAndHost)
	} else {
		// no prefix supported, strip leading - off host to prevent cmd like:
		// $ git config lfs.url ssh://-proxycmd=whatever
		// $ plink -P 12345 -proxycmd=foo git-lfs-authenticate ...
		//
		// Instead, it'll attempt this, and eventually return an error
		// $ plink -P 12345 proxycmd=foo git-lfs-authenticate ...
		args = append(args, sshOptPrefixRE.ReplaceAllString(e.SshUserAndHost, ""))
	}

	return ssh, args
}

const defaultSSHCmd = "ssh"

var (
	sshOptPrefixRE = regexp.MustCompile(`\A\-+`)
	sshSeparators  = map[string]string{
		"ssh":          "--",
		"lfs-ssh-echo": "--", // used in lfs integration tests only
	}
)
