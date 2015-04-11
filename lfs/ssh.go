package lfs

import (
	"encoding/json"
	"github.com/rubyist/tracerx"
	"net/http"
	"os/exec"
)

type sshAuthResponse struct {
	Message   string            `json:"-"`
	Header    map[string]string `json:"header"`
	ExpiresAt string            `json:"expires_at"`
}

func mergeSshHeader(header http.Header, endpoint Endpoint, operation, oid string) error {
	if len(endpoint.SshUserAndHost) == 0 {
		return nil
	}

	res, err := sshAuthenticate(endpoint, operation, oid)
	if err != nil {
		return err
	}

	if res.Header != nil {
		for key, value := range res.Header {
			header.Set(key, value)
		}
	}

	return nil
}

func sshAuthenticate(endpoint Endpoint, operation, oid string) (sshAuthResponse, error) {
	tracerx.Printf("ssh: %s git-lfs-authenticate %s %s %s",
		endpoint.SshUserAndHost, endpoint.SshPath, operation, oid)
	cmd := exec.Command("ssh", endpoint.SshUserAndHost,
		"git-lfs-authenticate",
		endpoint.SshPath,
		operation, oid,
	)

	out, err := cmd.CombinedOutput()
	res := sshAuthResponse{}

	if err != nil {
		res.Message = string(out)
	} else {
		err = json.Unmarshal(out, &res)
	}

	return res, err
}
