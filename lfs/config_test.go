package lfs

import (
	"testing"

	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
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
	assert.Equal(t, "", endpoint.SshPort)
}

func TestSSHEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "ssh://git@example.com/foo/bar"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "git@example.com", endpoint.SshUserAndHost)
	assert.Equal(t, "foo/bar", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestSSHCustomPortEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "ssh://git@example.com:9000/foo/bar"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "git@example.com", endpoint.SshUserAndHost)
	assert.Equal(t, "foo/bar", endpoint.SshPath)
	assert.Equal(t, "9000", endpoint.SshPort)
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
	assert.Equal(t, "", endpoint.SshPort)
}

func TestSSHEndpointFromGlobalLfsUrl(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"lfs.url": "git@example.com:foo/bar.git"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint()
	assert.Equal(t, "https://example.com/foo/bar.git", endpoint.Url)
	assert.Equal(t, "git@example.com", endpoint.SshUserAndHost)
	assert.Equal(t, "foo/bar.git", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
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
	assert.Equal(t, "", endpoint.SshPort)
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
	assert.Equal(t, "", endpoint.SshPort)
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

func TestConcurrentTransfersSetValue(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.concurrenttransfers": "5",
		},
	}

	n := config.ConcurrentTransfers()
	assert.Equal(t, 5, n)
}

func TestConcurrentTransfersDefault(t *testing.T) {
	config := &Configuration{}

	n := config.ConcurrentTransfers()
	assert.Equal(t, 3, n)
}

func TestConcurrentTransfersZeroValue(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.concurrenttransfers": "0",
		},
	}

	n := config.ConcurrentTransfers()
	assert.Equal(t, 3, n)
}

func TestConcurrentTransfersNonNumeric(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.concurrenttransfers": "elephant",
		},
	}

	n := config.ConcurrentTransfers()
	assert.Equal(t, 3, n)
}

func TestConcurrentTransfersNegativeValue(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.concurrenttransfers": "-5",
		},
	}

	n := config.ConcurrentTransfers()
	assert.Equal(t, 3, n)
}

func TestBatchTrue(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.batch": "true",
		},
	}

	v := config.BatchTransfer()
	assert.Equal(t, true, v)
}

func TestBatchNumeric1IsTrue(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.batch": "1",
		},
	}

	v := config.BatchTransfer()
	assert.Equal(t, true, v)
}

func TestBatchNumeric0IsFalse(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.batch": "0",
		},
	}

	v := config.BatchTransfer()
	assert.Equal(t, false, v)
}

func TestBatchOtherNumericsAreTrue(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.batch": "42",
		},
	}

	v := config.BatchTransfer()
	assert.Equal(t, true, v)
}

func TestBatchNegativeNumericsAreTrue(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.batch": "-1",
		},
	}

	v := config.BatchTransfer()
	assert.Equal(t, true, v)
}

func TestBatchNonBooleanIsFalse(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.batch": "elephant",
		},
	}

	v := config.BatchTransfer()
	assert.Equal(t, false, v)
}

func TestBatchPresentButBlankIsTrue(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.batch": "",
		},
	}

	v := config.BatchTransfer()
	assert.Equal(t, true, v)
}

func TestBatchAbsentIsFalse(t *testing.T) {
	config := &Configuration{}

	v := config.BatchTransfer()
	assert.Equal(t, false, v)
}

func TestLoadValidExtension(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{},
		extensions: map[string]Extension{
			"foo": Extension{
				"foo",
				"foo-clean %f",
				"foo-smudge %f",
				2,
			},
		},
	}

	ext := config.Extensions()["foo"]

	assert.Equal(t, "foo", ext.Name)
	assert.Equal(t, "foo-clean %f", ext.Clean)
	assert.Equal(t, "foo-smudge %f", ext.Smudge)
	assert.Equal(t, 2, ext.Priority)
}

func TestLoadInvalidExtension(t *testing.T) {
	config := &Configuration{}

	ext := config.Extensions()["foo"]

	assert.Equal(t, "", ext.Name)
	assert.Equal(t, "", ext.Clean)
	assert.Equal(t, "", ext.Smudge)
	assert.Equal(t, 0, ext.Priority)
}
