package lfs

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestEndpointDefaultsToOrigin(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.lfsurl": "abc"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "abc", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointOverridesOrigin(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.url":              "abc",
			"remote.origin.lfsurl": "def",
		},
		remotes: []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "abc", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointNoOverrideDefaultRemote(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.lfsurl": "abc",
			"remote.other.lfsurl":  "def",
		},
		remotes: []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "abc", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointUseAlternateRemote(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.lfsurl": "abc",
			"remote.other.lfsurl":  "def",
		},
		remotes: []string{},
	}

	config.CurrentRemote = "other"

	endpoint := config.Endpoint()
	assert.Equal(t, "def", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "https://example.com/foo/bar"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestBareEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "https://example.com/foo/bar.git"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestSSHEndpointOverridden(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.url":    "git@example.com:foo/bar",
			"remote.origin.lfsurl": "lfs",
		},
		remotes: []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestSSHEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "git@example.com:foo/bar"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "git@example.com", endpoint.SshUserAndHost)
	assert.Equal(t, "foo/bar", endpoint.SshPath)
}

func TestBareSSHEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "git@example.com:foo/bar.git"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "git@example.com", endpoint.SshUserAndHost)
	assert.Equal(t, "foo/bar.git", endpoint.SshPath)
}

func TestHTTPEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "http://example.com/foo/bar"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestBareHTTPEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "http://example.com/foo/bar.git"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestObjectUrl(t *testing.T) {
	tests := map[string]string{
		"http://example.com":      "http://example.com/objects/oid",
		"http://example.com/":     "http://example.com/objects/oid",
		"http://example.com/foo":  "http://example.com/foo/objects/oid",
		"http://example.com/foo/": "http://example.com/foo/objects/oid",
	}

	for endpoint, expected := range tests {
		Config.SetConfig("lfs.url", endpoint)
		u, err := Config.ObjectUrl("oid")
		if err != nil {
			t.Errorf("Error building URL for %s: %s", endpoint, err)
		} else {
			if actual := u.String(); expected != actual {
				t.Errorf("Expected %s, got %s", expected, u.String())
			}
		}
	}
}

func TestObjectsUrl(t *testing.T) {
	tests := map[string]string{
		"http://example.com":      "http://example.com/objects",
		"http://example.com/":     "http://example.com/objects",
		"http://example.com/foo":  "http://example.com/foo/objects",
		"http://example.com/foo/": "http://example.com/foo/objects",
	}

	for endpoint, expected := range tests {
		Config.SetConfig("lfs.url", endpoint)
		u, err := Config.ObjectUrl("")
		if err != nil {
			t.Errorf("Error building URL for %s: %s", endpoint, err)
		} else {
			if actual := u.String(); expected != actual {
				t.Errorf("Expected %s, got %s", expected, u.String())
			}
		}
	}
}
