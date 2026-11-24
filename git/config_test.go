package git_test // to avoid import cycles

import (
	"testing"

	. "github.com/git-lfs/git-lfs/v3/git"
	"github.com/stretchr/testify/assert"
)

func TestReadOnlyConfig(t *testing.T) {
	cfg := NewReadOnlyConfig("", "")
	_, err := cfg.SetLocal("lfs.this.should", "fail")
	assert.Equal(t, err, ErrReadOnly)
}

func TestParseConfigEntries(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		expected []ConfigurationEntry
	}{
		{
			name:     "no entries",
			config:   "",
			expected: []ConfigurationEntry{},
		},
		{
			name:     "key-only entry",
			config:   "lfs.flag\x00",
			expected: []ConfigurationEntry{},
		},
		{
			name:   "single entry",
			config: "lfs.url\nhttps://example.com/repo\x00",
			expected: []ConfigurationEntry{
				{Key: "lfs.url", Value: "https://example.com/repo"},
			},
		},
		{
			name:   "multiple entries",
			config: "lfs.url\nhttps://example.com/repo\x00user.name\nPat Doe\x00",
			expected: []ConfigurationEntry{
				{Key: "lfs.url", Value: "https://example.com/repo"},
				{Key: "user.name", Value: "Pat Doe"},
			},
		},
		{
			name:   "empty value",
			config: "lfs.url\n\x00",
			expected: []ConfigurationEntry{
				{Key: "lfs.url"},
			},
		},
		{
			name:   "value with whitespace",
			config: "lfs.url\n  spaced value  \x00",
			expected: []ConfigurationEntry{
				{Key: "lfs.url", Value: "  spaced value  "},
			},
		},
		{
			name:   "value with equals sign",
			config: "lfs.customtransfer.demo.args\n-c 1 --url=https://example.com\x00",
			expected: []ConfigurationEntry{
				{Key: "lfs.customtransfer.demo.args", Value: "-c 1 --url=https://example.com"},
			},
		},
		{
			name:   "value with newline",
			config: "lfs.url\nfirst\nsecond\x00",
			expected: []ConfigurationEntry{
				{Key: "lfs.url", Value: "first\nsecond"},
			},
		},
		{
			name:   "value with newline and equals sign",
			config: "lfs.url\nfirst\nsecond=third\x00",
			expected: []ConfigurationEntry{
				{Key: "lfs.url", Value: "first\nsecond=third"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := ParseConfigEntries(test.config, false)
			assert.Equal(t, test.expected, source.Entries)
			assert.False(t, source.OnlySafeKeys)
		})
	}
}

func TestParseConfigEntriesPreservesOnlySafeKeys(t *testing.T) {
	source := ParseConfigEntries("lfs.url\nhttps://example.com/repo\x00", true)

	assert.Equal(t, []ConfigurationEntry{
		{Key: "lfs.url", Value: "https://example.com/repo"},
	}, source.Entries)

	// OnlySafeKeys should retain the value of the onlySafeKeys parameter.
	assert.True(t, source.OnlySafeKeys)
}
