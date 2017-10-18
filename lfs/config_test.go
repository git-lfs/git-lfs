package lfs

import (
	"testing"

	"github.com/git-lfs/git-lfs/config"
	"github.com/stretchr/testify/assert"
)

func TestFetchPruneConfigDefault(t *testing.T) {
	cfg := config.NewFrom(config.Values{})
	fp := NewFetchPruneConfig(cfg.Git)

	assert.Equal(t, 7, fp.FetchRecentRefsDays)
	assert.Equal(t, 0, fp.FetchRecentCommitsDays)
	assert.Equal(t, 3, fp.PruneOffsetDays)
	assert.True(t, fp.FetchRecentRefsIncludeRemotes)
	assert.Equal(t, 3, fp.PruneOffsetDays)
	assert.Equal(t, "origin", fp.PruneRemoteName)
	assert.False(t, fp.PruneVerifyRemoteAlways)
}

func TestFetchPruneConfigCustom(t *testing.T) {
	cfg := config.NewFrom(config.Values{
		Git: map[string][]string{
			"lfs.fetchrecentrefsdays":     []string{"12"},
			"lfs.fetchrecentremoterefs":   []string{"false"},
			"lfs.fetchrecentcommitsdays":  []string{"9"},
			"lfs.pruneoffsetdays":         []string{"30"},
			"lfs.pruneverifyremotealways": []string{"true"},
			"lfs.pruneremotetocheck":      []string{"upstream"},
		},
	})
	fp := NewFetchPruneConfig(cfg.Git)

	assert.Equal(t, 12, fp.FetchRecentRefsDays)
	assert.Equal(t, 9, fp.FetchRecentCommitsDays)
	assert.False(t, fp.FetchRecentRefsIncludeRemotes)
	assert.Equal(t, 30, fp.PruneOffsetDays)
	assert.Equal(t, "upstream", fp.PruneRemoteName)
	assert.True(t, fp.PruneVerifyRemoteAlways)
}
