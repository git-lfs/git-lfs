package lfsapi

import (
	"fmt"
	"sync"

	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfshttp"
)

type Client struct {
	Endpoints   EndpointFinder
	Credentials CredentialHelper
	Netrc       NetrcFinder

	ntlmSessions map[string]ntlm.ClientSession
	ntlmMu       sync.Mutex

	commandCredHelper *commandCredentialHelper
	askpassCredHelper *AskPassCredentialHelper
	cachingCredHelper *credentialCacher

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
		Endpoints: NewEndpointFinder(ctx),
		Netrc:     netrc,
		commandCredHelper: &commandCredentialHelper{
			SkipPrompt: osEnv.Bool("GIT_TERMINAL_PROMPT", false),
		},
		client: httpClient,
	}

	askpass, ok := osEnv.Get("GIT_ASKPASS")
	if !ok {
		askpass, ok = gitEnv.Get("core.askpass")
	}
	if !ok {
		askpass, _ = osEnv.Get("SSH_ASKPASS")
	}
	if len(askpass) > 0 {
		c.askpassCredHelper = &AskPassCredentialHelper{
			Program: askpass,
		}
	}

	cacheCreds := gitEnv.Bool("lfs.cachecredentials", true)
	if cacheCreds {
		c.cachingCredHelper = newCredentialCacher()
	}

	return c, nil
}
