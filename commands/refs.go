package commands

import (
	"fmt"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/git"
	"github.com/rubyist/tracerx"
)

type refUpdate struct {
	git    config.Environment
	remote string
	left   *git.Ref
	right  *git.Ref
}

func newRefUpdate(g config.Environment, remote string, l, r *git.Ref) *refUpdate {
	return &refUpdate{
		git:    g,
		remote: remote,
		left:   l,
		right:  r,
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
		u.right = defaultRemoteRef(u.git, u.remote, u.Left())
	}
	return u.right
}

// defaultRemoteRef returns the remote ref receiving a push based on the current
// repository config and local ref being pushed.
//
// See push.default rules in https://git-scm.com/docs/git-config
func defaultRemoteRef(g config.Environment, remote string, left *git.Ref) *git.Ref {
	pushMode, _ := g.Get("push.default")
	switch pushMode {
	case "", "simple":
		brRemote, _ := g.Get(fmt.Sprintf("branch.%s.remote", left.Name))
		if brRemote == remote {
			// in centralized workflow, work like 'upstream' with an added safety to
			// refuse to push if the upstream branch’s name is different from the
			// local one.
			return trackingRef(g, left)
		}

		// When pushing to a remote that is different from the remote you normally
		// pull from, work as current.
		return &git.Ref{Name: left.Name}
	case "upstream", "tracking":
		// push the current branch back to the branch whose changes are usually
		// integrated into the current branch
		return trackingRef(g, left)
	case "current":
		// push the current branch to update a branch with the same name on the
		// receiving end.
		return &git.Ref{Name: left.Name}
	default:
		tracerx.Printf("WARNING: %q push mode not supported", pushMode)
		return &git.Ref{Name: left.Name}
	}
}

func trackingRef(g config.Environment, left *git.Ref) *git.Ref {
	if merge, ok := g.Get(fmt.Sprintf("branch.%s.merge", left.Name)); ok {
		return git.ParseRef(merge, "")
	}
	return &git.Ref{Name: left.Name}
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
