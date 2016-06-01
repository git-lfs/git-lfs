package tools_test

import (
	"testing"

	"github.com/github/git-lfs/tools"
	"github.com/stretchr/testify/assert"
)

func TestCleanPathsCleansPaths(t *testing.T) {
	cleaned := tools.CleanPaths("/foo/bar/,/foo/bar/baz", ",")

	assert.Equal(t, []string{"/foo/bar", "/foo/bar/baz"}, cleaned)
}

func TestCleanPathsReturnsNoResultsWhenGivenNoPaths(t *testing.T) {
	cleaned := tools.CleanPaths("", ",")

	assert.Empty(t, cleaned)
}

func TestCleanPathsDefaultReturnsInputWhenResultsPresent(t *testing.T) {
	cleaned := tools.CleanPathsDefault("/foo/bar/", ",", []string{"/default"})

	assert.Equal(t, []string{"/foo/bar"}, cleaned)
}

func TestCleanPathsDefaultReturnsDefaultWhenResultsAbsent(t *testing.T) {
	cleaned := tools.CleanPathsDefault("", ",", []string{"/default"})

	assert.Equal(t, []string{"/default"}, cleaned)
}
