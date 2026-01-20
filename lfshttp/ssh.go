package lfshttp

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/ssh"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tools"
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
	if len(e.SSHMetadata.UserAndHost) == 0 {
		return sshAuthResponse{}, nil
	}

	key := strings.Join([]string{e.SSHMetadata.UserAndHost, e.SSHMetadata.Port, e.SSHMetadata.Path, method}, "//")
	if res, ok := c.endpoints[key]; ok {
		if _, expired := res.IsExpiredWithin(5 * time.Second); !expired {
			tracerx.Printf("ssh cache: %s git-lfs-authenticate %s %s",
				e.SSHMetadata.UserAndHost, e.SSHMetadata.Path, endpointOperation(e, method))
			return *res, nil
		} else {
			tracerx.Printf("ssh cache expired: %s git-lfs-authenticate %s %s",
				e.SSHMetadata.UserAndHost, e.SSHMetadata.Path, endpointOperation(e, method))
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
	return tools.IsExpiredAtOrIn(r.createdAt, d, r.ExpiresAt,
		time.Duration(r.ExpiresIn)*time.Second)
}

type sshAuthClient struct {
	os  config.Environment
	git config.Environment
}

func (c *sshAuthClient) Resolve(e Endpoint, method string) (sshAuthResponse, error) {
	res := sshAuthResponse{}
	if len(e.SSHMetadata.UserAndHost) == 0 {
		return res, nil
	}

	exe, args, _, _ := ssh.GetLFSExeAndArgs(c.os, c.git, &e.SSHMetadata, "git-lfs-authenticate", endpointOperation(e, method), false, "")
	cmd, err := subprocess.ExecCommand(exe, args...)
	if err != nil {
		return res, err
	}

	// Save stdout and stderr in separate buffers
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	now := time.Now()

	// Execute command
	err = cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	// Processing result
	if err != nil {
		res.Message = strings.TrimSpace(errbuf.String())
	} else {
		err = json.Unmarshal(outbuf.Bytes(), &res)
		if res.ExpiresIn == 0 && res.ExpiresAt.IsZero() {
			ttl := c.git.Int("lfs.defaulttokenttl", 0)
			if ttl < 0 {
				ttl = 0
			}
			res.ExpiresIn = ttl
		}
		res.createdAt = now
	}

	return res, err
}
