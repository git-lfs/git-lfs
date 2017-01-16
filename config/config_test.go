package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
