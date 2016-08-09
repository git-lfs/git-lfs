package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEndpointDefaultsToOrigin(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"remote.origin.lfsurl": "abc"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "abc", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointOverridesOrigin(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.url":              "abc",
			"remote.origin.lfsurl": "def",
		},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "abc", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointNoOverrideDefaultRemote(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"remote.origin.lfsurl": "abc",
			"remote.other.lfsurl":  "def",
		},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "abc", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointUseAlternateRemote(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"remote.origin.lfsurl": "abc",
			"remote.other.lfsurl":  "def",
		},
	})

	cfg.CurrentRemote = "other"

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "def", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointAddsLfsSuffix(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"remote.origin.url": "https://example.com/foo/bar"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestBareEndpointAddsLfsSuffix(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"remote.origin.url": "https://example.com/foo/bar.git"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointSeparateClonePushUrl(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"remote.origin.url":     "https://example.com/foo/bar.git",
			"remote.origin.pushurl": "https://readwrite.com/foo/bar.git"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)

	endpoint = cfg.Endpoint("upload")
	assert.Equal(t, "https://readwrite.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointOverriddenSeparateClonePushLfsUrl(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"remote.origin.url":        "https://example.com/foo/bar.git",
			"remote.origin.pushurl":    "https://readwrite.com/foo/bar.git",
			"remote.origin.lfsurl":     "https://examplelfs.com/foo/bar",
			"remote.origin.lfspushurl": "https://readwritelfs.com/foo/bar"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://examplelfs.com/foo/bar", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)

	endpoint = cfg.Endpoint("upload")
	assert.Equal(t, "https://readwritelfs.com/foo/bar", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestEndpointGlobalSeparateLfsPush(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.url":     "https://readonly.com/foo/bar",
			"lfs.pushurl": "https://write.com/foo/bar",
		},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://readonly.com/foo/bar", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)

	endpoint = cfg.Endpoint("upload")
	assert.Equal(t, "https://write.com/foo/bar", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
}

func TestSSHEndpointOverridden(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"remote.origin.url":    "git@example.com:foo/bar",
			"remote.origin.lfsurl": "lfs",
		},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestSSHEndpointAddsLfsSuffix(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"remote.origin.url": "ssh://git@example.com/foo/bar"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "git@example.com", endpoint.SshUserAndHost)
	assert.Equal(t, "foo/bar", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestSSHCustomPortEndpointAddsLfsSuffix(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"remote.origin.url": "ssh://git@example.com:9000/foo/bar"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "git@example.com", endpoint.SshUserAndHost)
	assert.Equal(t, "foo/bar", endpoint.SshPath)
	assert.Equal(t, "9000", endpoint.SshPort)
}

func TestBareSSHEndpointAddsLfsSuffix(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"remote.origin.url": "git@example.com:foo/bar.git"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "git@example.com", endpoint.SshUserAndHost)
	assert.Equal(t, "foo/bar.git", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestSSHEndpointFromGlobalLfsUrl(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"lfs.url": "git@example.com:foo/bar.git"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git", endpoint.Url)
	assert.Equal(t, "git@example.com", endpoint.SshUserAndHost)
	assert.Equal(t, "foo/bar.git", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestHTTPEndpointAddsLfsSuffix(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"remote.origin.url": "http://example.com/foo/bar"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestBareHTTPEndpointAddsLfsSuffix(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"remote.origin.url": "http://example.com/foo/bar.git"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestGitEndpointAddsLfsSuffix(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"remote.origin.url": "git://example.com/foo/bar"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestGitEndpointAddsLfsSuffixWithCustomProtocol(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"remote.origin.url": "git://example.com/foo/bar",
			"lfs.gitprotocol":   "http",
		},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "http://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestBareGitEndpointAddsLfsSuffix(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{"remote.origin.url": "git://example.com/foo/bar.git"},
	})

	endpoint := cfg.Endpoint("download")
	assert.Equal(t, "https://example.com/foo/bar.git/info/lfs", endpoint.Url)
	assert.Equal(t, "", endpoint.SshUserAndHost)
	assert.Equal(t, "", endpoint.SshPath)
	assert.Equal(t, "", endpoint.SshPort)
}

func TestConcurrentTransfersSetValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.concurrenttransfers": "5",
		},
	})

	n := cfg.ConcurrentTransfers()
	assert.Equal(t, 5, n)
}

func TestConcurrentTransfersDefault(t *testing.T) {
	cfg := NewFrom(Values{})

	n := cfg.ConcurrentTransfers()
	assert.Equal(t, 3, n)
}

func TestConcurrentTransfersZeroValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.concurrenttransfers": "0",
		},
	})

	n := cfg.ConcurrentTransfers()
	assert.Equal(t, 3, n)
}

func TestConcurrentTransfersNonNumeric(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.concurrenttransfers": "elephant",
		},
	})

	n := cfg.ConcurrentTransfers()
	assert.Equal(t, 3, n)
}

func TestConcurrentTransfersNegativeValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.concurrenttransfers": "-5",
		},
	})

	n := cfg.ConcurrentTransfers()
	assert.Equal(t, 3, n)
}

func TestBasicTransfersOnlySetValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.basictransfersonly": "true",
		},
	})

	b := cfg.BasicTransfersOnly()
	assert.Equal(t, true, b)
}

func TestBasicTransfersOnlyDefault(t *testing.T) {
	cfg := NewFrom(Values{})

	b := cfg.BasicTransfersOnly()
	assert.Equal(t, false, b)
}

func TestBasicTransfersOnlyInvalidValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.basictransfersonly": "wat",
		},
	})

	b := cfg.BasicTransfersOnly()
	assert.Equal(t, false, b)
}

func TestTusTransfersAllowedSetValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.tustransfers": "true",
		},
	})

	b := cfg.TusTransfersAllowed()
	assert.Equal(t, true, b)
}

func TestTusTransfersAllowedDefault(t *testing.T) {
	cfg := NewFrom(Values{})

	b := cfg.TusTransfersAllowed()
	assert.Equal(t, false, b)
}

func TestTusTransfersAllowedInvalidValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.tustransfers": "wat",
		},
	})

	b := cfg.TusTransfersAllowed()
	assert.Equal(t, false, b)
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
		cfg := NewFrom(Values{
			Git: map[string]string{"lfs.batch": value},
		})

		if actual := cfg.BatchTransfer(); actual != expected {
			t.Errorf("lfs.batch %q == %v, not %v", value, actual, expected)
		}
	}
}

func TestBatchAbsentIsTrue(t *testing.T) {
	cfg := NewFrom(Values{})
	v := cfg.BatchTransfer()
	assert.True(t, v)
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
		cfg := NewFrom(Values{
			Git: map[string]string{
				"lfs.url":                        "http://example.com",
				"lfs.http://example.com.access":  value,
				"lfs.https://example.com.access": "bad",
			},
		})

		if access := cfg.Access("download"); access != expected.Access {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}
		if access := cfg.Access("upload"); access != expected.Access {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}

		if priv := cfg.PrivateAccess("download"); priv != expected.PrivateAccess {
			t.Errorf("Expected PrivateAccess() with value %q to be %v, got %v", value, expected.PrivateAccess, priv)
		}
		if priv := cfg.PrivateAccess("upload"); priv != expected.PrivateAccess {
			t.Errorf("Expected PrivateAccess() with value %q to be %v, got %v", value, expected.PrivateAccess, priv)
		}
	}

	// Test again but with separate push url
	for value, expected := range tests {
		cfg := NewFrom(Values{
			Git: map[string]string{
				"lfs.url":                           "http://example.com",
				"lfs.pushurl":                       "http://examplepush.com",
				"lfs.http://example.com.access":     value,
				"lfs.http://examplepush.com.access": value,
				"lfs.https://example.com.access":    "bad",
			},
		})

		if access := cfg.Access("download"); access != expected.Access {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}
		if access := cfg.Access("upload"); access != expected.Access {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}

		if priv := cfg.PrivateAccess("download"); priv != expected.PrivateAccess {
			t.Errorf("Expected PrivateAccess() with value %q to be %v, got %v", value, expected.PrivateAccess, priv)
		}
		if priv := cfg.PrivateAccess("upload"); priv != expected.PrivateAccess {
			t.Errorf("Expected PrivateAccess() with value %q to be %v, got %v", value, expected.PrivateAccess, priv)
		}
	}

}

func TestAccessAbsentConfig(t *testing.T) {
	cfg := NewFrom(Values{})
	assert.Equal(t, "none", cfg.Access("download"))
	assert.Equal(t, "none", cfg.Access("upload"))
	assert.False(t, cfg.PrivateAccess("download"))
	assert.False(t, cfg.PrivateAccess("upload"))
}

func TestLoadValidExtension(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{},
	})

	cfg.extensions = map[string]Extension{
		"foo": Extension{
			"foo",
			"foo-clean %f",
			"foo-smudge %f",
			2,
		},
	}

	ext := cfg.Extensions()["foo"]

	assert.Equal(t, "foo", ext.Name)
	assert.Equal(t, "foo-clean %f", ext.Clean)
	assert.Equal(t, "foo-smudge %f", ext.Smudge)
	assert.Equal(t, 2, ext.Priority)
}

func TestLoadInvalidExtension(t *testing.T) {
	cfg := NewFrom(Values{})
	ext := cfg.Extensions()["foo"]

	assert.Equal(t, "", ext.Name)
	assert.Equal(t, "", ext.Clean)
	assert.Equal(t, "", ext.Smudge)
	assert.Equal(t, 0, ext.Priority)
}

func TestFetchPruneConfigDefault(t *testing.T) {
	cfg := NewFrom(Values{})
	fp := cfg.FetchPruneConfig()

	assert.Equal(t, 7, fp.FetchRecentRefsDays)
	assert.Equal(t, 0, fp.FetchRecentCommitsDays)
	assert.Equal(t, 3, fp.PruneOffsetDays)
	assert.True(t, fp.FetchRecentRefsIncludeRemotes)
	assert.Equal(t, 3, fp.PruneOffsetDays)
	assert.Equal(t, "origin", fp.PruneRemoteName)
	assert.False(t, fp.PruneVerifyRemoteAlways)

}
func TestFetchPruneConfigCustom(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.fetchrecentrefsdays":     "12",
			"lfs.fetchrecentremoterefs":   "false",
			"lfs.fetchrecentcommitsdays":  "9",
			"lfs.pruneoffsetdays":         "30",
			"lfs.pruneverifyremotealways": "true",
			"lfs.pruneremotetocheck":      "upstream",
		},
	})
	fp := cfg.FetchPruneConfig()

	assert.Equal(t, 12, fp.FetchRecentRefsDays)
	assert.Equal(t, 9, fp.FetchRecentCommitsDays)
	assert.False(t, fp.FetchRecentRefsIncludeRemotes)
	assert.Equal(t, 30, fp.PruneOffsetDays)
	assert.Equal(t, "upstream", fp.PruneRemoteName)
	assert.True(t, fp.PruneVerifyRemoteAlways)
}

func TestFetchIncludeExcludesAreCleaned(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"lfs.fetchinclude": "/path/to/clean/",
			"lfs.fetchexclude": "/other/path/to/clean/",
		},
	})

	assert.Equal(t, []string{"/path/to/clean"}, cfg.FetchIncludePaths())
	assert.Equal(t, []string{"/other/path/to/clean"}, cfg.FetchExcludePaths())
}

func TestUnmarshalMultipleTypes(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string]string{
			"string": "string",
			"int":    "1",
			"bool":   "true",
		},
		Os: map[string]string{
			"string": "string",
			"int":    "1",
			"bool":   "true",
		},
	})

	v := &struct {
		GitString string `git:"string"`
		GitInt    int    `git:"int"`
		GitBool   bool   `git:"bool"`
		OsString  string `os:"string"`
		OsInt     int    `os:"int"`
		OsBool    bool   `os:"bool"`
	}{}

	assert.Nil(t, cfg.Unmarshal(v))

	assert.Equal(t, "string", v.GitString)
	assert.Equal(t, 1, v.GitInt)
	assert.Equal(t, true, v.GitBool)
	assert.Equal(t, "string", v.OsString)
	assert.Equal(t, 1, v.OsInt)
	assert.Equal(t, true, v.OsBool)
}

func TestUnmarshalErrsOnNonPointerType(t *testing.T) {
	type T struct {
		Foo string `git:"foo"`
	}

	cfg := NewFrom(Values{})

	err := cfg.Unmarshal(T{})

	assert.Equal(t, "lfs/config: unable to parse non-pointer type of config.T", err.Error())
}

func TestUnmarshalLeavesNonZeroValuesWhenKeysEmpty(t *testing.T) {
	v := &struct {
		String string `git:"string"`
		Int    int    `git:"int"`
		Bool   bool   `git:"bool"`
	}{"foo", 1, true}

	cfg := NewFrom(Values{})

	err := cfg.Unmarshal(v)

	assert.Nil(t, err)
	assert.Equal(t, "foo", v.String)
	assert.Equal(t, 1, v.Int)
	assert.Equal(t, true, v.Bool)
}

func TestUnmarshalOverridesNonZeroValuesWhenValuesPresent(t *testing.T) {
	v := &struct {
		String string `git:"string"`
		Int    int    `git:"int"`
		Bool   bool   `git:"bool"`
	}{"foo", 1, true}

	cfg := NewFrom(Values{
		Git: map[string]string{
			"string": "bar",
			"int":    "2",
			"bool":   "false",
		},
	})

	err := cfg.Unmarshal(v)

	assert.Nil(t, err)
	assert.Equal(t, "bar", v.String)
	assert.Equal(t, 2, v.Int)
	assert.Equal(t, false, v.Bool)
}

func TestUnmarshalDoesNotAllowBothOsAndGitTags(t *testing.T) {
	v := &struct {
		String string `git:"string" os:"STRING"`
	}{}

	cfg := NewFrom(Values{})

	err := cfg.Unmarshal(v)

	assert.Equal(t, "lfs/config: ambiguous tags", err.Error())
}

func TestUnmarshalIgnoresUnknownEnvironments(t *testing.T) {
	v := &struct {
		String string `unknown:"string"`
	}{}

	cfg := NewFrom(Values{})

	assert.Nil(t, cfg.Unmarshal(v))
}

func TestUnmarshalErrsOnUnsupportedTypes(t *testing.T) {
	v := &struct {
		Unsupported time.Duration `git:"duration"`
	}{}

	cfg := NewFrom(Values{
		Git: map[string]string{"duration": "foo"},
	})

	err := cfg.Unmarshal(v)

	assert.Equal(t, "lfs/config: unsupported target type for field \"Unsupported\": time.Duration", err.Error())
}
