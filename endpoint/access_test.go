package endpoint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		cfg := NewConfig(gitEnv(map[string]string{
			"lfs.url":                        "http://example.com",
			"lfs.http://example.com.access":  value,
			"lfs.https://example.com.access": "bad",
		}))

		dl := cfg.Endpoint("upload", "")
		ul := cfg.Endpoint("download", "")

		if access := cfg.AccessFor(dl); access != Access(expected.Access) {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}
		if access := cfg.AccessFor(ul); access != Access(expected.Access) {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}
	}

	// Test again but with separate push url
	for value, expected := range tests {
		cfg := NewConfig(gitEnv(map[string]string{
			"lfs.url":                           "http://example.com",
			"lfs.pushurl":                       "http://examplepush.com",
			"lfs.http://example.com.access":     value,
			"lfs.http://examplepush.com.access": value,
			"lfs.https://example.com.access":    "bad",
		}))

		dl := cfg.Endpoint("upload", "")
		ul := cfg.Endpoint("download", "")

		if access := cfg.AccessFor(dl); access != Access(expected.Access) {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}
		if access := cfg.AccessFor(ul); access != Access(expected.Access) {
			t.Errorf("Expected Access() with value %q to be %v, got %v", value, expected.Access, access)
		}
	}
}

func TestAccessAbsentConfig(t *testing.T) {
	cfg := NewConfig(nil)
	assert.Equal(t, NoneAccess, cfg.AccessFor(cfg.Endpoint("download", "")))
	assert.Equal(t, NoneAccess, cfg.AccessFor(cfg.Endpoint("upload", "")))
}
