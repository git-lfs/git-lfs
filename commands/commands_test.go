package commands

import (
	"testing"

	"github.com/github/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

var (
	testcfg = config.NewFrom(config.Values{
		Git: map[string]string{
			"lfs.fetchinclude": "/default/include",
			"lfs.fetchexclude": "/default/exclude",
		},
	})
)

func TestDetermineIncludeExcludePathsReturnsCleanedPaths(t *testing.T) {
	inc := "/some/include"
	exc := "/some/exclude"
	i, e := determineIncludeExcludePaths(testcfg, &inc, &exc)

	assert.Equal(t, []string{"/some/include"}, i)
	assert.Equal(t, []string{"/some/exclude"}, e)
}

func TestDetermineIncludeExcludePathsReturnsEmptyPaths(t *testing.T) {
	inc := ""
	exc := ""
	i, e := determineIncludeExcludePaths(testcfg, &inc, &exc)

	assert.Empty(t, i)
	assert.Empty(t, e)
}

func TestDetermineIncludeExcludePathsReturnsDefaultsWhenAbsent(t *testing.T) {
	i, e := determineIncludeExcludePaths(testcfg, nil, nil)

	assert.Equal(t, []string{"/default/include"}, i)
	assert.Equal(t, []string{"/default/exclude"}, e)
}

func TestCommandEnabledFromEnvironmentVariables(t *testing.T) {
	cfg := config.New()
	err := cfg.Setenv("GITLFSLOCKSENABLED", "1")

	assert.Nil(t, err)
	assert.True(t, isCommandEnabled(cfg, "locks"))
}

func TestCommandEnabledDisabledByDefault(t *testing.T) {
	cfg := config.New()

	// Since config.Configuration.Setenv makes a call to os.Setenv, we have
	// to make sure that the LFSLOCKSENABLED enviornment variable is not
	// present in the configuration object during the lifecycle of this
	// test.
	//
	// This behavior can cause race conditions with the above test when
	// running in parallel, so this should be investigated further in the
	// future.
	err := cfg.Setenv("GITLFSLOCKSENABLED", "")

	assert.Nil(t, err)
	assert.False(t, isCommandEnabled(cfg, "locks"))
}
