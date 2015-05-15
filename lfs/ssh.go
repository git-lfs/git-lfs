package lfs

import (
	"encoding/json"
	"os/exec"

	"github.com/rubyist/tracerx"
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
	// SJS TODO - however still needs upgrading to support Windows, PuTTY & Tortoise like mainline Git
	//            respect GIT_SSH etc

	res := sshAuthResponse{}
	if len(endpoint.SshUserAndHost) == 0 {
		return res, nil
	}

	tracerx.Printf("ssh: %s git-lfs-authenticate %s %s %s",
		endpoint.SshUserAndHost, endpoint.SshPath, operation, oid)
	cmd := exec.Command("ssh", endpoint.SshUserAndHost,
		"git-lfs-authenticate",
		endpoint.SshPath,
		operation, oid,
	)

	out, err := cmd.CombinedOutput()

	if err != nil {
		res.Message = string(out)
	} else {
		err = json.Unmarshal(out, &res)
	}

	return res, err
}
