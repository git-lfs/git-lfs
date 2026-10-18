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

func TestParseConfigLines(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		expected []string
	}{
		{
			name:     "no entries",
			config:   "",
			expected: []string{""},
		},
		{
			name:     "key-only entry",
			config:   "lfs.flag",
			expected: []string{"lfs.flag"},
		},
		{
			name:   "single entry",
			config: "lfs.url=https://example.com/repo",
			expected: []string{
				"lfs.url=https://example.com/repo",
			},
		},
		{
			name:   "multiple entries",
			config: "lfs.url=https://example.com/repo\nuser.name=Pat Doe",
			expected: []string{
				"lfs.url=https://example.com/repo",
				"user.name=Pat Doe",
			},
		},
		{
			name:   "empty value",
			config: "lfs.url=",
			expected: []string{
				"lfs.url=",
			},
		},
		{
			name:   "value with whitespace",
			config: "lfs.url=  spaced value  ",
			expected: []string{
				"lfs.url=  spaced value  ",
			},
		},
		{
			name:   "value with equals sign",
			config: "lfs.customtransfer.demo.args=-c 1 --url=https://example.com",
			expected: []string{
				"lfs.customtransfer.demo.args=-c 1 --url=https://example.com",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := ParseConfigLines(test.config, false)

			assert.Equal(t, test.expected, source.Lines)
			assert.False(t, source.OnlySafeKeys)
		})
	}
}

func TestParseConfigLinesPreservesSafeKeysFlag(t *testing.T) {
	source := ParseConfigLines("lfs.url=https://example.com/repo", true)

	assert.Equal(t, []string{
		"lfs.url=https://example.com/repo",
	}, source.Lines)

	// OnlySafeKeys should retain the value of the onlySafeKeys parameter.
	assert.True(t, source.OnlySafeKeys)
}
