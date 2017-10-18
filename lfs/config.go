package lfs

import "github.com/git-lfs/git-lfs/config"

// FetchPruneConfig collects together the config options that control fetching and pruning
type FetchPruneConfig struct {
	// The number of days prior to current date for which (local) refs other than HEAD
	// will be fetched with --recent (default 7, 0 = only fetch HEAD)
	FetchRecentRefsDays int
	// Makes the FetchRecentRefsDays option apply to remote refs from fetch source as well (default true)
	FetchRecentRefsIncludeRemotes bool
	// number of days prior to latest commit on a ref that we'll fetch previous
	// LFS changes too (default 0 = only fetch at ref)
	FetchRecentCommitsDays int
	// Whether to always fetch recent even without --recent
	FetchRecentAlways bool
	// Number of days added to FetchRecent*; data outside combined window will be
	// deleted when prune is run. (default 3)
	PruneOffsetDays int
	// Always verify with remote before pruning
	PruneVerifyRemoteAlways bool
	// Name of remote to check for unpushed and verify checks
	PruneRemoteName string
}

func NewFetchPruneConfig(git config.Environment) FetchPruneConfig {
	pruneRemote, _ := git.Get("lfs.pruneremotetocheck")
	if len(pruneRemote) == 0 {
		pruneRemote = "origin"
	}

	return FetchPruneConfig{
		FetchRecentRefsDays:           git.Int("lfs.fetchrecentrefsdays", 7),
		FetchRecentRefsIncludeRemotes: git.Bool("lfs.fetchrecentremoterefs", true),
		FetchRecentCommitsDays:        git.Int("lfs.fetchrecentcommitsdays", 0),
		FetchRecentAlways:             git.Bool("lfs.fetchrecentalways", false),
		PruneOffsetDays:               git.Int("lfs.pruneoffsetdays", 3),
		PruneVerifyRemoteAlways:       git.Bool("lfs.pruneverifyremotealways", false),
		PruneRemoteName:               pruneRemote,
	}
}
