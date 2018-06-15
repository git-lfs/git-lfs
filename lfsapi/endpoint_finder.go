package lfsapi

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/rubyist/tracerx"
)

type Access string

const (
	NoneAccess      Access = "none"
	BasicAccess     Access = "basic"
	PrivateAccess   Access = "private"
	NegotiateAccess Access = "negotiate"
	NTLMAccess      Access = "ntlm"
	emptyAccess     Access = ""
	defaultRemote          = "origin"
)

type EndpointFinder interface {
	NewEndpointFromCloneURL(rawurl string) Endpoint
	NewEndpoint(rawurl string) Endpoint
	Endpoint(operation, remote string) Endpoint
	RemoteEndpoint(operation, remote string) Endpoint
	GitRemoteURL(remote string, forpush bool) string
	AccessFor(rawurl string) Access
	SetAccess(rawurl string, access Access)
	GitProtocol() string
}

type endpointGitFinder struct {
	gitConfig   *git.Configuration
	gitEnv      config.Environment
	gitProtocol string

	aliasMu sync.Mutex
	aliases map[string]string

	accessMu  sync.Mutex
	urlAccess map[string]Access
	urlConfig *config.URLConfig
}

func NewEndpointFinder(ctx Context) EndpointFinder {
	if ctx == nil {
		ctx = NewContext(nil, nil, nil)
	}

	e := &endpointGitFinder{
		gitConfig:   ctx.GitConfig(),
		gitEnv:      ctx.GitEnv(),
		gitProtocol: "https",
		aliases:     make(map[string]string),
		urlAccess:   make(map[string]Access),
	}

	e.urlConfig = config.NewURLConfig(e.gitEnv)
	if v, ok := e.gitEnv.Get("lfs.gitprotocol"); ok {
		e.gitProtocol = v
	}
	initAliases(e, e.gitEnv)

	return e
}

func (e *endpointGitFinder) Endpoint(operation, remote string) Endpoint {
	ep := e.getEndpoint(operation, remote)
	ep.Operation = operation
	return ep
}

func (e *endpointGitFinder) getEndpoint(operation, remote string) Endpoint {
	if e.gitEnv == nil {
		return Endpoint{}
	}

	if operation == "upload" {
		if url, ok := e.gitEnv.Get("lfs.pushurl"); ok {
			return e.NewEndpoint(url)
		}
	}

	if url, ok := e.gitEnv.Get("lfs.url"); ok {
		return e.NewEndpoint(url)
	}

	if len(remote) > 0 && remote != defaultRemote {
		if e := e.RemoteEndpoint(operation, remote); len(e.Url) > 0 {
			return e
		}
	}

	return e.RemoteEndpoint(operation, defaultRemote)
}

func (e *endpointGitFinder) RemoteEndpoint(operation, remote string) Endpoint {
	if e.gitEnv == nil {
		return Endpoint{}
	}

	if len(remote) == 0 {
		remote = defaultRemote
	}

	// Support separate push URL if specified and pushing
	if operation == "upload" {
		if url, ok := e.gitEnv.Get("remote." + remote + ".lfspushurl"); ok {
			return e.NewEndpoint(url)
		}
	}
	if url, ok := e.gitEnv.Get("remote." + remote + ".lfsurl"); ok {
		return e.NewEndpoint(url)
	}

	// finally fall back on git remote url (also supports pushurl)
	if url := e.GitRemoteURL(remote, operation == "upload"); url != "" {
		return e.NewEndpointFromCloneURL(url)
	}

	return Endpoint{}
}

func (e *endpointGitFinder) GitRemoteURL(remote string, forpush bool) string {
	if e.gitEnv != nil {
		if forpush {
			if u, ok := e.gitEnv.Get("remote." + remote + ".pushurl"); ok {
				return u
			}
		}

		if u, ok := e.gitEnv.Get("remote." + remote + ".url"); ok {
			return u
		}
	}

	if err := git.ValidateRemote(remote); err == nil {
		return remote
	}

	return ""
}

func (e *endpointGitFinder) NewEndpointFromCloneURL(rawurl string) Endpoint {
	ep := e.NewEndpoint(rawurl)
	if ep.Url == UrlUnknown {
		return ep
	}

	if strings.HasSuffix(rawurl, "/") {
		ep.Url = rawurl[0 : len(rawurl)-1]
	}

	// When using main remote URL for HTTP, append info/lfs
	if path.Ext(ep.Url) == ".git" {
		ep.Url += "/info/lfs"
	} else {
		ep.Url += ".git/info/lfs"
	}

	return ep
}

func (e *endpointGitFinder) NewEndpoint(rawurl string) Endpoint {
	rawurl = e.ReplaceUrlAlias(rawurl)
	if strings.HasPrefix(rawurl, "/") {
		return endpointFromLocalPath(rawurl)
	}
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
		return endpointFromGitUrl(u, e)
	case "":
		return endpointFromBareSshUrl(u.String())
	default:
		// Just passthrough to preserve
		return Endpoint{Url: rawurl}
	}
}

func (e *endpointGitFinder) AccessFor(rawurl string) Access {
	if e.gitEnv == nil {
		return NoneAccess
	}

	accessurl := urlWithoutAuth(rawurl)

	e.accessMu.Lock()
	defer e.accessMu.Unlock()

	if cached, ok := e.urlAccess[accessurl]; ok {
		return cached
	}

	e.urlAccess[accessurl] = e.fetchGitAccess(accessurl)
	return e.urlAccess[accessurl]
}

func (e *endpointGitFinder) SetAccess(rawurl string, access Access) {
	accessurl := urlWithoutAuth(rawurl)
	key := fmt.Sprintf("lfs.%s.access", accessurl)
	tracerx.Printf("setting repository access to %s", access)

	e.accessMu.Lock()
	defer e.accessMu.Unlock()

	switch access {
	case emptyAccess, NoneAccess:
		e.gitConfig.UnsetLocalKey(key)
		e.urlAccess[accessurl] = NoneAccess
	default:
		e.gitConfig.SetLocal(key, string(access))
		e.urlAccess[accessurl] = access
	}
}

func urlWithoutAuth(rawurl string) string {
	if !strings.Contains(rawurl, "@") {
		return rawurl
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing URL %q: %s", rawurl, err)
		return rawurl
	}

	u.User = nil
	return u.String()
}

func (e *endpointGitFinder) fetchGitAccess(rawurl string) Access {
	if v, _ := e.urlConfig.Get("lfs", rawurl, "access"); len(v) > 0 {
		access := Access(strings.ToLower(v))
		if access == PrivateAccess {
			return BasicAccess
		}
		return access
	}
	return NoneAccess
}

func (e *endpointGitFinder) GitProtocol() string {
	return e.gitProtocol
}

// ReplaceUrlAlias returns a url with a prefix from a `url.*.insteadof` git
// config setting. If multiple aliases match, use the longest one.
// See https://git-scm.com/docs/git-config for Git's docs.
func (e *endpointGitFinder) ReplaceUrlAlias(rawurl string) string {
	e.aliasMu.Lock()
	defer e.aliasMu.Unlock()

	var longestalias string
	for alias, _ := range e.aliases {
		if !strings.HasPrefix(rawurl, alias) {
			continue
		}

		if longestalias < alias {
			longestalias = alias
		}
	}

	if len(longestalias) > 0 {
		return e.aliases[longestalias] + rawurl[len(longestalias):]
	}

	return rawurl
}

func initAliases(e *endpointGitFinder, git config.Environment) {
	prefix := "url."
	suffix := ".insteadof"
	for gitkey, gitval := range git.All() {
		if len(gitval) == 0 || !(strings.HasPrefix(gitkey, prefix) && strings.HasSuffix(gitkey, suffix)) {
			continue
		}
		if _, ok := e.aliases[gitval[len(gitval)-1]]; ok {
			fmt.Fprintf(os.Stderr, "WARNING: Multiple 'url.*.insteadof' keys with the same alias: %q\n", gitval)
		}
		e.aliases[gitval[len(gitval)-1]] = gitkey[len(prefix) : len(gitkey)-len(suffix)]
	}
}
