package lfsapi

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/git-lfs/git-lfs/errors"
)

var (
	lfsMediaTypeRE  = regexp.MustCompile(`\Aapplication/vnd\.git\-lfs\+json(;|\z)`)
	jsonMediaTypeRE = regexp.MustCompile(`\Aapplication/json(;|\z)`)
)

type Client struct {
	Endpoints   EndpointFinder
	Credentials CredentialHelper
	Netrc       NetrcFinder
}

func NewClient(osEnv env, gitEnv env) (*Client, error) {
	netrc, err := ParseNetrc(osEnv)
	if err != nil {
		return nil, err
	}

	return &Client{
		Endpoints: NewEndpointFinder(gitEnv),
		Credentials: &CommandCredentialHelper{
			SkipPrompt: !osEnv.Bool("GIT_TERMINAL_PROMPT", true),
		},
		Netrc: netrc,
	}, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if seeker, ok := req.Body.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, err
	}

	return res, c.handleResponse(res)
}

func decodeResponse(res *http.Response, obj interface{}) error {
	ctype := res.Header.Get("Content-Type")
	if !(lfsMediaTypeRE.MatchString(ctype) || jsonMediaTypeRE.MatchString(ctype)) {
		return nil
	}

	err := json.NewDecoder(res.Body).Decode(obj)
	res.Body.Close()

	if err != nil {
		return errors.Wrapf(err, "Unable to parse HTTP response for %s %s", res.Request.Method, res.Request.URL)
	}

	return nil
}
