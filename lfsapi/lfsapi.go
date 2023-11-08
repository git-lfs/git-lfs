package lfsapi

import (
	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/creds"
	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/git-lfs/git-lfs/v3/ssh"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
)

type Client struct {
	Endpoints   EndpointFinder
	Credentials creds.CredentialHelper

	credContext *creds.CredentialHelperContext

	client  *lfshttp.Client
	context lfshttp.Context
	access  []creds.AccessMode
}

func NewClient(ctx lfshttp.Context) (*Client, error) {
	if ctx == nil {
		ctx = lfshttp.NewContext(nil, nil, nil)
	}

	gitEnv := ctx.GitEnv()
	osEnv := ctx.OSEnv()

	httpClient, err := lfshttp.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, tr.Tr.Get("error creating HTTP client"))
	}

	c := &Client{
		Endpoints:   NewEndpointFinder(ctx),
		client:      httpClient,
		context:     ctx,
		credContext: creds.NewCredentialHelperContext(gitEnv, osEnv),
		access:      creds.AllAccessModes(),
	}

	return c, nil
}

func (c *Client) Context() lfshttp.Context {
	return c.context
}

// SSHTransfer returns either an suitable transfer object or nil if the
// server is not using an SSH remote or the git-lfs-transfer style of SSH
// remote.
func (c *Client) SSHTransfer(operation, remote string) *ssh.SSHTransfer {
	if len(operation) == 0 {
		return nil
	}
	endpoint := c.Endpoints.Endpoint(operation, remote)
	if len(endpoint.SSHMetadata.UserAndHost) == 0 {
		return nil
	}
	uc := config.NewURLConfig(c.context.GitEnv())
	if val, ok := uc.Get("lfs", endpoint.OriginalUrl, "sshtransfer"); ok && val != "negotiate" && val != "always" {
		tracerx.Printf("skipping pure SSH protocol connection by request")
		return nil
	}
	ctx := c.Context()
	tracerx.Printf("attempting pure SSH protocol connection")
	sshTransfer, err := ssh.NewSSHTransfer(ctx.OSEnv(), ctx.GitEnv(), &endpoint.SSHMetadata, operation)
	if err != nil {
		tracerx.Printf("pure SSH protocol connection failed: %s", err)
		return nil
	}
	return sshTransfer
}
