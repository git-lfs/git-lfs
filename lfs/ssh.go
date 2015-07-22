package lfs

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

type sshAuthResponse struct {
	Message   string            `json:"-"`
	Href      string            `json:"href"`
	Header    map[string]string `json:"header"`
	ExpiresAt string            `json:"expires_at"`
}

func sshAuthenticate(endpoint Endpoint, operation, oid string) (sshAuthResponse, error) {

	// This is only used as a fallback where the Git URL is SSH but server doesn't support a full SSH binary protocol
	// and therefore we derive a HTTPS endpoint for binaries instead; but check authentication here via SSH

	res := sshAuthResponse{}
	if len(endpoint.SshUserAndHost) == 0 {
		return res, nil
	}

	tracerx.Printf("ssh: %s git-lfs-authenticate %s %s %s",
		endpoint.SshUserAndHost, endpoint.SshPath, operation, oid)

	exe, args := sshGetExeAndArgs(endpoint)
	args = append(args,
		"git-lfs-authenticate",
		endpoint.SshPath,
		operation, oid)

	cmd := exec.Command(exe, args...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		res.Message = string(out)
	} else {
		err = json.Unmarshal(out, &res)
	}

	return res, err
}

// Return the executable name for ssh on this machine and the base args
// Base args includes port settings, user/host, everything pre the command to execute
func sshGetExeAndArgs(endpoint Endpoint) (exe string, baseargs []string) {
	if len(endpoint.SshUserAndHost) == 0 {
		return "", nil
	}

	isPlink := false
	isTortoise := false

	ssh := Config.Getenv("GIT_SSH")
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
