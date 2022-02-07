package git

import (
	"fmt"

	"github.com/rubyist/tracerx"
)

type RefUpdate struct {
	git       Env
	remote    string
	localRef  *Ref
	remoteRef *Ref
}

func NewRefUpdate(g Env, remote string, localRef, remoteRef *Ref) *RefUpdate {
	return &RefUpdate{
		git:       g,
		remote:    remote,
		localRef:  localRef,
		remoteRef: remoteRef,
	}
}

func (u *RefUpdate) LocalRef() *Ref {
	return u.localRef
}

func (u *RefUpdate) LocalRefCommitish() string {
	return refCommitish(u.LocalRef())
}

func (u *RefUpdate) RemoteRef() *Ref {
	if u.remoteRef == nil {
		u.remoteRef = defaultRemoteRef(u.git, u.remote, u.LocalRef())
	}
	return u.remoteRef
}

// defaultRemoteRef returns the remote ref receiving a push based on the current
// repository config and local ref being pushed.
//
// See push.default rules in https://git-scm.com/docs/git-config
func defaultRemoteRef(g Env, remote string, localRef *Ref) *Ref {
	pushMode, _ := g.Get("push.default")
	switch pushMode {
	case "", "simple":
		brRemote, _ := g.Get(fmt.Sprintf("branch.%s.remote", localRef.Name))
		if brRemote == remote {
			// in centralized workflow, work like 'upstream' with an added safety to
			// refuse to push if the upstream branch’s name is different from the
			// local one.
			return trackingRef(g, localRef)
		}

		// When pushing to a remote that is different from the remote you normally
		// pull from, work as current.
		return localRef
	case "upstream", "tracking":
		// push the current branch back to the branch whose changes are usually
		// integrated into the current branch
		return trackingRef(g, localRef)
	case "current":
		// push the current branch to update a branch with the same name on the
		// receiving end.
		return localRef
	default:
		tracerx.Printf("WARNING: %q push mode not supported", pushMode)
		return localRef
	}
}

func trackingRef(g Env, localRef *Ref) *Ref {
	if merge, ok := g.Get(fmt.Sprintf("branch.%s.merge", localRef.Name)); ok {
		return ParseRef(merge, "")
	}
	return localRef
}

func (u *RefUpdate) RemoteRefCommitish() string {
	return refCommitish(u.RemoteRef())
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
