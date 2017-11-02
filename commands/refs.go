package commands

import (
	"fmt"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
)

type refUpdate struct {
	git   config.Environment
	left  *git.Ref
	right *git.Ref
}

func newRefUpdate(g config.Environment, l, r *git.Ref) *refUpdate {
	return &refUpdate{
		git:   g,
		left:  l,
		right: r,
	}
}

func (u *refUpdate) Left() *git.Ref {
	return u.left
}

func (u *refUpdate) LeftCommitish() string {
	return refCommitish(u.Left())
}

func (u *refUpdate) Right() *git.Ref {
	if u.right == nil {
		l := u.Left()
		if merge, ok := u.git.Get(fmt.Sprintf("branch.%s.merge", l.Name)); ok {
			u.right = git.ParseRef(merge, "")
		} else {
			u.right = &git.Ref{Name: l.Name}
		}
	}
	return u.right
}

func (u *refUpdate) RightCommitish() string {
	return refCommitish(u.Right())
}

func refCommitish(r *git.Ref) string {
	if len(r.Sha) > 0 {
		return r.Sha
	}
	return r.Name
}
