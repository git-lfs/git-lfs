package config

import (
	"testing"

	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCanonicalization(t *testing.T) {
	vals := map[string][]string{
		"user.name":                   []string{"Pat Doe"},
		"branch.MixedCase.pushremote": []string{"Somewhere"},
		"http.https://example.com/BIG-TEXT.git.extraheader": []string{"X-Foo: Bar"},
	}

	fetcher := GitFetcher{vals: vals}
	assert.Equal(t, []string{"Somewhere"}, fetcher.GetAll("bRanch.MixedCase.pushRemote"))
	assert.Equal(t, []string{"Somewhere"}, fetcher.GetAll("branch.MixedCase.pushremote"))
	assert.Equal(t, []string(nil), fetcher.GetAll("branch.mixedcase.pushremote"))
	assert.Equal(t, []string{"Pat Doe"}, fetcher.GetAll("user.name"))
	assert.Equal(t, []string{"Pat Doe"}, fetcher.GetAll("User.Name"))
	assert.Equal(t, []string{"X-Foo: Bar"}, fetcher.GetAll("http.https://example.com/BIG-TEXT.git.extraheader"))
	assert.Equal(t, []string{"X-Foo: Bar"}, fetcher.GetAll("http.https://example.com/BIG-TEXT.git.extraHeader"))
	assert.Equal(t, []string(nil), fetcher.GetAll("http.https://example.com/big-text.git.extraHeader"))
}

func TestReadGitConfigKeys(t *testing.T) {
	tests := []struct {
		name         string
		config       string
		expectedKeys map[string][]string
		safeKeys     []string
	}{
		{
			name:   "lfs remote key",
			config: "remote.origin.lfsurl=https://example.com/foo/bar.git/info/lfs",
			expectedKeys: map[string][]string{
				"remote.origin.lfsurl": []string{"https://example.com/foo/bar.git/info/lfs"},
			},
			safeKeys: []string{"remote.origin.lfsurl"},
		},
		{
			name:   "lfs remote key (multi-part subsection)",
			config: "remote.other.origin.lfsurl=https://example.com/foo/bar.git/info/lfs",
			expectedKeys: map[string][]string{
				"remote.other.origin.lfsurl": []string{"https://example.com/foo/bar.git/info/lfs"},
			},
			safeKeys: []string{"remote.other.origin.lfsurl"},
		},
		{
			name:   "non-lfs remote key",
			config: "remote.origin.pushurl=https://example.com/foo/bar.git",
			expectedKeys: map[string][]string{
				"remote.origin.pushurl": []string{"https://example.com/foo/bar.git"},
			},
		},
		{
			name:   "non-lfs remote key (multi-part subsection)",
			config: "remote.other.origin.pushurl=https://example.com/foo/bar.git",
			expectedKeys: map[string][]string{
				"remote.other.origin.pushurl": []string{"https://example.com/foo/bar.git"},
			},
		},
		{
			name: "non-lfs remote key (multiple values)",
			config: "remote.origin.pushurl=https://example.com/foo/bar.git\n" +
				"remote.origin.pushurl=https://example.com/foo/baz.git",
			expectedKeys: map[string][]string{
				"remote.origin.pushurl": []string{
					"https://example.com/foo/bar.git",
					"https://example.com/foo/baz.git",
				},
			},
		},
		{
			name:   "non-lfs remote key (empty value)",
			config: "remote.origin.pushurl=",
			expectedKeys: map[string][]string{
				"remote.origin.pushurl": []string{""},
			},
		},
		{
			name:   "non-lfs remote key (no subsection)",
			config: "remote.origin=unknown",
			expectedKeys: map[string][]string{
				"remote.origin": []string{"unknown"},
			},
		},
		{
			name:   "lfs remote default key",
			config: "remote.lfsdefault=upstream",
			expectedKeys: map[string][]string{
				"remote.lfsdefault": []string{"upstream"},
			},
		},
		{
			name:   "non-lfs remote default key",
			config: "remote.pushdefault=upstream",
			expectedKeys: map[string][]string{
				"remote.pushdefault": []string{"upstream"},
			},
		},
		{
			name:   "lfs access key",
			config: "lfs.https://example.com.access=basic",
			expectedKeys: map[string][]string{
				"lfs.https://example.com.access": []string{"basic"},
			},
			safeKeys: []string{"lfs.https://example.com.access"},
		},
		{
			name:   "non-lfs access key",
			config: "other.example.access=none",
			expectedKeys: map[string][]string{
				"other.example.access": []string{"none"},
			},
		},
	}

	assert := func(t *testing.T, config string, onlySafeKeys bool, expectedKeys map[string][]string) {
		source := git.ParseConfigLines(config, onlySafeKeys)
		fetcher, _, _ := readGitConfig(source)

		if expectedKeys == nil {
			assert.Empty(t, fetcher.All())
		} else {
			assert.Equal(t, expectedKeys, fetcher.All())
		}
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert(t, test.config, false, test.expectedKeys)
		})

		expectedSafeKeys := make(map[string][]string, len(test.safeKeys))
		for _, key := range test.safeKeys {
			values, ok := test.expectedKeys[key]
			require.True(t, ok, "safe key %q has no expected value", key)
			expectedSafeKeys[key] = values
		}

		t.Run(test.name+" (safe keys only)", func(t *testing.T) {
			assert(t, test.config, true, expectedSafeKeys)
		})
	}
}
