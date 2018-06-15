package lfs

import (
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/fs"
	"github.com/git-lfs/git-lfs/git"
)

// GitFilter provides clean and smudge capabilities
type GitFilter struct {
	cfg *config.Configuration
	fs  *fs.Filesystem
}

// NewGitFilter initializes a new *GitFilter
func NewGitFilter(cfg *config.Configuration) *GitFilter {
	return &GitFilter{cfg: cfg, fs: cfg.Filesystem()}
}

func (f *GitFilter) ObjectPath(oid string) (string, error) {
	return f.fs.ObjectPath(oid)
}

func (f *GitFilter) RemoteRef() *git.Ref {
	return git.NewRefUpdate(f.cfg.Git, f.cfg.PushRemote(), f.cfg.CurrentRef(), nil).Right()
}
