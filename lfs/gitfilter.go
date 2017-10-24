package lfs

import (
	"github.com/git-lfs/git-lfs/config"
)

// GitFilter provides clean and smudge capabilities
type GitFilter struct {
	cfg *config.Configuration
}

// NewGitFilter initializes a new *GitFilter
func NewGitFilter(cfg *config.Configuration) *GitFilter {
	return &GitFilter{cfg: cfg}
}
