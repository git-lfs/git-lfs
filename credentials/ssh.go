package credentials

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

type SshAuthResponse struct {
	Message   string            `json:"-"`
	Href      string            `json:"href"`
	Header    map[string]string `json:"header"`
	ExpiresAt string            `json:"expires_at"`
}

func SshAuthenticate(endpoint config.Endpoint, operation, oid string) (SshAuthResponse, error) {

	// This is only used as a fallback where the Git URL is SSH but server doesn't support a full SSH binary protocol
	// and therefore we derive a HTTPS endpoint for binaries instead; but check authentication here via SSH

	res := SshAuthResponse{}
	if len(endpoint.SshUserAndHost) == 0 {
		return res, nil
	}

	tracerx.Printf("ssh: %s git-lfs-authenticate %s %s %s",
		endpoint.SshUserAndHost, endpoint.SshPath, operation, oid)

	exe, args := sshGetExeAndArgs(endpoint)
	args = append(args,
		fmt.Sprintf("git-lfs-authenticate %s %s %s", endpoint.SshPath, operation, oid))

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
		res.Message = errbuf.String()
	} else {
		err = json.Unmarshal(outbuf.Bytes(), &res)
	}

	return res, err
}

// Return the executable name for ssh on this machine and the base args
// Base args includes port settings, user/host, everything pre the command to execute
func sshGetExeAndArgs(endpoint config.Endpoint) (exe string, baseargs []string) {
	if len(endpoint.SshUserAndHost) == 0 {
		return "", nil
	}

	isPlink := false
	isTortoise := false

	ssh := config.Config.Getenv("GIT_SSH")
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

	args := make([]string, 0, 4)
	if isTortoise {
		// TortoisePlink requires the -batch argument to behave like ssh/plink
		args = append(args, "-batch")
	}

	if len(endpoint.SshPort) > 0 {
		if isPlink || isTortoise {
			args = append(args, "-P")
		} else {
			args = append(args, "-p")
		}
		args = append(args, endpoint.SshPort)
	}
	args = append(args, endpoint.SshUserAndHost)

	return ssh, args
}
