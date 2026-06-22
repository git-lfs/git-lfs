package commands

import (
	"testing"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/stretchr/testify/assert"
)

var (
	testcfg = config.NewFrom(config.Values{
		Git: map[string][]string{
			"lfs.fetchinclude": []string{"/default/include"},
			"lfs.fetchexclude": []string{"/default/exclude"},
		},
	})
)

func TestDetermineIncludeExcludePathsReturnsCleanedPaths(t *testing.T) {
	inc := "/some/include"
	exc := "/some/exclude"
	i, e := determineIncludeExcludePaths(testcfg, &inc, &exc, true)

	assert.Equal(t, []string{"/some/include"}, i)
	assert.Equal(t, []string{"/some/exclude"}, e)
}

func TestDetermineIncludeExcludePathsReturnsEmptyPaths(t *testing.T) {
	inc := ""
	exc := ""
	i, e := determineIncludeExcludePaths(testcfg, &inc, &exc, true)

	assert.Empty(t, i)
	assert.Empty(t, e)
}

func TestDetermineIncludeExcludePathsReturnsDefaultsWhenAbsent(t *testing.T) {
	i, e := determineIncludeExcludePaths(testcfg, nil, nil, true)

	assert.Equal(t, []string{"/default/include"}, i)
	assert.Equal(t, []string{"/default/exclude"}, e)
}

func TestDetermineIncludeExcludePathsReturnsNothingWhenAbsent(t *testing.T) {
	i, e := determineIncludeExcludePaths(testcfg, nil, nil, false)

	assert.Empty(t, i)
	assert.Empty(t, e)
}

func TestSpecialGitRefsExclusion(t *testing.T) {
	assert.True(t, isSpecialGitRef("refs/notes/commits"))
	assert.True(t, isSpecialGitRef("refs/bisect/bad"))
	assert.True(t, isSpecialGitRef("refs/replace/abcdef90"))
	assert.True(t, isSpecialGitRef("refs/stash"))
	assert.False(t, isSpecialGitRef("refs/commits/abcdef90"))
}

func TestGitCriticalExcluded(t *testing.T) {
	// Files that must never become LFS pointers
	critical := []string{
		".gitattributes",
		".gitignore",
		".gitmodules",
		".mailmap",
		".gitkeep",
	}
	notCritical := []string{
		"normal.dat",
		"readme.md",
		".gitlab-ci.yml",
		"",
		".github/workflows/ci.yml",
	}

	for _, name := range critical {
		assert.True(t, gitCriticalExcluded(name), "%q must be git-critical", name)
		assert.True(t, gitCriticalExcluded("path/to/"+name), "%q must be git-critical in subdir", name)
	}
	for _, name := range notCritical {
		assert.False(t, gitCriticalExcluded(name), "%q must not be git-critical", name)
	}
}

func testCfg() *config.Configuration {
	return config.NewFrom(config.Values{
		Git: map[string][]string{},
	})
}

func setTestCfg() (restore func()) {
	saved := cfg
	cfg = testCfg()
	return func() { cfg = saved }
}

func TestIsAutoTrackExcludedBuiltinDefaults(t *testing.T) {
	defer setTestCfg()()

	// Patterns that match by basename (work in any directory)
	basenamePatterns := []string{
		".gitlab-ci.yml",
		"readme.md",
		"CHANGELOG.md",
		"notes.txt",
		"output.txt",
		"settings.cfg",
		"config.cfg",
		"setup.ini",
		"config.ini",
	}
	// .github/* matches the full path at root level only
	assert.True(t, isAutoTrackExcluded(".github/workflows"))
	assert.True(t, isAutoTrackExcluded(".github/ci.yml"))

	notExcluded := []string{
		"normal.dat",
		"file.txt.bak",
		"readme",
		"ini",
		".gitattributes",
		".gitignore",
		"",
	}

	for _, name := range basenamePatterns {
		assert.True(t, isAutoTrackExcluded(name), "default: %q must be excluded", name)
		assert.True(t, isAutoTrackExcluded("subdir/"+name), "default: %q must be excluded in subdir", name)
	}
	for _, name := range notExcluded {
		assert.False(t, isAutoTrackExcluded(name), "default: %q must not be excluded", name)
	}
}

func TestIsAutoTrackExcludedGithubDir(t *testing.T) {
	defer setTestCfg()()

	assert.True(t, isAutoTrackExcluded(".github/workflows"))
	assert.True(t, isAutoTrackExcluded(".github/ci.yml"))
	// One level deep matches; deeper paths don't via .github/* glob
	assert.False(t, isAutoTrackExcluded(".github/workflows/ci.yml"))
	assert.False(t, isAutoTrackExcluded("my.github/normal.dat"))
	assert.False(t, isAutoTrackExcluded(".githubfile"))
	assert.False(t, isAutoTrackExcluded("github"))
}

func TestIsAutoTrackExcludedUserConfig(t *testing.T) {
	saved := cfg
	cfg = config.NewFrom(config.Values{
		Git: map[string][]string{
			"lfs.autotrackexclude": {"*.pdf *.log"},
		},
	})
	defer func() { cfg = saved }()

	// User patterns replace defaults
	assert.True(t, isAutoTrackExcluded("document.pdf"))
	assert.True(t, isAutoTrackExcluded("trace.log"))
	assert.True(t, isAutoTrackExcluded("path/to/document.pdf"))
	assert.True(t, isAutoTrackExcluded("path/to/trace.log"))

	// Defaults no longer apply (replaced, not additive)
	assert.False(t, isAutoTrackExcluded("readme.md"), "default *.md should not apply when user overrides")
	assert.False(t, isAutoTrackExcluded("notes.txt"), "default *.txt should not apply when user overrides")

	// .github/ is NOT excluded when user overrides (replaced defaults)
	assert.False(t, isAutoTrackExcluded(".github/workflows"))

	// Non-excluded files pass through
	assert.False(t, isAutoTrackExcluded("normal.dat"))
}

func TestIsAutoTrackExcludedUserConfigEmpty(t *testing.T) {
	saved := cfg
	cfg = config.NewFrom(config.Values{
		Git: map[string][]string{
			"lfs.autotrackexclude": {""},
		},
	})
	defer func() { cfg = saved }()

	// Empty means no user patterns → use defaults
	assert.True(t, isAutoTrackExcluded("readme.md"))
	assert.True(t, isAutoTrackExcluded(".gitlab-ci.yml"))
	assert.False(t, isAutoTrackExcluded("normal.dat"))
}
