package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/lfs"
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

// setupAutoTrackTest creates a temporary directory with a .gitattributes file
// and a .git/lfs/tmp directory (needed by cfg.TempDir()), then replaces the
// global cfg with a config pointing to that directory. The returned function
// restores the original cfg.
func setupAutoTrackTest(t *testing.T, gitattributesContent string) func() {
	t.Helper()
	dir := t.TempDir()

	err := os.MkdirAll(filepath.Join(dir, ".git", "lfs", "tmp"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	if gitattributesContent != "" {
		err = os.WriteFile(filepath.Join(dir, ".gitattributes"), []byte(gitattributesContent), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	saved := cfg
	cfg = config.NewIn(dir, "")
	return func() { cfg = saved }
}

// --- Plan point 6: Memory concern with io.ReadAll — spooling in clean() ---

func TestCleanAutoTrackPassThrough(t *testing.T) {
	defer setupAutoTrackTest(t, "* autotracksize=100")()

	gf := lfs.NewGitFilter(cfg)
	var buf bytes.Buffer
	data := strings.NewReader("small file content")

	ptr, err := clean(gf, &buf, data, "test.dat", -1)
	assert.Nil(t, ptr)
	assert.NoError(t, err)
	assert.Equal(t, "small file content", buf.String())
}

func TestCleanAutoTrackGitCriticalExcluded(t *testing.T) {
	defer setupAutoTrackTest(t, "* autotracksize=10")()

	gf := lfs.NewGitFilter(cfg)
	var buf bytes.Buffer
	data := strings.NewReader("this is a much longer file content that exceeds")

	ptr, err := clean(gf, &buf, data, ".gitattributes", -1)
	assert.Nil(t, ptr)
	assert.NoError(t, err)
	assert.Equal(t, "this is a much longer file content that exceeds", buf.String())
}

func TestCleanAutoTrackExcludedFile(t *testing.T) {
	defer setupAutoTrackTest(t, "* autotracksize=10")()

	gf := lfs.NewGitFilter(cfg)
	var buf bytes.Buffer
	data := strings.NewReader("this is a much longer file content that exceeds")

	ptr, err := clean(gf, &buf, data, "readme.md", -1)
	assert.Nil(t, ptr)
	assert.NoError(t, err)
	assert.Equal(t, "this is a much longer file content that exceeds", buf.String())
}

func TestCleanAutoTrackAlreadyPointer(t *testing.T) {
	defer setupAutoTrackTest(t, "* autotracksize=100")()

	gf := lfs.NewGitFilter(cfg)
	var buf bytes.Buffer
	pointerContent := "version https://git-lfs.github.com/spec/v1\noid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5\nsize 1024\n"
	data := strings.NewReader(pointerContent)

	ptr, err := clean(gf, &buf, data, "test.dat", -1)
	assert.NotNil(t, ptr)
	assert.NoError(t, err)
	assert.Equal(t, pointerContent, buf.String())
}

// --- Plan point 7: Smudge pass-through check ---

func TestSmudgePassThroughAutoTrack(t *testing.T) {
	defer setupAutoTrackTest(t, "* autotracksize=100")()

	// File under threshold passes through
	assert.True(t, smudgePassThrough("test.dat", 50))

	// File over threshold does not pass through
	assert.False(t, smudgePassThrough("test.dat", 200))

	// Excluded file passes through regardless of size
	assert.True(t, smudgePassThrough("readme.md", 200))

	// Git-critical file passes through regardless of size
	assert.True(t, smudgePassThrough(".gitattributes", 200))
}

// --- Plan point 8: fileSize = -1 (PR #6011 scenario) ---

func TestCleanAutoTrackFileSizeNegativeOne(t *testing.T) {
	defer setupAutoTrackTest(t, "* autotracksize=100")()

	gf := lfs.NewGitFilter(cfg)
	var buf bytes.Buffer
	data := strings.NewReader("small content")

	// fileSize=-1 simulates behavior after PR #6011; the autotrack
	// pass-through must not depend on this parameter.
	ptr, err := clean(gf, &buf, data, "test.dat", -1)
	assert.Nil(t, ptr)
	assert.NoError(t, err)
	assert.Equal(t, "small content", buf.String())
}

func TestCleanAutoTrackFileSizeNegativeOneOverThreshold(t *testing.T) {
	defer setupAutoTrackTest(t, "* autotracksize=10")()

	gf := lfs.NewGitFilter(cfg)
	var buf bytes.Buffer
	data := strings.NewReader("this is a much longer file content that exceeds")

	ptr, err := clean(gf, &buf, data, "nontracked.dat", -1)
	// With autotrack active, file over threshold should NOT pass through
	// as (nil, nil). It hits gf.Clean which needs a real git repo, so
	// we expect some error (ExitWithError or similar), but the key is
	// that it did NOT silently pass through via the old fileSize >= 0 gating.
	if ptr == nil && err == nil {
		t.Errorf("file over threshold with autotrack should not return (nil, nil); must reach gf.Clean")
	}
}
