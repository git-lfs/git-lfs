package hawser

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestEndpointDefaultsToOrigin(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.hawser": "abc"},
		remotes:   []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointOverridesOrigin(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"hawser.url":           "abc",
			"remote.origin.hawser": "def",
		},
		remotes: []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointNoOverrideDefaultRemote(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.hawser": "abc",
			"remote.other.hawser":  "def",
		},
		remotes: []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointUseAlternateRemote(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.hawser": "abc",
			"remote.other.hawser":  "def",
		},
		remotes: []string{},
	}

	config.CurrentRemote = "other"

	assert.Equal(t, "def", config.Endpoint())
}

func TestEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "https://example.com/foo/bar"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/media", config.Endpoint())
}

func TestBareEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "https://example.com/foo/bar.git"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/media", config.Endpoint())
}

func TestSSHEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "git@example.com:foo/bar"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/media", config.Endpoint())
}

func TestBareSSHEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "git@example.com:foo/bar.git"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/media", config.Endpoint())
}

func TestHTTPEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "http://example.com/foo/bar"},
		remotes:   []string{},
	}

	assert.Equal(t, "http://example.com/foo/bar.git/info/media", config.Endpoint())
}

func TestBareHTTPEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "http://example.com/foo/bar.git"},
		remotes:   []string{},
	}

	assert.Equal(t, "http://example.com/foo/bar.git/info/media", config.Endpoint())
}

func TestObjectUrl(t *testing.T) {
	oid := "oid"
	tests := map[string]string{
		"http://example.com":      "http://example.com/objects/oid",
		"http://example.com/":     "http://example.com/objects/oid",
		"http://example.com/foo":  "http://example.com/foo/objects/oid",
		"http://example.com/foo/": "http://example.com/foo/objects/oid",
	}

	for endpoint, expected := range tests {
		Config.SetConfig("hawser.url", endpoint)
		assert.Equal(t, expected, Config.ObjectUrl(oid).String())
	}
}
