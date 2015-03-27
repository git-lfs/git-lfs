package lfs

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestEndpointDefaultsToOrigin(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.lfs_url": "abc"},
		remotes:   []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointOverridesOrigin(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.url":               "abc",
			"remote.origin.lfs_url": "def",
		},
		remotes: []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointNoOverrideDefaultRemote(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.lfs_url": "abc",
			"remote.other.lfs_url":  "def",
		},
		remotes: []string{},
	}

	assert.Equal(t, "abc", config.Endpoint())
}

func TestEndpointUseAlternateRemote(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.lfs_url": "abc",
			"remote.other.lfs_url":  "def",
		},
		remotes: []string{},
	}

	config.CurrentRemote = "other"

	assert.Equal(t, "def", config.Endpoint())
}

func TestEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "https://example.com/foo/bar"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", config.Endpoint())
}

func TestBareEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "https://example.com/foo/bar.git"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", config.Endpoint())
}

func TestSSHEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "git@example.com:foo/bar"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", config.Endpoint())
}

func TestBareSSHEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "git@example.com:foo/bar.git"},
		remotes:   []string{},
	}

	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", config.Endpoint())
}

func TestHTTPEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "http://example.com/foo/bar"},
		remotes:   []string{},
	}

	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", config.Endpoint())
}

func TestBareHTTPEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "http://example.com/foo/bar.git"},
		remotes:   []string{},
	}

	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", config.Endpoint())
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
