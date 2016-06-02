package commands

import (
	"testing"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

var (
	cfg = lfs.NewFromValues(map[string]string{
		"lfs.fetchinclude": "/default/include",
		"lfs.fetchexclude": "/default/exclude",
	})
)

func TestDetermineIncludeExcludePathsReturnsCleanedPaths(t *testing.T) {
	i, e := determineIncludeExcludePaths(cfg, "/some/include", "/some/exclude")

	assert.Equal(t, []string{"/some/include"}, i)
	assert.Equal(t, []string{"/some/exclude"}, e)
}

func TestDetermineIncludeExcludePathsReturnsDefaultsWhenAbsent(t *testing.T) {
	i, e := determineIncludeExcludePaths(cfg, "", "")

	assert.Equal(t, []string{"/default/include"}, i)
	assert.Equal(t, []string{"/default/exclude"}, e)
}
