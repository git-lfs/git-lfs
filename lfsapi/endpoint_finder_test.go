package lfsapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpointDefaultsToOrigin(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.lfsurl": "abc",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "abc", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
}

func TestEndpointOverridesOrigin(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"lfs.url":              "abc",
		"remote.origin.lfsurl": "def",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "abc", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
}

func TestEndpointNoOverrideDefaultRemote(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.lfsurl": "abc",
		"remote.other.lfsurl":  "def",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "abc", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
}

func TestEndpointUseAlternateRemote(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.lfsurl": "abc",
		"remote.other.lfsurl":  "def",
	}))

	e := finder.Endpoint("download", "other")
	assert.Equal(t, "def", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
}

func TestEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url": "https://example.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
}

func TestBareEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url": "https://example.com/foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
}

func TestEndpointSeparateClonePushUrl(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url":     "https://example.com/foo/bar.git",
		"remote.origin.pushurl": "https://readwrite.com/foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)

	e = finder.Endpoint("upload", "")
	assert.Equal(t, "https://readwrite.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
}

func TestEndpointOverriddenSeparateClonePushLfsUrl(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url":        "https://example.com/foo/bar.git",
		"remote.origin.pushurl":    "https://readwrite.com/foo/bar.git",
		"remote.origin.lfsurl":     "https://examplelfs.com/foo/bar",
		"remote.origin.lfspushurl": "https://readwritelfs.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://examplelfs.com/foo/bar", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)

	e = finder.Endpoint("upload", "")
	assert.Equal(t, "https://readwritelfs.com/foo/bar", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
}

func TestEndpointGlobalSeparateLfsPush(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"lfs.url":     "https://readonly.com/foo/bar",
		"lfs.pushurl": "https://write.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://readonly.com/foo/bar", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)

	e = finder.Endpoint("upload", "")
	assert.Equal(t, "https://write.com/foo/bar", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
}

func TestSSHEndpointOverridden(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url":    "git@example.com:foo/bar",
		"remote.origin.lfsurl": "lfs",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "lfs", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
	assert.Equal(t, "", e.SshPort)
}

func TestSSHEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url": "ssh://git@example.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SshUserAndHost)
	assert.Equal(t, "foo/bar", e.SshPath)
	assert.Equal(t, "", e.SshPort)
}

func TestSSHCustomPortEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url": "ssh://git@example.com:9000/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SshUserAndHost)
	assert.Equal(t, "foo/bar", e.SshPath)
	assert.Equal(t, "9000", e.SshPort)
}

func TestBareSSHEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url": "git@example.com:foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "git@example.com", e.SshUserAndHost)
	assert.Equal(t, "foo/bar.git", e.SshPath)
	assert.Equal(t, "", e.SshPort)
}

func TestSSHEndpointFromGlobalLfsUrl(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"lfs.url": "git@example.com:foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git", e.Url)
	assert.Equal(t, "git@example.com", e.SshUserAndHost)
	assert.Equal(t, "foo/bar.git", e.SshPath)
	assert.Equal(t, "", e.SshPort)
}

func TestHTTPEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url": "http://example.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
	assert.Equal(t, "", e.SshPort)
}

func TestBareHTTPEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url": "http://example.com/foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
	assert.Equal(t, "", e.SshPort)
}

func TestGitEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url": "git://example.com/foo/bar",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
	assert.Equal(t, "", e.SshPort)
}

func TestGitEndpointAddsLfsSuffixWithCustomProtocol(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url": "git://example.com/foo/bar",
		"lfs.gitprotocol":   "http",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
	assert.Equal(t, "", e.SshPort)
}

func TestBareGitEndpointAddsLfsSuffix(t *testing.T) {
	finder := NewEndpointFinder(gitEnv(map[string]string{
		"remote.origin.url": "git://example.com/foo/bar.git",
	}))

	e := finder.Endpoint("download", "")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", e.Url)
	assert.Equal(t, "", e.SshUserAndHost)
	assert.Equal(t, "", e.SshPath)
	assert.Equal(t, "", e.SshPort)
}

type gitEnv map[string]string

func (e gitEnv) Get(key string) (string, bool) {
	v, ok := e[key]
	return v, ok
}

func (e gitEnv) All() map[string]string {
	return e
}
