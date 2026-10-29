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
		name             string
		config           string
		expectedKeys     map[string][]string
		safeKeys         []string
		expectedExts     map[string]Extension
		expectedSafeExts map[string]Extension
	}{
		{
			name:   "lfs safe key",
			config: "lfs.fetchinclude\n/foo/bar\x00",
			expectedKeys: map[string][]string{
				"lfs.fetchinclude": []string{"/foo/bar"},
			},
			safeKeys: []string{"lfs.fetchinclude"},
		},
		{
			name:   "lfs safe key (value with newline)",
			config: "lfs.fetchinclude\n/foo\n/bar\x00",
			expectedKeys: map[string][]string{
				"lfs.fetchinclude": []string{"/foo\n/bar"},
			},
			safeKeys: []string{"lfs.fetchinclude"},
		},
		{
			name:   "lfs unsafe key",
			config: "lfs.storage\n/foo/bar\x00",
			expectedKeys: map[string][]string{
				"lfs.storage": []string{"/foo/bar"},
			},
		},
		{
			name:   "lfs remote key",
			config: "remote.origin.lfsurl\nhttps://example.com/foo/bar.git/info/lfs\x00",
			expectedKeys: map[string][]string{
				"remote.origin.lfsurl": []string{"https://example.com/foo/bar.git/info/lfs"},
			},
			safeKeys: []string{"remote.origin.lfsurl"},
		},
		{
			name:   "lfs remote key (multi-part subsection)",
			config: "remote.other.origin.lfsurl\nhttps://example.com/foo/bar.git/info/lfs\x00",
			expectedKeys: map[string][]string{
				"remote.other.origin.lfsurl": []string{"https://example.com/foo/bar.git/info/lfs"},
			},
			safeKeys: []string{"remote.other.origin.lfsurl"},
		},
		{
			name:   "non-lfs remote key",
			config: "remote.origin.pushurl\nhttps://example.com/foo/bar.git\x00",
			expectedKeys: map[string][]string{
				"remote.origin.pushurl": []string{"https://example.com/foo/bar.git"},
			},
		},
		{
			name:   "non-lfs remote key (multi-part subsection)",
			config: "remote.other.origin.pushurl\nhttps://example.com/foo/bar.git\x00",
			expectedKeys: map[string][]string{
				"remote.other.origin.pushurl": []string{"https://example.com/foo/bar.git"},
			},
		},
		{
			name: "non-lfs remote key (multiple values)",
			config: "remote.origin.pushurl\nhttps://example.com/foo/bar.git\x00" +
				"remote.origin.pushurl\nhttps://example.com/foo/baz.git\x00",
			expectedKeys: map[string][]string{
				"remote.origin.pushurl": []string{
					"https://example.com/foo/bar.git",
					"https://example.com/foo/baz.git",
				},
			},
		},
		{
			name:   "non-lfs remote key (empty value)",
			config: "remote.origin.pushurl\n\x00",
			expectedKeys: map[string][]string{
				"remote.origin.pushurl": []string{""},
			},
		},
		{
			name:   "non-lfs remote key (no subsection)",
			config: "remote.origin\nunknown\x00",
			expectedKeys: map[string][]string{
				"remote.origin": []string{"unknown"},
			},
		},
		{
			name:   "lfs remote default key",
			config: "remote.lfsdefault\nupstream\x00",
			expectedKeys: map[string][]string{
				"remote.lfsdefault": []string{"upstream"},
			},
		},
		{
			name:   "non-lfs remote default key",
			config: "remote.pushdefault\nupstream\x00",
			expectedKeys: map[string][]string{
				"remote.pushdefault": []string{"upstream"},
			},
		},
		{
			name:   "lfs access key",
			config: "lfs.https://example.com.access\nbasic\x00",
			expectedKeys: map[string][]string{
				"lfs.https://example.com.access": []string{"basic"},
			},
			safeKeys: []string{"lfs.https://example.com.access"},
		},
		{
			name:   "non-lfs access key",
			config: "other.example.access\nnone\x00",
			expectedKeys: map[string][]string{
				"other.example.access": []string{"none"},
			},
		},
		{
			name: "lfs custom transfer keys",
			config: "lfs.customtransfer.demo.path\n/usr/bin/demo\x00" +
				"lfs.customtransfer.demo.args\n-v --quiet\x00" +
				"lfs.customtransfer.demo.direction\nupload\x00",
			expectedKeys: map[string][]string{
				"lfs.customtransfer.demo.path":      []string{"/usr/bin/demo"},
				"lfs.customtransfer.demo.args":      []string{"-v --quiet"},
				"lfs.customtransfer.demo.direction": []string{"upload"},
			},
		},
		{
			name: "lfs extension keys",
			config: "lfs.extension.demo.clean\next-clean %f\x00" +
				"lfs.extension.demo.smudge\next-smudge %f\x00" +
				"lfs.extension.demo.priority\n1\x00",
			expectedKeys: map[string][]string{
				"lfs.extension.demo.clean":    []string{"ext-clean %f"},
				"lfs.extension.demo.smudge":   []string{"ext-smudge %f"},
				"lfs.extension.demo.priority": []string{"1"},
			},
			safeKeys: []string{"lfs.extension.demo.priority"},
			expectedExts: map[string]Extension{
				"demo": Extension{
					Name:     "demo",
					Clean:    "ext-clean %f",
					Smudge:   "ext-smudge %f",
					Priority: 1,
				},
			},
			expectedSafeExts: map[string]Extension{
				"demo": Extension{
					Name:     "demo",
					Priority: 1,
				},
			},
		},
	}

	assert := func(t *testing.T, config string, onlySafeKeys bool, expectedKeys map[string][]string, expectedExts map[string]Extension) {
		source := git.ParseConfigEntries(config, onlySafeKeys)
		fetcher, exts, _ := readGitConfig(source)

		if expectedKeys == nil {
			assert.Empty(t, fetcher.All())
		} else {
			assert.Equal(t, expectedKeys, fetcher.All())
		}

		if expectedExts == nil {
			assert.Empty(t, exts)
		} else {
			assert.Equal(t, expectedExts, exts)
		}
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert(t, test.config, false, test.expectedKeys, test.expectedExts)
		})

		expectedSafeKeys := make(map[string][]string, len(test.safeKeys))
		for _, key := range test.safeKeys {
			values, ok := test.expectedKeys[key]
			require.True(t, ok, "safe key %q has no expected value", key)
			expectedSafeKeys[key] = values
		}

		t.Run(test.name+" (safe keys only)", func(t *testing.T) {
			assert(t, test.config, true, expectedSafeKeys, test.expectedSafeExts)
		})
	}
}

func TestReadGitConfigValues(t *testing.T) {
	fetcher, _, _ := readGitConfig(&git.ConfigurationSource{
		Entries: []git.ConfigurationEntry{
			{Key: "lfs.fetchinclude", Value: "/foo"},
			{Key: "lfs.fetchinclude", Value: "/bar"},
			{Key: "lfs.locksverify", Value: "false"},
			{Key: "lfs.url", Value: "https://example.com/foo\nlfs.customtransfer.demo.path=injected"},
			{Key: "remote.origin.url", Value: "https://example.com/foo\nlfs.extension.demo.clean=injected"},
		},
	})

	assert.Equal(t, []string{"/foo", "/bar"}, fetcher.GetAll("lfs.fetchinclude"))
	assert.Equal(t, []string{"false"}, fetcher.GetAll("lfs.locksverify"))
	assert.Equal(t, []string{"https://example.com/foo\nlfs.customtransfer.demo.path=injected"}, fetcher.GetAll("lfs.url"))
	assert.Equal(t, []string{"https://example.com/foo\nlfs.extension.demo.clean=injected"}, fetcher.GetAll("remote.origin.url"))

	// Get returns the last value, matching Git's precedence.
	value, ok := fetcher.Get("lfs.fetchinclude")
	assert.True(t, ok)
	assert.Equal(t, "/bar", value)

	value, ok = fetcher.Get("lfs.locksverify")
	assert.True(t, ok)
	assert.Equal(t, "false", value)

	value, ok = fetcher.Get("lfs.url")
	assert.True(t, ok)
	assert.Equal(t, "https://example.com/foo\nlfs.customtransfer.demo.path=injected", value)

	value, ok = fetcher.Get("remote.origin.url")
	assert.True(t, ok)
	assert.Equal(t, "https://example.com/foo\nlfs.extension.demo.clean=injected", value)

	// Get returns false and an empty string for a nonexistent key.
	value, ok = fetcher.Get("lfs.pushurl")
	assert.False(t, ok)
	assert.Empty(t, value)

	// Injected keys should be ignored.
	value, ok = fetcher.Get("lfs.customtransfer.demo.path")
	assert.False(t, ok)
	assert.Empty(t, value)

	value, ok = fetcher.Get("lfs.extension.demo.clean")
	assert.False(t, ok)
	assert.Empty(t, value)
}
