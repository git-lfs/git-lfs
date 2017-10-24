package lfs

import (
	"github.com/git-lfs/git-lfs/config"
)

type GitFilter struct {
	cfg *config.Configuration
}

func NewGitFilter(cfg *config.Configuration) *GitFilter {
	return &GitFilter{cfg: cfg}
}
