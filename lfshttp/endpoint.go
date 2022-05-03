package lfshttp

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/ssh"
)

const UrlUnknown = "<unknown>"

// An Endpoint describes how to access a Git LFS server.
type Endpoint struct {
	Url         string
	SSHMetadata ssh.SSHMetadata
	Operation   string
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

var sshURIHostPortRE = regexp.MustCompile(`^([^\:]+)(?:\:(\d+))?$`)

// EndpointFromSshUrl constructs a new endpoint from an ssh:// URL
func EndpointFromSshUrl(u *url.URL) Endpoint {
	var endpoint Endpoint
	// Pull out port now, we need it separately for SSH
	match := sshURIHostPortRE.FindStringSubmatch(u.Host)
	if match == nil || len(match) < 2 {
		endpoint.Url = UrlUnknown
		return endpoint
	}

	host := match[1]
	if u.User != nil && u.User.Username() != "" {
		endpoint.SSHMetadata.UserAndHost = fmt.Sprintf("%s@%s", u.User.Username(), host)
	} else {
		endpoint.SSHMetadata.UserAndHost = host
	}

	if len(match) > 2 {
		endpoint.SSHMetadata.Port = match[2]
	}

	endpoint.SSHMetadata.Path = u.Path
	endpoint.SSHMetadata.Scheme = u.Scheme

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
	parts := strings.SplitN(rawurl, ":", 3)
	partsLen := len(parts)
	if partsLen < 2 {
		return Endpoint{Url: rawurl}
	}

	// Treat presence of ':' as a bare URL
	var userHostAndPort string
	var path string
	if len(parts) > 2 { // port included; really should only ever be 3 parts
		// Correctly handle [host:port]:path URLs
//// DEBUG: user@[host:port]:... also OK
//// if [ found ] must be found too (and last!)
//// DEBUG: need tests for these
		parts[0] = strings.TrimPrefix(parts[0], "[")
		parts[1] = strings.TrimSuffix(parts[1], "]")
		userHostAndPort = fmt.Sprintf("%v:%v", parts[0], parts[1])
		path = parts[2]
	} else {
		userHostAndPort = parts[0]
		path = parts[1]
	}

	var absPath bool
	if absPath = strings.HasPrefix(path, "/"); absPath {
		path = strings.TrimLeft(path, "/")
	}

	newrawurl := fmt.Sprintf("ssh://%v/%v", userHostAndPort, path)
	newu, err := url.Parse(newrawurl)
	if err != nil {
		return Endpoint{Url: UrlUnknown}
	}

	endpoint := EndpointFromSshUrl(newu)
	if !absPath {
		endpoint.SSHMetadata.Path = strings.TrimLeft(endpoint.SSHMetadata.Path, "/")
	}
//// DEBUG: skip Scheme and just canonicalize as bare SSH in env
	endpoint.SSHMetadata.Scheme = ""
	return endpoint
}

// Construct a new endpoint from a HTTP URL
func EndpointFromHttpUrl(u *url.URL) Endpoint {
	// just pass this straight through
	return Endpoint{Url: u.String()}
}

func EndpointFromLocalPath(path string) Endpoint {
	return Endpoint{Url: git.RewriteLocalPathAsURL(path)}
}

// Construct a new endpoint from a file URL
func EndpointFromFileUrl(u *url.URL) Endpoint {
	// just pass this straight through
	return Endpoint{Url: u.String()}
}
