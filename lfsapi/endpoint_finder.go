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
	"github.com/git-lfs/git-lfs/lfshttp"
	"github.com/rubyist/tracerx"
)

type AccessMode string

const (
	NoneAccess      AccessMode = "none"
	BasicAccess     AccessMode = "basic"
	PrivateAccess   AccessMode = "private"
	NegotiateAccess AccessMode = "negotiate"
	NTLMAccess      AccessMode = "ntlm"
	emptyAccess     AccessMode = ""
	defaultRemote              = "origin"
)

type EndpointFinder interface {
	NewEndpointFromCloneURL(rawurl string) lfshttp.Endpoint
	NewEndpoint(rawurl string) lfshttp.Endpoint
	Endpoint(operation, remote string) lfshttp.Endpoint
	RemoteEndpoint(operation, remote string) lfshttp.Endpoint
	GitRemoteURL(remote string, forpush bool) string
	AccessFor(rawurl string) AccessMode
	SetAccess(rawurl string, access AccessMode)
	GitProtocol() string
}

type endpointGitFinder struct {
	gitConfig   *git.Configuration
	gitEnv      config.Environment
	gitProtocol string

	aliasMu sync.Mutex
	aliases map[string]string

	accessMu  sync.Mutex
	urlAccess map[string]AccessMode
	urlConfig *config.URLConfig
}

func NewEndpointFinder(ctx lfshttp.Context) EndpointFinder {
	if ctx == nil {
		ctx = lfshttp.NewContext(nil, nil, nil)
	}

	e := &endpointGitFinder{
		gitConfig:   ctx.GitConfig(),
		gitEnv:      ctx.GitEnv(),
		gitProtocol: "https",
		aliases:     make(map[string]string),
		urlAccess:   make(map[string]AccessMode),
	}

	e.urlConfig = config.NewURLConfig(e.gitEnv)
	if v, ok := e.gitEnv.Get("lfs.gitprotocol"); ok {
		e.gitProtocol = v
	}
	initAliases(e, e.gitEnv)

	return e
}

func (e *endpointGitFinder) Endpoint(operation, remote string) lfshttp.Endpoint {
	ep := e.getEndpoint(operation, remote)
	ep.Operation = operation
	return ep
}

func (e *endpointGitFinder) getEndpoint(operation, remote string) lfshttp.Endpoint {
	if e.gitEnv == nil {
		return lfshttp.Endpoint{}
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

func (e *endpointGitFinder) RemoteEndpoint(operation, remote string) lfshttp.Endpoint {
	if e.gitEnv == nil {
		return lfshttp.Endpoint{}
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

	return lfshttp.Endpoint{}
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

func (e *endpointGitFinder) NewEndpointFromCloneURL(rawurl string) lfshttp.Endpoint {
	ep := e.NewEndpoint(rawurl)
	if ep.Url == lfshttp.UrlUnknown {
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

func (e *endpointGitFinder) NewEndpoint(rawurl string) lfshttp.Endpoint {
	rawurl = e.ReplaceUrlAlias(rawurl)
	if strings.HasPrefix(rawurl, "/") {
		return lfshttp.EndpointFromLocalPath(rawurl)
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return lfshttp.EndpointFromBareSshUrl(rawurl)
	}

	switch u.Scheme {
	case "ssh":
		return lfshttp.EndpointFromSshUrl(u)
	case "http", "https":
		return lfshttp.EndpointFromHttpUrl(u)
	case "git":
		return endpointFromGitUrl(u, e)
	case "":
		return lfshttp.EndpointFromBareSshUrl(u.String())
	default:
		if strings.HasPrefix(rawurl, u.Scheme+"::") {
			// Looks like a remote helper; just pass it through.
			return lfshttp.Endpoint{Url: rawurl}
		}
		// We probably got here because the "scheme" that was parsed is
		// a hostname (whether FQDN or single word) and the URL parser
		// didn't know what to do with it.  Do what Git does and treat
		// it as an SSH URL.  This ensures we handle SSH config aliases
		// properly.
		return lfshttp.EndpointFromBareSshUrl(u.String())
	}
}

func (e *endpointGitFinder) AccessFor(rawurl string) AccessMode {
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

func (e *endpointGitFinder) SetAccess(rawurl string, access AccessMode) {
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

func (e *endpointGitFinder) fetchGitAccess(rawurl string) AccessMode {
	if v, _ := e.urlConfig.Get("lfs", rawurl, "access"); len(v) > 0 {
		access := AccessMode(strings.ToLower(v))
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

func endpointFromGitUrl(u *url.URL, e *endpointGitFinder) lfshttp.Endpoint {
	u.Scheme = e.gitProtocol
	return lfshttp.Endpoint{Url: u.String()}
}
