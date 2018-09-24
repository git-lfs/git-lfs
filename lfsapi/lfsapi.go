package lfsapi

import (
	"fmt"
	"sync"

	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/git-lfs/git-lfs/creds"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfshttp"
)

type Client struct {
	Endpoints   EndpointFinder
	Credentials creds.CredentialHelper
	Netrc       NetrcFinder

	ntlmSessions map[string]ntlm.ClientSession
	ntlmMu       sync.Mutex

	credContext *creds.CredentialHelperContext

	client *lfshttp.Client
}

func NewClient(ctx lfshttp.Context) (*Client, error) {
	if ctx == nil {
		ctx = lfshttp.NewContext(nil, nil, nil)
	}

	gitEnv := ctx.GitEnv()
	osEnv := ctx.OSEnv()
	netrc, netrcfile, err := ParseNetrc(osEnv)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("bad netrc file %s", netrcfile))
	}

	httpClient, err := lfshttp.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error creating http client"))
	}

	c := &Client{
		Endpoints:   NewEndpointFinder(ctx),
		Netrc:       netrc,
		client:      httpClient,
		credContext: creds.NewCredentialHelperContext(gitEnv, osEnv),
	}

	return c, nil
}
