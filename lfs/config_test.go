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

	endpoint := config.Endpoint("download")
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

	endpoint := config.Endpoint("download")
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

	endpoint := config.Endpoint("download")
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

	endpoint := config.Endpoint("download")
	assert.Equal(t, "def", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "https://example.com/foo/bar"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestBareEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "https://example.com/foo/bar.git"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointSeparateClonePushUrl(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.url":     "https://example.com/foo/bar.git",
			"remote.origin.pushurl": "https://readwrite.com/foo/bar.git"},
		remotes: []string{},
	}

	endpoint := config.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)

	endpoint = config.Endpoint("upload")
	assert.Equal(t, "https://readwrite.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointOverriddenSeparateClonePushLfsUrl(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.url":        "https://example.com/foo/bar.git",
			"remote.origin.pushurl":    "https://readwrite.com/foo/bar.git",
			"remote.origin.lfsurl":     "https://examplelfs.com/foo/bar",
			"remote.origin.lfspushurl": "https://readwritelfs.com/foo/bar"},
		remotes: []string{},
	}

	endpoint := config.Endpoint("download")
	assert.Equal(t, "https://examplelfs.com/foo/bar", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)

	endpoint = config.Endpoint("upload")
	assert.Equal(t, "https://readwritelfs.com/foo/bar", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointGlobalSeparateLfsPush(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.url":     "https://readonly.com/foo/bar",
			"lfs.pushurl": "https://write.com/foo/bar",
		},
		remotes: []string{},
	}

	endpoint := config.Endpoint("download")
	assert.Equal(t, "https://readonly.com/foo/bar", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)

	endpoint = config.Endpoint("upload")
	assert.Equal(t, "https://write.com/foo/bar", endpoint.Url)
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

	endpoint := config.Endpoint("download")
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

	endpoint := config.Endpoint("download")
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

	endpoint := config.Endpoint("download")
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

	endpoint := config.Endpoint("download")
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

	endpoint := config.Endpoint("download")
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

	endpoint := config.Endpoint("download")
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

	endpoint := config.Endpoint("download")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestGitEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "git://example.com/foo/bar"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestGitEndpointAddsLfsSuffixWithCustomProtocol(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"remote.origin.url": "git://example.com/foo/bar",
			"lfs.gitprotocol":   "http",
		},
		remotes: []string{},
	}

	endpoint := config.Endpoint("download")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestBareGitEndpointAddsLfsSuffix(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{"remote.origin.url": "git://example.com/foo/bar.git"},
		remotes:   []string{},
	}

	endpoint := config.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
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
		"":         true,
		"true":     true,
		"1":        true,
		"42":       false,
		"-1":       false,
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

		if access := config.Access("download"); access != expected.Access {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}
		if access := config.Access("upload"); access != expected.Access {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}

		if priv := config.PrivateAccess("download"); priv != expected.PrivateAccess {
			t.Errorf("Expected PrivateAccess() with value %q to be %v, got %v", value, expected.PrivateAccess, priv)
		}
		if priv := config.PrivateAccess("upload"); priv != expected.PrivateAccess {
			t.Errorf("Expected PrivateAccess() with value %q to be %v, got %v", value, expected.PrivateAccess, priv)
		}
	}

	// Test again but with separate push url
	for value, expected := range tests {
		config := &Configuration{
			gitConfig: map[string]string{
				"lfs.url":                           "http://example.com",
				"lfs.pushurl":                       "http://examplepush.com",
				"lfs.http://example.com.access":     value,
				"lfs.http://examplepush.com.access": value,
				"lfs.https://example.com.access":    "bad",
			},
		}

		if access := config.Access("download"); access != expected.Access {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}
		if access := config.Access("upload"); access != expected.Access {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}

		if priv := config.PrivateAccess("download"); priv != expected.PrivateAccess {
			t.Errorf("Expected PrivateAccess() with value %q to be %v, got %v", value, expected.PrivateAccess, priv)
		}
		if priv := config.PrivateAccess("upload"); priv != expected.PrivateAccess {
			t.Errorf("Expected PrivateAccess() with value %q to be %v, got %v", value, expected.PrivateAccess, priv)
		}
	}

}

func TestAccessAbsentConfig(t *testing.T) {
	config := &Configuration{}
	assert.Equal(t, "none", config.Access("download"))
	assert.Equal(t, "none", config.Access("upload"))
	assert.Equal(t, false, config.PrivateAccess("download"))
	assert.Equal(t, false, config.PrivateAccess("upload"))
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
	assert.Equal(t, 3, fp.PruneOffsetDays)
	assert.Equal(t, "origin", fp.PruneRemoteName)
	assert.Equal(t, false, fp.PruneVerifyRemoteAlways)

}
func TestFetchPruneConfigCustom(t *testing.T) {
	config := &Configuration{
		gitConfig: map[string]string{
			"lfs.fetchrecentrefsdays":     "12",
			"lfs.fetchrecentremoterefs":   "false",
			"lfs.fetchrecentcommitsdays":  "9",
			"lfs.pruneoffsetdays":         "30",
			"lfs.pruneverifyremotealways": "true",
			"lfs.pruneremotetocheck":      "upstream",
		},
	}
	fp := config.FetchPruneConfig()

	assert.Equal(t, 12, fp.FetchRecentRefsDays)
	assert.Equal(t, 9, fp.FetchRecentCommitsDays)
	assert.Equal(t, false, fp.FetchRecentRefsIncludeRemotes)
	assert.Equal(t, 30, fp.PruneOffsetDays)
	assert.Equal(t, "upstream", fp.PruneRemoteName)
	assert.Equal(t, true, fp.PruneVerifyRemoteAlways)
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

func TestFetchIncludeExcludesAreCleaned(t *testing.T) {
	config := NewFromValues(map[string]string{
		"lfs.fetchinclude": "/path/to/clean/",
		"lfs.fetchexclude": "/other/path/to/clean/",
	})

	assert.Equal(t, []string{"/path/to/clean"}, config.FetchIncludePaths())
	assert.Equal(t, []string{"/other/path/to/clean"}, config.FetchExcludePaths())
}
