package lfsapi

import (
	"fmt"
	"sync"

	"github.com/git-lfs/git-lfs/v3/config"
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
	access  []creds.AccessMode

	sshTransfers      map[string]*ssh.SSHTransfer
	sshTransfersMutex *sync.Mutex
}

func NewClient(ctx lfshttp.Context) *Client {
	if ctx == nil {
		ctx = lfshttp.NewContext(nil, nil, nil)
	}

	gitEnv := ctx.GitEnv()
	osEnv := ctx.OSEnv()

	return &Client{
		Endpoints:   NewEndpointFinder(ctx),
		client:      lfshttp.NewClient(ctx),
		context:     ctx,
		credContext: creds.NewCredentialHelperContext(gitEnv, osEnv),
		access:      creds.AllAccessModes(),

		sshTransfers:      make(map[string]*ssh.SSHTransfer),
		sshTransfersMutex: &sync.Mutex{},
	}
}

func (c *Client) Context() lfshttp.Context {
	return c.context
}

// SSHTransfer returns either an suitable transfer object or nil.  For a
// given operation and remote, the same transfer object (or nil) will
// always be returned.
func (c *Client) SSHTransfer(operation, remote string) *ssh.SSHTransfer {
	if len(operation) == 0 {
		return nil
	}

	k := fmt.Sprintf("%s.%s", operation, remote)

	c.sshTransfersMutex.Lock()
	defer c.sshTransfersMutex.Unlock()

	sshTransfer, ok := c.sshTransfers[k]
	if !ok {
		var err error
		sshTransfer, err = c.initSSHTransfer(operation, remote)
		if err == nil {
			c.sshTransfers[k] = sshTransfer
		}
	}

	return sshTransfer
}

// initSSHTransfer returns either an suitable transfer object or nil if the
// server is not using an SSH remote or the git-lfs-transfer style of SSH
// remote.
func (c *Client) initSSHTransfer(operation, remote string) (*ssh.SSHTransfer, error) {
	endpoint := c.Endpoints.Endpoint(operation, remote)
	if len(endpoint.SSHMetadata.UserAndHost) == 0 {
		return nil, nil
	}

	uc := config.NewURLConfig(c.context.GitEnv())
	if val, ok := uc.Get("lfs", endpoint.OriginalUrl, "sshtransfer"); ok && val != "negotiate" && val != "always" {
		tracerx.Printf("skipping pure SSH protocol connection by request (%s, %s)", operation, remote)
		return nil, nil
	}

	ctx := c.Context()
	tracerx.Printf("attempting pure SSH protocol connection (%s, %s)", operation, remote)
	sshTransfer, err := ssh.NewSSHTransfer(ctx.OSEnv(), ctx.GitEnv(), &endpoint.SSHMetadata, operation)
	if err != nil {
		tracerx.Printf("pure SSH protocol connection failed (%s, %s): %s", operation, remote, err)
		return nil, err
	}

	return sshTransfer, nil
}

func (c *Client) closeSSHTransfers() error {
	c.sshTransfersMutex.Lock()
	defer c.sshTransfersMutex.Unlock()

	var multiErr error
	for _, sshTransfer := range c.sshTransfers {
		if sshTransfer != nil {
			multiErr = errors.Join(multiErr, sshTransfer.Shutdown())
		}
	}

	clear(c.sshTransfers)

	return multiErr
}
