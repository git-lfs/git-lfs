package lfs

import (
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/tools"
)

// calcSkippedRefs checks that locally cached versions of remote refs are still
// present on the remote before they are used as a 'from' point. If the server
// implements garbage collection and a remote branch had been deleted since we
// last did 'git fetch --prune', then the objects in that branch may have also
// been deleted on the server if unreferenced. If some refs are missing on the
// remote, use a more explicit diff command.
func calcSkippedRefs(remote string) []string {
	cachedRemoteRefs, _ := git.CachedRemoteRefs(remote, false)

	// Since CachedRemoteRefs() only returns branches, request that
	// RemoteRefs() ignore tags and also return only branches.
	actualRemoteRefs, _ := git.RemoteRefs(remote, false)

	// The list of remote refs can be very large, so convert them to
	// a set for faster lookups in the skip calculation loop.
	actualRemoteRefsSet := tools.NewStringSet()
	for _, ref := range actualRemoteRefs {
		actualRemoteRefsSet.Add(ref.Name)
	}

	// Only check for missing refs on remote; if the ref is different it has moved
	// forward probably, and if not and the ref has changed to a non-descendant
	// (force push) then that will cause a re-evaluation in a subsequent command.
	var skippedRefs []string
	for _, cachedRef := range cachedRemoteRefs {
		if actualRemoteRefsSet.Contains(cachedRef.Name) {
			skippedRefs = append(skippedRefs, "^"+cachedRef.Sha)
		}
	}
	return skippedRefs
}
