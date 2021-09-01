package lfsapi

import (
	"fmt"

	"github.com/git-lfs/git-lfs/v3/creds"
	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/lfshttp"
	"github.com/git-lfs/git-lfs/v3/ssh"
	"github.com/rubyist/tracerx"
)

type Client struct {
	Endpoints   EndpointFinder
	Credentials creds.CredentialHelper

	credContext *creds.CredentialHelperContext

	client  *lfshttp.Client
	context lfshttp.Context
}

func NewClient(ctx lfshttp.Context) (*Client, error) {
	if ctx == nil {
		ctx = lfshttp.NewContext(nil, nil, nil)
	}

	gitEnv := ctx.GitEnv()
	osEnv := ctx.OSEnv()

	httpClient, err := lfshttp.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error creating http client"))
	}

	c := &Client{
		Endpoints:   NewEndpointFinder(ctx),
		client:      httpClient,
		context:     ctx,
		credContext: creds.NewCredentialHelperContext(gitEnv, osEnv),
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
	ctx := c.Context()
	tracerx.Printf("attempting pure SSH protocol connection")
	sshTransfer, err := ssh.NewSSHTransfer(ctx.OSEnv(), ctx.GitEnv(), &endpoint.SSHMetadata, operation)
	if err != nil {
		tracerx.Printf("pure SSH protocol connection failed: %s", err)
		return nil
	}
	return sshTransfer
}
