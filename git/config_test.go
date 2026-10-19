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
			config:   "lfs.flag",
			expected: []ConfigurationEntry{},
		},
		{
			name:   "single entry",
			config: "lfs.url=https://example.com/repo",
			expected: []ConfigurationEntry{
				{Key: "lfs.url", Value: "https://example.com/repo"},
			},
		},
		{
			name:   "multiple entries",
			config: "lfs.url=https://example.com/repo\nuser.name=Pat Doe",
			expected: []ConfigurationEntry{
				{Key: "lfs.url", Value: "https://example.com/repo"},
				{Key: "user.name", Value: "Pat Doe"},
			},
		},
		{
			name:   "empty value",
			config: "lfs.url=",
			expected: []ConfigurationEntry{
				{Key: "lfs.url"},
			},
		},
		{
			name:   "value with whitespace",
			config: "lfs.url=  spaced value  ",
			expected: []ConfigurationEntry{
				{Key: "lfs.url", Value: "  spaced value  "},
			},
		},
		{
			name:   "value with equals sign",
			config: "lfs.customtransfer.demo.args=-c 1 --url=https://example.com",
			expected: []ConfigurationEntry{
				{Key: "lfs.customtransfer.demo.args", Value: "-c 1 --url=https://example.com"},
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
	source := ParseConfigEntries("lfs.url=https://example.com/repo", true)

	assert.Equal(t, []ConfigurationEntry{
		{Key: "lfs.url", Value: "https://example.com/repo"},
	}, source.Entries)

	// OnlySafeKeys should retain the value of the onlySafeKeys parameter.
	assert.True(t, source.OnlySafeKeys)
}
