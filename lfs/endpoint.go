package lfs

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const ENDPOINT_URL_UNKNOWN = "<unknown>"

var httpPrefixRe = regexp.MustCompile("\\Ahttps?://")

// An Endpoint describes how to access a Git LFS server.
type Endpoint struct {
	Url            string
	UrlUser        string
	UrlPassword    string
	SshUserAndHost string
	SshPath        string
}

// NewEndpointFromCloneURL creates an Endpoint from a git clone URL by appending
// "[.git]/info/lfs".
func NewEndpointFromCloneURL(url string) Endpoint {
	e := NewEndpoint(url)
	if e.Url == ENDPOINT_URL_UNKNOWN {
		return e
	}

	// When using main remote URL for HTTP, append info/lfs
	if path.Ext(url) == ".git" {
		e.Url += "/info/lfs"
	} else {
		e.Url += ".git/info/lfs"
	}
	return e
}

// NewEndpoint initializes a new Endpoint for a given URL.
func NewEndpoint(url string) Endpoint {
	e := Endpoint{Url: url}

	if httpPrefixRe.MatchString(url) {
		processHttpUrl(&e)
	} else {
		processSshUrl(&e)
	}

	return e
}

func (e Endpoint) HasUrlAuth() bool {
	return len(e.UrlUser) > 0 || len(e.UrlPassword) > 0
}

func ObjectUrl(endpoint Endpoint, oid string) (*url.URL, error) {
	u, err := url.Parse(endpoint.Url)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "objects")
	if len(oid) > 0 {
		u.Path = path.Join(u.Path, oid)
	}
	return u, nil
}

func processHttpUrl(e *Endpoint) {
	u, err := url.Parse(e.Url)
	if err != nil {
		// drop the error, this will probably come up later in the Git LFS command.
		return
	}

	if u.User != nil {
		e.UrlUser = u.User.Username()
		pwd, _ := u.User.Password()
		e.UrlPassword = pwd
	}
}

func processSshUrl(e *Endpoint) {
	pieces := strings.SplitN(e.Url, ":", 2)
	hostPieces := strings.SplitN(pieces[0], "@", 2)
	if len(hostPieces) == 2 {
		e.SshUserAndHost = pieces[0]
		e.SshPath = pieces[1]
		e.Url = fmt.Sprintf("https://%s/%s", hostPieces[1], pieces[1])
	}
}
