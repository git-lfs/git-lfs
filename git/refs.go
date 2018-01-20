package git

import (
	"fmt"

	"github.com/rubyist/tracerx"
)

type RefUpdate struct {
	git    Env
	remote string
	left   *Ref
	right  *Ref
}

func NewRefUpdate(g Env, remote string, l, r *Ref) *RefUpdate {
	return &RefUpdate{
		git:    g,
		remote: remote,
		left:   l,
		right:  r,
	}
}

func (u *RefUpdate) Left() *Ref {
	return u.left
}

func (u *RefUpdate) LeftCommitish() string {
	return refCommitish(u.Left())
}

func (u *RefUpdate) Right() *Ref {
	if u.right == nil {
		u.right = defaultRemoteRef(u.git, u.remote, u.Left())
	}
	return u.right
}

// defaultRemoteRef returns the remote ref receiving a push based on the current
// repository config and local ref being pushed.
//
// See push.default rules in https://git-scm.com/docs/git-config
func defaultRemoteRef(g Env, remote string, left *Ref) *Ref {
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
		return left
	case "upstream", "tracking":
		// push the current branch back to the branch whose changes are usually
		// integrated into the current branch
		return trackingRef(g, left)
	case "current":
		// push the current branch to update a branch with the same name on the
		// receiving end.
		return left
	default:
		tracerx.Printf("WARNING: %q push mode not supported", pushMode)
		return left
	}
}

func trackingRef(g Env, left *Ref) *Ref {
	if merge, ok := g.Get(fmt.Sprintf("branch.%s.merge", left.Name)); ok {
		return ParseRef(merge, "")
	}
	return left
}

func (u *RefUpdate) RightCommitish() string {
	return refCommitish(u.Right())
}

func refCommitish(r *Ref) string {
	if len(r.Sha) > 0 {
		return r.Sha
	}
	return r.Name
}

// copy of env
type Env interface {
	Get(key string) (val string, ok bool)
}
