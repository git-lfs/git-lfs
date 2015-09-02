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
	defer Config.ResetConfig()
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
	defer Config.ResetConfig()

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

func TestBatch(t *testing.T) {
	tests := map[string]bool{
		"true":     true,
		"1":        true,
		"42":       true,
		"-1":       true,
		"":         true,
		"0":        false,
		"false":    false,
		"elephant": false,
	}

	for value, expected := range tests {
		config := &Configuration{
			gitConfig: map[string]string{"lfs.batch": value},
		}

		if actual := config.BatchTransfer(); actual != expected {
			t.Errorf("lfs.batch %q == %v, not %v", value, actual, expected)
		}
	}
}

func TestBatchAbsentIsTrue(t *testing.T) {
	config := &Configuration{}

	v := config.BatchTransfer()
	assert.Equal(t, true, v)
}

func TestAccessConfig(t *testing.T) {
	type accessTest struct {
		Access        string
		PrivateAccess bool
	}

	tests := map[string]accessTest{
		"":            {"none", false},
		"basic":       {"basic", true},
		"BASIC":       {"basic", true},
		"private":     {"basic", true},
		"PRIVATE":     {"basic", true},
		"invalidauth": {"invalidauth", true},
	}

	for value, expected := range tests {
		config := &Configuration{
			gitConfig: map[string]string{
				"lfs.url":                        "http://example.com",
				"lfs.http://example.com.access":  value,
				"lfs.https://example.com.access": "bad",
			},
		}

		if access := config.Access(); access != expected.Access {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}

		if priv := config.PrivateAccess(); priv != expected.PrivateAccess {
			t.Errorf("Expected PrivateAccess() with value %q to be %v, got %v", value, expected.PrivateAccess, priv)
		}
	}
}

func TestAccessAbsentConfig(t *testing.T) {
	config := &Configuration{}
	assert.Equal(t, "none", config.Access())
	assert.Equal(t, false, config.PrivateAccess())
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

func TestFetchPruneConfigDefault(t *testing.T) {
	config := &Configuration{}
	fp := config.FetchPruneConfig()

	assert.Equal(t, 7, fp.FetchRecentRefsDays)
	assert.Equal(t, 0, fp.FetchRecentCommitsDays)
	assert.Equal(t, 3, fp.PruneOffsetDays)
	assert.Equal(t, true, fp.FetchRecentRefsIncludeRemotes)

}
func TestFetchPruneConfigCustom(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.fetchrecentrefsdays":    "12",
			"lfs.fetchrecentremoterefs":  "false",
			"lfs.fetchrecentcommitsdays": "9",
			"lfs.pruneoffsetdays":        "30",
		},
	}
	fp := config.FetchPruneConfig()

	assert.Equal(t, 12, fp.FetchRecentRefsDays)
	assert.Equal(t, 9, fp.FetchRecentCommitsDays)
	assert.Equal(t, 30, fp.PruneOffsetDays)
	assert.Equal(t, false, fp.FetchRecentRefsIncludeRemotes)
}

// only used for tests
func (c *Configuration) SetConfig(key, value string) {
	if c.loadGitConfig() {
		c.loading.Lock()
		c.origConfig = make(map[string]string)
		for k, v := range c.gitConfig {
			c.origConfig[k] = v
		}
		c.loading.Unlock()
	}

	c.gitConfig[key] = value
}

func (c *Configuration) ResetConfig() {
	c.loading.Lock()
	c.gitConfig = make(map[string]string)
	for k, v := range c.origConfig {
		c.gitConfig[k] = v
	}
	c.loading.Unlock()
}
