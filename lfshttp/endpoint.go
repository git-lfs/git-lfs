package lfshttp

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const UrlUnknown = "<unknown>"

// An Endpoint describes how to access a Git LFS server.
type Endpoint struct {
	Url            string
	SshUserAndHost string
	SshPath        string
	SshPort        string
	Operation      string
}

func endpointOperation(e Endpoint, method string) string {
	if len(e.Operation) > 0 {
		return e.Operation
	}

	switch method {
	case "GET", "HEAD":
		return "download"
	default:
		return "upload"
	}
}

// EndpointFromSshUrl constructs a new endpoint from an ssh:// URL
func EndpointFromSshUrl(u *url.URL) Endpoint {
	var endpoint Endpoint
	// Pull out port now, we need it separately for SSH
	regex := regexp.MustCompile(`^([^\:]+)(?:\:(\d+))?$`)
	match := regex.FindStringSubmatch(u.Host)
	if match == nil || len(match) < 2 {
		endpoint.Url = UrlUnknown
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

// EndpointFromBareSshUrl constructs a new endpoint from a bare SSH URL:
//
//   user@host.com:path/to/repo.git or
//   [user@host.com:port]:path/to/repo.git
//
func EndpointFromBareSshUrl(rawurl string) Endpoint {
	parts := strings.Split(rawurl, ":")
	partsLen := len(parts)
	if partsLen < 2 {
		return Endpoint{Url: rawurl}
	}

	// Treat presence of ':' as a bare URL
	var newPath string
	if len(parts) > 2 { // port included; really should only ever be 3 parts
		// Correctly handle [host:port]:path URLs
		parts[0] = strings.TrimPrefix(parts[0], "[")
		parts[1] = strings.TrimSuffix(parts[1], "]")
		newPath = fmt.Sprintf("%v:%v", parts[0], strings.Join(parts[1:], "/"))
	} else {
		newPath = strings.Join(parts, "/")
	}
	newrawurl := fmt.Sprintf("ssh://%v", newPath)
	newu, err := url.Parse(newrawurl)
	if err != nil {
		return Endpoint{Url: UrlUnknown}
	}

	return EndpointFromSshUrl(newu)
}

// Construct a new endpoint from a HTTP URL
func EndpointFromHttpUrl(u *url.URL) Endpoint {
	// just pass this straight through
	return Endpoint{Url: u.String()}
}

func EndpointFromLocalPath(path string) Endpoint {
	if !strings.HasSuffix(path, ".git") {
		path = fmt.Sprintf("%s/.git", path)
	}
	return Endpoint{Url: fmt.Sprintf("file://%s", path)}
}

// Construct a new endpoint from a file URL
func EndpointFromFileUrl(u *url.URL) Endpoint {
	// just pass this straight through
	return Endpoint{Url: u.String()}
}
