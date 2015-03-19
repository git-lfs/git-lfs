package lfs

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestEndpointDefaultsToOrigin(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.lfs": "abc"},
		remotes:   []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointOverridesOrigin(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.url":           "abc",
			"remote.origin.lfs": "def",
		},
		remotes: []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointNoOverrideDefaultRemote(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.lfs": "abc",
			"remote.other.lfs":  "def",
		},
		remotes: []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointUseAlternateRemote(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.lfs": "abc",
			"remote.other.lfs":  "def",
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

	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", config.Endpoint())
}

func TestBareEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "https://example.com/foo/bar.git"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", config.Endpoint())
}

func TestSSHEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "git@example.com:foo/bar"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", config.Endpoint())
}

func TestBareSSHEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "git@example.com:foo/bar.git"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", config.Endpoint())
}

func TestHTTPEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "http://example.com/foo/bar"},
		remotes:   []string{},
	}

	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", config.Endpoint())
}

func TestBareHTTPEndpointAddsMediaSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "http://example.com/foo/bar.git"},
		remotes:   []string{},
	}

	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", config.Endpoint())
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
		Config.SetConfig("lfs.url", endpoint)
		assert.Equal(t, expected, Config.ObjectUrl(oid).String())
	}
}
