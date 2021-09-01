package lfshttp

import (
	"testing"
	"time"

	"github.com/git-lfs/git-lfs/v3/errors"
	sshp "github.com/git-lfs/git-lfs/v3/ssh"
	"github.com/stretchr/testify/assert"
)

func TestSSHCacheResolveFromCache(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh).(*sshCache)
	cache.endpoints["userandhost//1//path//post"] = &sshAuthResponse{
		Href:      "cache",
		createdAt: time.Now(),
	}
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
	cache.endpoints["userandhost//1//path//post"] = &sshAuthResponse{
		Href:      "cache",
		ExpiresAt: time.Now().Add(time.Duration(1) * time.Hour),
		createdAt: time.Now(),
	}
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
	cache.endpoints["userandhost//1//path//post"] = &sshAuthResponse{
		Href:      "cache",
		ExpiresIn: 60 * 60,
		createdAt: time.Now(),
	}
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
	cache.endpoints["userandhost//1//path//post"] = &sshAuthResponse{
		Href:      "cache",
		ExpiresAt: time.Now().Add(time.Duration(-1) * time.Hour),
		createdAt: time.Now(),
	}
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
	cache.endpoints["userandhost//1//path//post"] = &sshAuthResponse{
		Href:      "cache",
		ExpiresIn: -60 * 60,
		createdAt: time.Now(),
	}
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
	cache.endpoints["userandhost//1//path//post"] = &sshAuthResponse{
		Href:      "cache",
		ExpiresIn: 60 * 60,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		createdAt: time.Now(),
	}
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

	assert.Equal(t, 0, len(cache.endpoints))

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

	assert.Equal(t, 1, len(cache.endpoints))
	cacheres, ok := cache.endpoints["userandhost//1//path//post"]
	assert.True(t, ok)
	assert.NotNil(t, cacheres)
	assert.Equal(t, "real", cacheres.Href)

	delete(ssh.responses, "userandhost")
	res2, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "real", res2.Href)
}

func TestSSHCacheResolveWithError(t *testing.T) {
	ssh := newFakeResolver()
	cache := withSSHCache(ssh).(*sshCache)

	assert.Equal(t, 0, len(cache.endpoints))

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

	assert.Equal(t, 0, len(cache.endpoints))
	delete(ssh.responses, "userandhost")
	res2, err := cache.Resolve(e, "post")
	assert.Nil(t, err)
	assert.Equal(t, "", res2.Href)
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
