package config

import (
	"testing"

	"github.com/git-lfs/git-lfs/git"
	"github.com/stretchr/testify/assert"
)

func TestRemoteDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.unused.remote":     []string{"a"},
			"branch.unused.pushRemote": []string{"b"},
		},
	})
	assert.Equal(t, "origin", cfg.Remote())
	assert.Equal(t, "origin", cfg.PushRemote())
}

func TestRemoteBranchConfig(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.master.remote":    []string{"a"},
			"branch.other.pushRemote": []string{"b"},
		},
	})
	cfg.ref = &git.Ref{Name: "master"}

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "a", cfg.PushRemote())
}

func TestRemotePushDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.master.remote":    []string{"a"},
			"remote.pushDefault":      []string{"b"},
			"branch.other.pushRemote": []string{"c"},
		},
	})
	cfg.ref = &git.Ref{Name: "master"}

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "b", cfg.PushRemote())
}

func TestRemoteBranchPushDefault(t *testing.T) {
	cfg := NewFrom(Values{
		Git: map[string][]string{
			"branch.master.remote":     []string{"a"},
			"remote.pushDefault":       []string{"b"},
			"branch.master.pushRemote": []string{"c"},
		},
	})
	cfg.ref = &git.Ref{Name: "master"}

	assert.Equal(t, "a", cfg.Remote())
	assert.Equal(t, "c", cfg.PushRemote())
}

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
		Git: map[string][]string{
			"lfs.extension.foo.clean":    []string{"foo-clean %f"},
			"lfs.extension.foo.smudge":   []string{"foo-smudge %f"},
			"lfs.extension.foo.priority": []string{"2"},
		},
	})

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
