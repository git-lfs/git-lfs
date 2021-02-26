package lfsapi

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/creds"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfshttp"
	"github.com/rubyist/tracerx"
)

const (
	defaultRemote = "origin"
)

type EndpointFinder interface {
	NewEndpointFromCloneURL(operation, rawurl string) lfshttp.Endpoint
	NewEndpoint(operation, rawurl string) lfshttp.Endpoint
	Endpoint(operation, remote string) lfshttp.Endpoint
	RemoteEndpoint(operation, remote string) lfshttp.Endpoint
	GitRemoteURL(remote string, forpush bool) string
	AccessFor(rawurl string) creds.Access
	SetAccess(access creds.Access)
	GitProtocol() string
}

type endpointGitFinder struct {
	gitConfig   *git.Configuration
	gitEnv      config.Environment
	gitProtocol string

	aliasMu     sync.Mutex
	aliases     map[string]string
	pushAliases map[string]string

	accessMu  sync.Mutex
	urlAccess map[string]creds.AccessMode
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
		pushAliases: make(map[string]string),
		urlAccess:   make(map[string]creds.AccessMode),
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
			return e.NewEndpoint(operation, url)
		}
	}

	if url, ok := e.gitEnv.Get("lfs.url"); ok {
		return e.NewEndpoint(operation, url)
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
			return e.NewEndpoint(operation, url)
		}
	}
	if url, ok := e.gitEnv.Get("remote." + remote + ".lfsurl"); ok {
		return e.NewEndpoint(operation, url)
	}

	// finally fall back on git remote url (also supports pushurl)
	if url := e.GitRemoteURL(remote, operation == "upload"); url != "" {
		return e.NewEndpointFromCloneURL(operation, url)
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

func (e *endpointGitFinder) NewEndpointFromCloneURL(operation, rawurl string) lfshttp.Endpoint {
	ep := e.NewEndpoint(operation, rawurl)
	if ep.Url == lfshttp.UrlUnknown {
		return ep
	}

	if strings.HasSuffix(rawurl, "/") {
		ep.Url = rawurl[0 : len(rawurl)-1]
	}

	if strings.HasPrefix(ep.Url, "file://") {
		return ep
	}

	// When using main remote URL for HTTP, append info/lfs
	if path.Ext(ep.Url) == ".git" {
		ep.Url += "/info/lfs"
	} else {
		ep.Url += ".git/info/lfs"
	}

	return ep
}

func (e *endpointGitFinder) NewEndpoint(operation, rawurl string) lfshttp.Endpoint {
	rawurl = e.ReplaceUrlAlias(operation, rawurl)
	if strings.HasPrefix(rawurl, "/") {
		return lfshttp.EndpointFromLocalPath(rawurl)
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return lfshttp.EndpointFromBareSshUrl(rawurl)
	}

	switch u.Scheme {
	case "ssh", "git+ssh", "ssh+git":
		return lfshttp.EndpointFromSshUrl(u)
	case "http", "https":
		return lfshttp.EndpointFromHttpUrl(u)
	case "git":
		return endpointFromGitUrl(u, e)
	case "file":
		return lfshttp.EndpointFromFileUrl(u)
	case "":
		// If it looks like a local path, it probably is.
		if _, err := os.Stat(rawurl); err == nil {
			return lfshttp.EndpointFromLocalPath(rawurl)
		}
		return lfshttp.EndpointFromBareSshUrl(u.String())
	default:
		if strings.HasPrefix(rawurl, u.Scheme+"::") {
			// Looks like a remote helper; just pass it through.
			return lfshttp.Endpoint{Url: rawurl}
		}
		// If it looks like a local path, it probably is.
		if _, err := os.Stat(rawurl); err == nil {
			return lfshttp.EndpointFromLocalPath(rawurl)
		}
		// We probably got here because the "scheme" that was parsed is
		// a hostname (whether FQDN or single word) and the URL parser
		// didn't know what to do with it.  Do what Git does and treat
		// it as an SSH URL.  This ensures we handle SSH config aliases
		// properly.
		return lfshttp.EndpointFromBareSshUrl(u.String())
	}
}

func (e *endpointGitFinder) AccessFor(rawurl string) creds.Access {
	accessurl := urlWithoutAuth(rawurl)

	if e.gitEnv == nil {
		return creds.NewAccess(creds.NoneAccess, accessurl)
	}

	e.accessMu.Lock()
	defer e.accessMu.Unlock()

	if cached, ok := e.urlAccess[accessurl]; ok {
		return creds.NewAccess(cached, accessurl)
	}

	e.urlAccess[accessurl] = e.fetchGitAccess(accessurl)
	return creds.NewAccess(e.urlAccess[accessurl], accessurl)
}

func (e *endpointGitFinder) SetAccess(access creds.Access) {
	key := fmt.Sprintf("lfs.%s.access", access.URL())
	tracerx.Printf("setting repository access to %s", access.Mode())

	e.accessMu.Lock()
	defer e.accessMu.Unlock()

	switch access.Mode() {
	case creds.EmptyAccess, creds.NoneAccess:
		e.gitConfig.UnsetLocalKey(key)
		e.urlAccess[access.URL()] = creds.NoneAccess
	default:
		e.gitConfig.SetLocal(key, string(access.Mode()))
		e.urlAccess[access.URL()] = access.Mode()
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

func (e *endpointGitFinder) fetchGitAccess(rawurl string) creds.AccessMode {
	if v, _ := e.urlConfig.Get("lfs", rawurl, "access"); len(v) > 0 {
		access := creds.AccessMode(strings.ToLower(v))
		if access == creds.PrivateAccess {
			return creds.BasicAccess
		}
		return access
	}
	return creds.NoneAccess
}

func (e *endpointGitFinder) GitProtocol() string {
	return e.gitProtocol
}

// ReplaceUrlAlias returns a url with a prefix from a `url.*.insteadof` git
// config setting. If multiple aliases match, use the longest one.
// See https://git-scm.com/docs/git-config for Git's docs.
func (e *endpointGitFinder) ReplaceUrlAlias(operation, rawurl string) string {
	e.aliasMu.Lock()
	defer e.aliasMu.Unlock()

	if operation == "upload" {
		if rawurl, replaced := e.replaceUrlAlias(e.pushAliases, rawurl); replaced {
			return rawurl
		}
	}
	rawurl, _ = e.replaceUrlAlias(e.aliases, rawurl)

	return rawurl
}

// replaceUrlAlias is a helper function for ReplaceUrlAlias.  It must only be
// called while the e.aliasMu mutex is held.
func (e *endpointGitFinder) replaceUrlAlias(aliases map[string]string, rawurl string) (string, bool) {
	var longestalias string
	for alias, _ := range aliases {
		if !strings.HasPrefix(rawurl, alias) {
			continue
		}

		if longestalias < alias {
			longestalias = alias
		}
	}

	if len(longestalias) > 0 {
		return aliases[longestalias] + rawurl[len(longestalias):], true
	}

	return rawurl, false
}

const (
	aliasPrefix = "url."
)

func initAliases(e *endpointGitFinder, git config.Environment) {
	suffix := ".insteadof"
	pushSuffix := ".pushinsteadof"
	for gitkey, gitval := range git.All() {
		if len(gitval) == 0 || !strings.HasPrefix(gitkey, aliasPrefix) {
			continue
		}
		if strings.HasSuffix(gitkey, suffix) {
			storeAlias(e.aliases, gitkey, gitval, suffix)
		} else if strings.HasSuffix(gitkey, pushSuffix) {
			storeAlias(e.pushAliases, gitkey, gitval, pushSuffix)
		}
	}
}

func storeAlias(aliases map[string]string, key string, values []string, suffix string) {
	for _, value := range values {
		url := key[len(aliasPrefix) : len(key)-len(suffix)]
		if v, ok := aliases[value]; ok && v != url {
			fmt.Fprintf(os.Stderr, "WARNING: Multiple 'url.*.%s' keys with the same alias: %q\n", suffix, value)
		}
		aliases[value] = url
	}
}

func endpointFromGitUrl(u *url.URL, e *endpointGitFinder) lfshttp.Endpoint {
	u.Scheme = e.gitProtocol
	return lfshttp.Endpoint{Url: u.String()}
}
