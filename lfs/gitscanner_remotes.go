package lfs

import (
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tools"
)

// calcSkippedRefs checks that locally cached versions of remote refs are still
// present on the remote before they are used as a 'from' point. If the server
// implements garbage collection and a remote branch had been deleted since we
// last did 'git fetch --prune', then the objects in that branch may have also
// been deleted on the server if unreferenced. If some refs are missing on the
// remote, use a more explicit diff command.
func calcSkippedRefs(remote string) []string {
	cachedRemoteRefs, _ := git.CachedRemoteRefs(remote)
	actualRemoteRefs, _ := git.RemoteRefs(remote)

	// Only check for missing refs on remote; if the ref is different it has moved
	// forward probably, and if not and the ref has changed to a non-descendant
	// (force push) then that will cause a re-evaluation in a subsequent command.
	missingRefs := tools.NewStringSet()
	for _, cachedRef := range cachedRemoteRefs {
		found := false
		for _, realRemoteRef := range actualRemoteRefs {
			if cachedRef.Type == realRemoteRef.Type && cachedRef.Name == realRemoteRef.Name {
				found = true
				break
			}
		}
		if !found {
			missingRefs.Add(cachedRef.Name)
		}
	}

	if len(missingRefs) == 0 {
		return nil
	}

	skippedRefs := make([]string, 0, len(cachedRemoteRefs)-missingRefs.Cardinality())
	for _, cachedRef := range cachedRemoteRefs {
		if !missingRefs.Contains(cachedRef.Name) {
			skippedRefs = append(skippedRefs, "^"+cachedRef.Sha)
		}
	}
	return skippedRefs
}
