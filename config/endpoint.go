package config

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const EndpointUrlUnknown = "<unknown>"

// An Endpoint describes how to access a Git LFS server.
type Endpoint struct {
	Url            string
	SshUserAndHost string
	SshPath        string
	SshPort        string
}

// NewEndpointFromCloneURL creates an Endpoint from a git clone URL by appending
// "[.git]/info/lfs".
func NewEndpointFromCloneURL(url string) Endpoint {
	return NewEndpointFromCloneURLWithConfig(url, New())
}

// NewEndpoint initializes a new Endpoint for a given URL.
func NewEndpoint(rawurl string) Endpoint {
	return NewEndpointWithConfig(rawurl, New())
}

// NewEndpointFromCloneURLWithConfig creates an Endpoint from a git clone URL by appending
// "[.git]/info/lfs".
func NewEndpointFromCloneURLWithConfig(url string, c *Configuration) Endpoint {
	e := NewEndpointWithConfig(url, c)
	if e.Url == EndpointUrlUnknown {
		return e
	}

	if strings.HasSuffix(url, "/") {
		e.Url = url[0 : len(url)-1]
	}

	// When using main remote URL for HTTP, append info/lfs
	if path.Ext(e.Url) == ".git" {
		e.Url += "/info/lfs"
	} else {
		e.Url += ".git/info/lfs"
	}

	return e
}

// NewEndpointWithConfig initializes a new Endpoint for a given URL.
func NewEndpointWithConfig(rawurl string, c *Configuration) Endpoint {
	rawurl = c.ReplaceUrlAlias(rawurl)
	u, err := url.Parse(rawurl)
	if err != nil {
		return endpointFromBareSshUrl(rawurl)
	}

	switch u.Scheme {
	case "ssh":
		return endpointFromSshUrl(u)
	case "http", "https":
		return endpointFromHttpUrl(u)
	case "git":
		return endpointFromGitUrl(u, c)
	case "":
		return endpointFromBareSshUrl(u.String())
	default:
		// Just passthrough to preserve
		return Endpoint{Url: rawurl}
	}
}

// endpointFromBareSshUrl constructs a new endpoint from a bare SSH URL:
//
//   user@host.com:path/to/repo.git
//
func endpointFromBareSshUrl(rawurl string) Endpoint {
	parts := strings.Split(rawurl, ":")
	partsLen := len(parts)
	if partsLen < 2 {
		return Endpoint{Url: rawurl}
	}

	// Treat presence of ':' as a bare URL
	var newPath string
	if len(parts) > 2 { // port included; really should only ever be 3 parts
		newPath = fmt.Sprintf("%v:%v", parts[0], strings.Join(parts[1:], "/"))
	} else {
		newPath = strings.Join(parts, "/")
	}
	newrawurl := fmt.Sprintf("ssh://%v", newPath)
	newu, err := url.Parse(newrawurl)
	if err != nil {
		return Endpoint{Url: EndpointUrlUnknown}
	}

	return endpointFromSshUrl(newu)
}

// endpointFromSshUrl constructs a new endpoint from an ssh:// URL
func endpointFromSshUrl(u *url.URL) Endpoint {
	var endpoint Endpoint
	// Pull out port now, we need it separately for SSH
	regex := regexp.MustCompile(`^([^\:]+)(?:\:(\d+))?$`)
	match := regex.FindStringSubmatch(u.Host)
	if match == nil || len(match) < 2 {
		endpoint.Url = EndpointUrlUnknown
		return endpoint
	}

	host := match[1]
	if u.User != nil && u.User.Username() != "" {
		endpoint.SshUserAndHost = fmt.Sprintf("%s@%s", u.User.Username(), host)
	} else {
		endpoint.SshUserAndHost = host
	}

	if len(match) > 2 {
		endpoint.SshPort = match[2]
	}

	// u.Path includes a preceding '/', strip off manually
	// rooted paths in the URL will be '//path/to/blah'
	// this is just how Go's URL parsing works
	if strings.HasPrefix(u.Path, "/") {
		endpoint.SshPath = u.Path[1:]
	} else {
		endpoint.SshPath = u.Path
	}

	// Fallback URL for using HTTPS while still using SSH for git
	// u.Host includes host & port so can't use SSH port
	endpoint.Url = fmt.Sprintf("https://%s%s", host, u.Path)

	return endpoint
}

// Construct a new endpoint from a HTTP URL
func endpointFromHttpUrl(u *url.URL) Endpoint {
	// just pass this straight through
	return Endpoint{Url: u.String()}
}

func endpointFromGitUrl(u *url.URL, c *Configuration) Endpoint {
	u.Scheme = c.GitProtocol()
	return Endpoint{Url: u.String()}
}
