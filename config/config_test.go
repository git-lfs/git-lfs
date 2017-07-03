package config

import (
	"testing"
	"time"

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
		Git: map[string][]string{
			"lfs.fetchrecentrefsdays":     []string{"12"},
			"lfs.fetchrecentremoterefs":   []string{"false"},
			"lfs.fetchrecentcommitsdays":  []string{"9"},
			"lfs.pruneoffsetdays":         []string{"30"},
			"lfs.pruneverifyremotealways": []string{"true"},
			"lfs.pruneremotetocheck":      []string{"upstream"},
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
		Git: map[string][]string{
			"lfs.fetchinclude": []string{"/path/to/clean/"},
			"lfs.fetchexclude": []string{"/other/path/to/clean/"},
		},
	})

	assert.Equal(t, []string{"/path/to/clean"}, cfg.FetchIncludePaths())
	assert.Equal(t, []string{"/other/path/to/clean"}, cfg.FetchExcludePaths())
}

func TestUnmarshalMultipleTypes(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"string": []string{"string"},
			"int":    []string{"1"},
			"bool":   []string{"true"},
		},
		Os: map[string][]string{
			"string": []string{"string"},
			"int":    []string{"1"},
			"bool":   []string{"true"},
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
		Git: map[string][]string{
			"string": []string{"bar"},
			"int":    []string{"2"},
			"bool":   []string{"false"},
		},
	})

	err := cfg.Unmarshal(v)

	assert.Nil(t, err)
	assert.Equal(t, "bar", v.String)
	assert.Equal(t, 2, v.Int)
	assert.Equal(t, false, v.Bool)
}

func TestUnmarshalAllowsBothOsAndGitTags(t *testing.T) {
	v := &struct {
		String string `git:"string" os:"STRING"`
	}{}

	cfg := NewFrom(Values{
		Git: map[string][]string{"string": []string{"foo"}},
		Os:  map[string][]string{"STRING": []string{"bar"}},
	})

	err := cfg.Unmarshal(v)

	assert.Nil(t, err)
	assert.Equal(t, "foo", v.String)
}

func TestUnmarshalYieldsToDefaultIfBothEnvsMissing(t *testing.T) {
	v := &struct {
		String string `git:"string" os:"STRING"`
	}{"foo"}

	cfg := NewFrom(Values{})

	err := cfg.Unmarshal(v)

	assert.Nil(t, err)
	assert.Equal(t, "foo", v.String)
}

func TestUnmarshalOverridesDefaultIfAnyEnvPresent(t *testing.T) {
	v := &struct {
		String string `git:"string" os:"STRING"`
	}{"foo"}

	cfg := NewFrom(Values{
		Git: map[string][]string{"string": []string{"bar"}},
		Os:  map[string][]string{"STRING": []string{"baz"}},
	})

	err := cfg.Unmarshal(v)

	assert.Nil(t, err)
	assert.Equal(t, "bar", v.String)
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
		Git: map[string][]string{"duration": []string{"foo"}},
	})

	err := cfg.Unmarshal(v)

	assert.Equal(t, "lfs/config: unsupported target type for field \"Unsupported\": time.Duration", err.Error())
}
