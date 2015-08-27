package commands

import (
	"time"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	fetchCmd = &cobra.Command{
		Use:   "fetch",
		Short: "Downloads LFS files",
		Run:   fetchCommand,
	}
	fetchIncludeArg string
	fetchExcludeArg string
	fetchRecentArg  bool
)

func fetchCommand(cmd *cobra.Command, args []string) {
	var refs []*git.Ref

	if len(args) > 0 {
		// Remote is first arg
		lfs.Config.CurrentRemote = args[0]
	} else {
		trackedRemote, err := git.CurrentRemote()
		if err == nil {
			lfs.Config.CurrentRemote = trackedRemote
		} // otherwise leave as default (origin)
	}
	if len(args) > 1 {
		for _, r := range args[1:] {
			ref, err := git.ResolveRef(r)
			if err != nil {
				Panic(err, "Invalid ref argument")
			}
			refs = append(refs, ref)
		}
	} else {
		ref, err := git.CurrentRef()
		if err != nil {
			Panic(err, "Could not fetch")
		}
		refs = []*git.Ref{ref}
	}

	includePaths, excludePaths := determineIncludeExcludePaths(fetchIncludeArg, fetchExcludeArg)

	// Fetch refs sequentially per arg order; duplicates in later refs will be ignored
	for _, ref := range refs {
		Print("Fetching %v", ref.Name)
		fetchRef(ref.Sha, includePaths, excludePaths)
	}

	if fetchRecentArg || lfs.Config.FetchPruneConfig().FetchRecentAlways {
		fetchRecent(refs, includePaths, excludePaths)
	}

}

func init() {
	fetchCmd.Flags().StringVarP(&fetchIncludeArg, "include", "I", "", "Include a list of paths")
	fetchCmd.Flags().StringVarP(&fetchExcludeArg, "exclude", "X", "", "Exclude a list of paths")
	fetchCmd.Flags().BoolVarP(&fetchRecentArg, "recent", "r", false, "Fetch recent refs & commits")
	RootCmd.AddCommand(fetchCmd)
}

func pointersToFetchForRef(ref string) ([]*lfs.WrappedPointer, error) {
	// Use SkipDeletedBlobs to avoid fetching ALL previous versions of modified files
	opts := &lfs.ScanRefsOptions{ScanMode: lfs.ScanRefsMode, SkipDeletedBlobs: true}
	return lfs.ScanRefs(ref, "", opts)
}

func fetchRefToChan(ref string, include, exclude []string) chan *lfs.WrappedPointer {
	c := make(chan *lfs.WrappedPointer)
	pointers, err := pointersToFetchForRef(ref)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	go fetchAndReportToChan(pointers, include, exclude, c)

	return c
}

// Fetch all binaries for a given ref (that we don't have already)
func fetchRef(ref string, include, exclude []string) {
	pointers, err := pointersToFetchForRef(ref)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}
	fetchPointers(pointers, include, exclude)
}

// Fetch all previous versions of objects from since to ref (not including final state at ref)
// So this will fetch all the '-' sides of the diff from since to ref
func fetchPreviousVersions(ref string, since time.Time, include, exclude []string) {
	pointers, err := lfs.ScanPreviousVersions(ref, since)
	if err != nil {
		Panic(err, "Could not scan for Git LFS previous versions")
	}
	fetchPointers(pointers, include, exclude)
}

// Fetch recent objects based on config
func fetchRecent(alreadyFetchedRefs []*git.Ref, include, exclude []string) {
	fetchconf := lfs.Config.FetchPruneConfig()

	if fetchconf.FetchRecentRefsDays == 0 && fetchconf.FetchRecentCommitsDays == 0 {
		return
	}

	// Make a list of what unique commits we've already fetched for to avoid duplicating work
	uniqueRefShas := make(map[string]string, len(alreadyFetchedRefs))
	for _, ref := range alreadyFetchedRefs {
		uniqueRefShas[ref.Sha] = ref.Name
	}
	// First find any other recent refs
	if fetchconf.FetchRecentRefsDays > 0 {
		Print("Fetching recent branches within %v days", fetchconf.FetchRecentRefsDays)
		refsSince := time.Now().AddDate(0, 0, -fetchconf.FetchRecentRefsDays)
		refs, err := git.RecentBranches(refsSince, fetchconf.FetchRecentRefsIncludeRemotes, "")
		if err != nil {
			Panic(err, "Could not scan for recent refs")
		}
		for _, ref := range refs {
			// Don't fetch for the same SHA twice
			if prevRefName, ok := uniqueRefShas[ref.Sha]; ok {
				if ref.Name != prevRefName {
					tracerx.Printf("Skipping fetch for %v, already fetched via %v", ref.Name, prevRefName)
				}
			} else {
				uniqueRefShas[ref.Sha] = ref.Name
				Print("Fetching %v", ref.Name)
				fetchRef(ref.Sha, include, exclude)
			}
		}
	}
	// For every unique commit we've fetched, check recent commits too
	if fetchconf.FetchRecentCommitsDays > 0 {
		for commit, refName := range uniqueRefShas {
			// We measure from the last commit at the ref
			summ, err := git.GetCommitSummary(commit)
			if err != nil {
				Error("Couldn't scan commits at %v: %v", refName, err)
				continue
			}
			Print("Fetching changes within %v days of %v", fetchconf.FetchRecentCommitsDays, refName)
			commitsSince := summ.CommitDate.AddDate(0, 0, -fetchconf.FetchRecentCommitsDays)
			fetchPreviousVersions(commit, commitsSince, include, exclude)
		}

	}
}

func fetchPointers(pointers []*lfs.WrappedPointer, include, exclude []string) {
	fetchAndReportToChan(pointers, include, exclude, nil)
}

// Fetch and report completion of each OID to a channel (optional, pass nil to skip)
func fetchAndReportToChan(pointers []*lfs.WrappedPointer, include, exclude []string, out chan<- *lfs.WrappedPointer) {

	totalSize := int64(0)
	for _, p := range pointers {
		totalSize += p.Size
	}
	q := lfs.NewDownloadQueue(len(pointers), totalSize, false)

	for _, p := range pointers {
		// Only add to download queue if local file is not the right size already
		// This avoids previous case of over-reporting a requirement for files we already have
		// which would only be skipped by PointerSmudgeObject later
		passFilter := lfs.FilenamePassesIncludeExcludeFilter(p.Name, include, exclude)
		if !lfs.ObjectExistsOfSize(p.Oid, p.Size) && passFilter {
			tracerx.Printf("fetch %v [%v]", p.Name, p.Oid)
			q.Add(lfs.NewDownloadable(p))
		} else {
			if !passFilter {
				tracerx.Printf("Skipping %v [%v], include/exclude filters applied", p.Name, p.Oid)
			} else {
				tracerx.Printf("Skipping %v [%v], already exists", p.Name, p.Oid)
			}

			// If we already have it, or it won't be fetched
			// report it to chan immediately to support pull/checkout
			if out != nil {
				out <- p
			}

		}
	}

	if out != nil {
		dlwatch := q.Watch()

		go func() {
			// fetch only reports single OID, but OID *might* be referenced by multiple
			// WrappedPointers if same content is at multiple paths, so map oid->slice
			oidToPointers := make(map[string][]*lfs.WrappedPointer, len(pointers))
			for _, pointer := range pointers {
				plist := oidToPointers[pointer.Oid]
				oidToPointers[pointer.Oid] = append(plist, pointer)
			}

			for oid := range dlwatch {
				plist, ok := oidToPointers[oid]
				if !ok {
					continue
				}
				for _, p := range plist {
					out <- p
				}
			}
			close(out)
		}()

	}
	processQueue := time.Now()
	q.Wait()
	tracerx.PerformanceSince("process queue", processQueue)

	for _, err := range q.Errors() {
		if Debugging || err.Panic {
			LoggedError(err.Err, err.Error())
		} else {
			Error(err.Error())
		}
	}
}
