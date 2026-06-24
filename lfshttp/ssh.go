package lfshttp

import (
	"bytes"
	"encoding/json"
	goerrors "errors"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/ssh"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/rubyist/tracerx"
)

// sshAuthenticateNotFoundExitCode is the conventional POSIX shell exit code
// returned when a command cannot be found.  When the remote `ssh` invocation
// exits with this status, we treat the `git-lfs-authenticate` command as
// unavailable on the server and fall back to the guessed LFS endpoint rather
// than treating the failure as fatal.
const sshAuthenticateNotFoundExitCode = 127

// sshAuthenticateNotFoundMessages are case-insensitive substrings of the
// stderr output produced when `git-lfs-authenticate` is not available on the
// server.  Not every server uses the conventional 127 exit code: Gerrit, for
// example, runs the command through a restricted shell that returns a generic
// exit code 1 with a "not found"-style message.  The substrings are anchored
// to the command name so that unrelated "not found" errors (such as a missing
// repository) are not misclassified.
var sshAuthenticateNotFoundMessages = []string{
	"git-lfs-authenticate: not found",         // Gerrit
	"git-lfs-authenticate: command not found", // bash, sh
	"git-lfs-authenticate: no such file",      // direct exec by path
	"command not found: git-lfs-authenticate", // zsh
}

// isSSHAuthenticateUnavailable reports whether the error and captured stderr
// from running `git-lfs-authenticate` over SSH indicate that the command is
// not available on the server, in which case Git LFS should fall back to the
// guessed LFS endpoint instead of failing the request.
func isSSHAuthenticateUnavailable(err error, stderr string) bool {
	var exitErr *exec.ExitError
	if !goerrors.As(err, &exitErr) {
		// The command never ran to completion (e.g. the local ssh
		// binary is missing); this is a genuine error, not an
		// unavailable git-lfs-authenticate command.
		return false
	}

	if exitErr.ExitCode() == sshAuthenticateNotFoundExitCode {
		return true
	}

	msg := strings.ToLower(stderr)
	for _, needle := range sshAuthenticateNotFoundMessages {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

// sshAuthenticateUnavailableError indicates that the `git-lfs-authenticate`
// command is not available on the SSH server.  Callers should fall back to the
// guessed LFS endpoint per the server discovery specification instead of
// failing the request.
type sshAuthenticateUnavailableError struct {
	err error
}

func (e *sshAuthenticateUnavailableError) Error() string {
	return e.err.Error()
}

func (e *sshAuthenticateUnavailableError) Unwrap() error {
	return e.err
}

type SSHResolver interface {
	Resolve(Endpoint, string) (sshAuthResponse, error)
}

func withSSHCache(ssh SSHResolver) SSHResolver {
	return &sshCache{ssh: ssh}
}

type sshCache struct {
	endpoints sync.Map // map[string]*sshAuthResponse
	ssh       SSHResolver
}

func (c *sshCache) Resolve(e Endpoint, method string) (sshAuthResponse, error) {
	if len(e.SSHMetadata.UserAndHost) == 0 {
		return sshAuthResponse{}, nil
	}

	key := strings.Join([]string{e.SSHMetadata.UserAndHost, e.SSHMetadata.Port, e.SSHMetadata.Path, method}, "//")
	if val, ok := c.endpoints.Load(key); ok {
		res := val.(*sshAuthResponse)
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
		c.endpoints.Store(key, &res)
	}
	return res, err
}

type sshAuthResponse struct {
	Message   string            `json:"-"`
	Href      string            `json:"href"`
	Header    map[string]string `json:"header"`
	ExpiresAt time.Time         `json:"expires_at"`
	ExpiresIn time.Duration     `json:"expires_in"`

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

		// If the git-lfs-authenticate command is not available on the
		// server, signal this distinctly so the caller can fall back to
		// the guessed LFS endpoint instead of failing outright.  This
		// covers both the conventional command-not-found exit code (127)
		// and servers such as Gerrit that return a generic exit code
		// with a "not found"-style message.
		if isSSHAuthenticateUnavailable(err, res.Message) {
			err = &sshAuthenticateUnavailableError{err: err}
		}
	} else {
		err = json.Unmarshal(outbuf.Bytes(), &res)
		if res.ExpiresIn == 0 && res.ExpiresAt.IsZero() {
			ttl := c.git.Int64("lfs.defaulttokenttl", 0)
			if ttl < 0 {
				ttl = 0
			}
			res.ExpiresIn = time.Duration(ttl)
		}
		res.createdAt = now
	}

	return res, err
}
