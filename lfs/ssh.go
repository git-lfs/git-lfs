package lfs

import (
	"encoding/json"

	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

type sshAuthResponse struct {
	Message   string            `json:"-"`
	Href      string            `json:"href"`
	Header    map[string]string `json:"header"`
	ExpiresAt string            `json:"expires_at"`
}

func sshAuthenticate(endpoint Endpoint, operation, oid string) (sshAuthResponse, error) {
	res := sshAuthResponse{}
	if len(endpoint.SshUserAndHost) == 0 {
		return res, nil
	}

	tracerx.Printf("ssh: %s git-lfs-authenticate %s %s %s",
		endpoint.SshUserAndHost, endpoint.SshPath, operation, oid)
	cmd := execCommand("ssh", endpoint.SshUserAndHost,
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
