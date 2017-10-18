package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicTransfersOnlySetValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.basictransfersonly": []string{"true"},
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
		Git: map[string][]string{
			"lfs.basictransfersonly": []string{"wat"},
		},
	})

	b := cfg.BasicTransfersOnly()
	assert.Equal(t, false, b)
}

func TestTusTransfersAllowedSetValue(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.tustransfers": []string{"true"},
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
		Git: map[string][]string{
			"lfs.tustransfers": []string{"wat"},
		},
	})

	b := cfg.TusTransfersAllowed()
	assert.Equal(t, false, b)
}

func TestLoadValidExtension(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{},
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

func TestFetchIncludeExcludesAreCleaned(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"lfs.fetchinclude": []string{"/path/to/clean/"},
			"lfs.fetchexclude": []string{"/other/path/to/clean/"},
		},
	})

	assert.Equal(t, []string{"/path/to/clean"}, cfg.FetchIncludePaths())
	assert.Equal(t, []string{"/other/path/to/clean"}, cfg.FetchExcludePaths())
}
