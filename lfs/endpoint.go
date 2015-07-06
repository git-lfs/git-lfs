package lfs

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
	e := NewEndpoint(url)
	if e.Url == EndpointUrlUnknown {
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
func NewEndpoint(rawurl string) Endpoint {
	var endpoint Endpoint
	// Check for SSH URLs
	// Support ssh://user@host.com/path/to/repo and user@host.com:path/to/repo
	u, err := url.Parse(rawurl)
	if err == nil {
		u = processBareSshUrl(u)
		switch u.Scheme {
		case "ssh":
			endpoint = endpointFromSshUrl(u)
		case "http", "https":
			endpoint = endpointFromHttpUrl(u)
		default:
			// Just passthrough to preserve
			endpoint = Endpoint{Url: rawurl}
		}
	} else {
		endpoint = Endpoint{Url: EndpointUrlUnknown}
	}

	return endpoint
}

// For ease of processing, sanitise a bare Git SSH URL into a ssh:// URL
func processBareSshUrl(u *url.URL) *url.URL {
	if u.Scheme != "" || u.Path == "" {
		return u
	}
	// This might be a bare SSH URL
	// Turn it into ssh:// for simplicity of extraction later
	parts := strings.Split(u.Path, ":")
	if len(parts) > 1 {
		// Treat presence of ':' as a bare URL
		var newPath string
		if len(parts) > 2 { // port included; really should only ever be 3 parts
			newPath = fmt.Sprintf("%v:%v", parts[0], strings.Join(parts[1:], "/"))
		} else {
			newPath = strings.Join(parts, "/")
		}
		newrawurl := fmt.Sprintf("ssh://%v", newPath)
		newu, err := url.Parse(newrawurl)
		if err == nil {
			return newu
		}
	}
	return u
}

// Construct a new endpoint from a SSH URL (sanitised to ssh://)
func endpointFromSshUrl(u *url.URL) Endpoint {
	if u.Scheme != "ssh" {
		return Endpoint{Url: EndpointUrlUnknown}
	}
	var host string
	var endpoint Endpoint
	// Pull out port now, we need it separately for SSH
	regex := regexp.MustCompile(`^([^\:]+)(?:\:(\d+))?$`)
	if match := regex.FindStringSubmatch(u.Host); match != nil {
		host = match[1]
		if u.User != nil && u.User.Username() != "" {
			endpoint.SshUserAndHost = fmt.Sprintf("%s@%s", u.User.Username(), host)
		} else {
			endpoint.SshUserAndHost = host
		}
		if len(match) > 2 {
			endpoint.SshPort = match[2]
		}
	} else {
		endpoint.Url = EndpointUrlUnknown
		return endpoint
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
