package lfshttp

import (
	goerrors "errors"
	"fmt"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/git-lfs/git-lfs/v3/errors"
	sshp "github.com/git-lfs/git-lfs/v3/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHCacheResolveFromCache(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh).(*sshCache)
	cache.endpoints.Store("userandhost//1//path//post", &sshAuthResponse{
		Href:      "cache",
		createdAt: time.Now(),
	})
	ssh.responses["userandhost"] = sshAuthResponse{Href: "real"}

	e := Endpoint{
		SSHMetadata: sshp.SSHMetadata{
			UserAndHost: "userandhost",
			Port:        "1",
			Path:        "path",
		},
	}

	res, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "cache", res.Href)
}

func TestSSHCacheResolveFromCacheWithFutureExpiresAt(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh).(*sshCache)
	cache.endpoints.Store("userandhost//1//path//post", &sshAuthResponse{
		Href:      "cache",
		ExpiresAt: time.Now().Add(time.Duration(1) * time.Hour),
		createdAt: time.Now(),
	})
	ssh.responses["userandhost"] = sshAuthResponse{Href: "real"}

	e := Endpoint{
		SSHMetadata: sshp.SSHMetadata{
			UserAndHost: "userandhost",
			Port:        "1",
			Path:        "path",
		},
	}

	res, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "cache", res.Href)
}

func TestSSHCacheResolveFromCacheWithFutureExpiresIn(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh).(*sshCache)
	cache.endpoints.Store("userandhost//1//path//post", &sshAuthResponse{
		Href:      "cache",
		ExpiresIn: 60 * 60,
		createdAt: time.Now(),
	})
	ssh.responses["userandhost"] = sshAuthResponse{Href: "real"}

	e := Endpoint{
		SSHMetadata: sshp.SSHMetadata{
			UserAndHost: "userandhost",
			Port:        "1",
			Path:        "path",
		},
	}

	res, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "cache", res.Href)
}

func TestSSHCacheResolveFromCacheWithPastExpiresAt(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh).(*sshCache)
	cache.endpoints.Store("userandhost//1//path//post", &sshAuthResponse{
		Href:      "cache",
		ExpiresAt: time.Now().Add(time.Duration(-1) * time.Hour),
		createdAt: time.Now(),
	})
	ssh.responses["userandhost"] = sshAuthResponse{Href: "real"}

	e := Endpoint{
		SSHMetadata: sshp.SSHMetadata{
			UserAndHost: "userandhost",
			Port:        "1",
			Path:        "path",
		},
	}

	res, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "real", res.Href)
}

func TestSSHCacheResolveFromCacheWithPastExpiresIn(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh).(*sshCache)
	cache.endpoints.Store("userandhost//1//path//post", &sshAuthResponse{
		Href:      "cache",
		ExpiresIn: -60 * 60,
		createdAt: time.Now(),
	})
	ssh.responses["userandhost"] = sshAuthResponse{Href: "real"}

	e := Endpoint{
		SSHMetadata: sshp.SSHMetadata{
			UserAndHost: "userandhost",
			Port:        "1",
			Path:        "path",
		},
	}

	res, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "real", res.Href)
}

func TestSSHCacheResolveFromCacheWithAmbiguousExpirationInfo(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh).(*sshCache)
	cache.endpoints.Store("userandhost//1//path//post", &sshAuthResponse{
		Href:      "cache",
		ExpiresIn: 60 * 60,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		createdAt: time.Now(),
	})
	ssh.responses["userandhost"] = sshAuthResponse{Href: "real"}

	e := Endpoint{
		SSHMetadata: sshp.SSHMetadata{
			UserAndHost: "userandhost",
			Port:        "1",
			Path:        "path",
		},
	}

	res, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "cache", res.Href)
}

func TestSSHCacheResolveWithoutError(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh).(*sshCache)

	assertCacheLen(t, cache, 0)

	ssh.responses["userandhost"] = sshAuthResponse{Href: "real"}

	e := Endpoint{
		SSHMetadata: sshp.SSHMetadata{
			UserAndHost: "userandhost",
			Port:        "1",
			Path:        "path",
		},
	}

	res, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "real", res.Href)

	assertCacheLen(t, cache, 1)
	val, ok := cache.endpoints.Load("userandhost//1//path//post")
	assert.True(t, ok)
	assert.NotNil(t, val)
	assert.Equal(t, "real", val.(*sshAuthResponse).Href)

	delete(ssh.responses, "userandhost")
	res2, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "real", res2.Href)
}

func TestSSHCacheResolveWithError(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh).(*sshCache)

	assertCacheLen(t, cache, 0)

	ssh.responses["userandhost"] = sshAuthResponse{Message: "resolve error", Href: "real"}

	e := Endpoint{
		SSHMetadata: sshp.SSHMetadata{
			UserAndHost: "userandhost",
			Port:        "1",
			Path:        "path",
		},
	}

	res, err := cache.Resolve(e, "post")
	assert.NotNil(t, err)
	assert.Equal(t, "real", res.Href)

	assertCacheLen(t, cache, 0)
	delete(ssh.responses, "userandhost")
	res2, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "", res2.Href)
}

// countingResolver always returns the configured error and records how many
// times Resolve was called.
type countingResolver struct {
	err   error
	calls int
}

func (r *countingResolver) Resolve(e Endpoint, method string) (sshAuthResponse, error) {
	r.calls++
	return sshAuthResponse{}, r.err
}

func newTestClient(t *testing.T, resolver SSHResolver) *Client {
	t.Helper()
	c, err := NewClient(NewContext(nil, nil, nil))
	assert.Nil(t, err)
	c.SSH = resolver
	c.sshTries = 5
	return c
}

func testSSHEndpoint() Endpoint {
	return Endpoint{
		Url:         "https://git-server.com/foo/bar.git/info/lfs",
		OriginalUrl: "git@git-server.com:foo/bar.git",
		SSHMetadata: sshp.SSHMetadata{
			UserAndHost: "git@git-server.com",
			Path:        "foo/bar.git",
		},
	}
}

func TestSSHResolveFallsBackWhenAuthenticateUnavailable(t *testing.T) {
	resolver := &countingResolver{
		err: &sshAuthenticateUnavailableError{
			err: errors.New("bash: git-lfs-authenticate: command not found"),
		},
	}
	c := newTestClient(t, resolver)

	res, err := c.sshResolveWithRetries(testSSHEndpoint(), "GET")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "", res.Href)

	// The unavailable command should not be retried.
	assert.Equal(t, 1, resolver.calls)
}

func TestSSHResolveUsesGuessedEndpointWhenAuthenticateUnavailable(t *testing.T) {
	resolver := &countingResolver{
		err: &sshAuthenticateUnavailableError{
			err: errors.New("bash: git-lfs-authenticate: command not found"),
		},
	}
	c := newTestClient(t, resolver)

	req, err := c.NewRequest("GET", testSSHEndpoint(), "objects/batch", nil)
	assert.Nil(t, err)
	assert.Equal(t,
		"https://git-server.com/foo/bar.git/info/lfs/objects/batch",
		req.URL.String())
}

func TestSSHResolveFailsForOtherErrors(t *testing.T) {
	resolver := &countingResolver{err: errors.New("permission denied")}
	c := newTestClient(t, resolver)

	_, err := c.sshResolveWithRetries(testSSHEndpoint(), "GET")
	assert.NotNil(t, err)

	// Non-fallback errors are retried sshTries+1 times.
	assert.Equal(t, 6, resolver.calls)
}

// exitErrorWithCode runs a trivial command that exits with the given status so
// that tests can obtain a real *exec.ExitError.
func exitErrorWithCode(t *testing.T, code int) error {
	t.Helper()
	err := exec.Command("sh", "-c", fmt.Sprintf("exit %d", code)).Run()
	var exitErr *exec.ExitError
	require.True(t, goerrors.As(err, &exitErr))
	require.Equal(t, code, exitErr.ExitCode())
	return err
}

func TestIsSSHAuthenticateUnavailable(t *testing.T) {
	for _, tc := range []struct {
		name     string
		exitCode int
		stderr   string
		want     bool
	}{
		{
			name:     "command not found exit code",
			exitCode: 127,
			stderr:   "",
			want:     true,
		},
		{
			name:     "gerrit generic exit code with message",
			exitCode: 1,
			stderr:   "fatal: Gerrit Code Review: git-lfs-authenticate: not found",
			want:     true,
		},
		{
			name:     "bash command not found message",
			exitCode: 1,
			stderr:   "bash: git-lfs-authenticate: command not found",
			want:     true,
		},
		{
			name:     "zsh command not found message",
			exitCode: 1,
			stderr:   "zsh: command not found: git-lfs-authenticate",
			want:     true,
		},
		{
			name:     "no such file message",
			exitCode: 1,
			stderr:   "git-lfs-authenticate: No such file or directory",
			want:     true,
		},
		{
			name:     "unrelated not found error is not a fallback",
			exitCode: 1,
			stderr:   "Repository not found",
			want:     false,
		},
		{
			name:     "generic error without message is fatal",
			exitCode: 1,
			stderr:   "",
			want:     false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := exitErrorWithCode(t, tc.exitCode)
			assert.Equal(t, tc.want, isSSHAuthenticateUnavailable(err, tc.stderr))
		})
	}
}

func TestIsSSHAuthenticateUnavailableNonExitError(t *testing.T) {
	// A failure to start the command (not an *exec.ExitError) is a genuine
	// error, never an "unavailable" signal, even with a matching message.
	err := errors.New("exec: \"ssh\": executable file not found in $PATH")
	assert.False(t, isSSHAuthenticateUnavailable(err,
		"git-lfs-authenticate: not found"))
}

func assertCacheLen(t *testing.T, cache *sshCache, expected int) {
	t.Helper()
	n := 0
	cache.endpoints.Range(func(_, _ any) bool { n++; return true })
	assert.Equal(t, expected, n)
}

func TestSSHCacheConcurrentResolve(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh)

	ssh.responses["userandhost"] = sshAuthResponse{Href: "real"}

	e := Endpoint{
		SSHMetadata: sshp.SSHMetadata{
			UserAndHost: "userandhost",
			Port:        "1",
			Path:        "path",
		},
	}

	// Two goroutines resolving the same endpoint concurrently is enough
	// for `go test -race` to detect an unprotected map access.
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			res, err := cache.Resolve(e, "post")
			assert.Nil(t, err)
			assert.Equal(t, "real", res.Href)
		}()
	}
	close(start)
	wg.Wait()
}

func newFakeResolver() *fakeResolver {
	return &fakeResolver{responses: make(map[string]sshAuthResponse)}
}

type fakeResolver struct {
	responses map[string]sshAuthResponse
}

func (r *fakeResolver) Resolve(e Endpoint, method string) (sshAuthResponse, error) {
	res := r.responses[e.SSHMetadata.UserAndHost]
	var err error
	if len(res.Message) > 0 {
		err = errors.New(res.Message)
	}

	res.createdAt = time.Now()

	return res, err
}
