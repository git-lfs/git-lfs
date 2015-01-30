package hawser

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestEndpointDefaultsToOrigin(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.media": "abc"},
		remotes:   []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointOverridesOrigin(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"media.url":           "abc",
			"remote.origin.media": "def",
		},
		remotes: []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
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
