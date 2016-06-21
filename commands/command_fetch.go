package commands

import (
	"fmt"
	"time"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/progress"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

var (
	fetchCmd = &cobra.Command{
		Use: "fetch",
		Run: fetchCommand,
	}
	fetchIncludeArg string
	fetchExcludeArg string
	fetchRecentArg  bool
	fetchAllArg     bool
	fetchPruneArg   bool
)

func fetchCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	var refs []*git.Ref

	if len(args) > 0 {
		// Remote is first arg
		if err := git.ValidateRemote(args[0]); err != nil {
			Exit("Invalid remote name %q", args[0])
		}
		config.Config.CurrentRemote = args[0]
	} else {
		// Actively find the default remote, don't just assume origin
		defaultRemote, err := git.DefaultRemote()
		if err != nil {
			Exit("No default remote")
		}
		config.Config.CurrentRemote = defaultRemote
	}

	if len(args) > 1 {
		resolvedrefs, err := git.ResolveRefs(args[1:])
		if err != nil {
			Panic(err, "Invalid ref argument: %v", args[1:])
		}
		refs = resolvedrefs
	} else if !fetchAllArg {
		ref, err := git.CurrentRef()
		if err != nil {
			Panic(err, "Could not fetch")
		}
		refs = []*git.Ref{ref}
	}

	success := true
	if fetchAllArg {
		if fetchRecentArg || len(args) > 1 {
			Exit("Cannot combine --all with ref arguments or --recent")
		}
		if fetchIncludeArg != "" || fetchExcludeArg != "" {
			Exit("Cannot combine --all with --include or --exclude")
		}
		if len(config.Config.FetchIncludePaths()) > 0 || len(config.Config.FetchExcludePaths()) > 0 {
			Print("Ignoring global include / exclude paths to fulfil --all")
		}
		success = fetchAll()

	} else { // !all
		includePaths, excludePaths := determineIncludeExcludePaths(config.Config, fetchIncludeArg, fetchExcludeArg)

		// Fetch refs sequentially per arg order; duplicates in later refs will be ignored
		for _, ref := range refs {
			Print("Fetching %v", ref.Name)
			s := fetchRef(ref.Sha, includePaths, excludePaths)
			success = success && s
		}

		if fetchRecentArg || config.Config.FetchPruneConfig().FetchRecentAlways {
			s := fetchRecent(refs, includePaths, excludePaths)
			success = success && s
		}
	}

	if fetchPruneArg {
		verify := config.Config.FetchPruneConfig().PruneVerifyRemoteAlways
		// no dry-run or verbose options in fetch, assume false
		prune(verify, false, false)
	}

	if !success {
		Exit("Warning: errors occurred")
	}
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchIncludeArg, "include", "I", "", "Include a list of paths")
	fetchCmd.Flags().StringVarP(&fetchExcludeArg, "exclude", "X", "", "Exclude a list of paths")
	fetchCmd.Flags().BoolVarP(&fetchRecentArg, "recent", "r", false, "Fetch recent refs & commits")
	fetchCmd.Flags().BoolVarP(&fetchAllArg, "all", "a", false, "Fetch all LFS files ever referenced")
	fetchCmd.Flags().BoolVarP(&fetchPruneArg, "prune", "p", false, "After fetching, prune old data")
	RootCmd.AddCommand(fetchCmd)
}

func pointersToFetchForRef(ref string) ([]*lfs.WrappedPointer, error) {
	// Use SkipDeletedBlobs to avoid fetching ALL previous versions of modified files
	opts := lfs.NewScanRefsOptions()
	opts.ScanMode = lfs.ScanRefsMode
	opts.SkipDeletedBlobs = true
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
func fetchRef(ref string, include, exclude []string) bool {
	pointers, err := pointersToFetchForRef(ref)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}
	return fetchPointers(pointers, include, exclude)
}

// Fetch all previous versions of objects from since to ref (not including final state at ref)
// So this will fetch all the '-' sides of the diff from since to ref
func fetchPreviousVersions(ref string, since time.Time, include, exclude []string) bool {
	pointers, err := lfs.ScanPreviousVersions(ref, since)
	if err != nil {
		Panic(err, "Could not scan for Git LFS previous versions")
	}
	return fetchPointers(pointers, include, exclude)
}

// Fetch recent objects based on config
func fetchRecent(alreadyFetchedRefs []*git.Ref, include, exclude []string) bool {
	fetchconf := config.Config.FetchPruneConfig()

	if fetchconf.FetchRecentRefsDays == 0 && fetchconf.FetchRecentCommitsDays == 0 {
		return true
	}

	ok := true
	// Make a list of what unique commits we've already fetched for to avoid duplicating work
	uniqueRefShas := make(map[string]string, len(alreadyFetchedRefs))
	for _, ref := range alreadyFetchedRefs {
		uniqueRefShas[ref.Sha] = ref.Name
	}
	// First find any other recent refs
	if fetchconf.FetchRecentRefsDays > 0 {
		Print("Fetching recent branches within %v days", fetchconf.FetchRecentRefsDays)
		refsSince := time.Now().AddDate(0, 0, -fetchconf.FetchRecentRefsDays)
		refs, err := git.RecentBranches(refsSince, fetchconf.FetchRecentRefsIncludeRemotes, config.Config.CurrentRemote)
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
				k := fetchRef(ref.Sha, include, exclude)
				ok = ok && k
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
			k := fetchPreviousVersions(commit, commitsSince, include, exclude)
			ok = ok && k
		}

	}
	return ok
}

func fetchAll() bool {
	pointers := scanAll()
	Print("Fetching objects...")
	return fetchPointers(pointers, nil, nil)
}

func scanAll() []*lfs.WrappedPointer {
	// converts to `git rev-list --all`
	// We only pick up objects in real commits and not the reflog
	opts := lfs.NewScanRefsOptions()
	opts.ScanMode = lfs.ScanAllMode
	opts.SkipDeletedBlobs = false

	// This could be a long process so use the chan version & report progress
	Print("Scanning for all objects ever referenced...")
	spinner := progress.NewSpinner()
	var numObjs int64
	pointerchan, err := lfs.ScanRefsToChan("", "", opts)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	pointers := make([]*lfs.WrappedPointer, 0)

	for p := range pointerchan.Results {
		numObjs++
		spinner.Print(OutputWriter, fmt.Sprintf("%d objects found", numObjs))
		pointers = append(pointers, p)
	}
	err = pointerchan.Wait()
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	spinner.Finish(OutputWriter, fmt.Sprintf("%d objects found", numObjs))
	return pointers
}

func fetchPointers(pointers []*lfs.WrappedPointer, include, exclude []string) bool {
	return fetchAndReportToChan(pointers, include, exclude, nil)
}

// Fetch and report completion of each OID to a channel (optional, pass nil to skip)
// Returns true if all completed with no errors, false if errors were written to stderr/log
func fetchAndReportToChan(pointers []*lfs.WrappedPointer, include, exclude []string, out chan<- *lfs.WrappedPointer) bool {
	totalSize := int64(0)
	for _, p := range pointers {
		totalSize += p.Size
	}
	q := lfs.NewDownloadQueue(len(pointers), totalSize, false)

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

	for _, p := range pointers {
		// Only add to download queue if local file is not the right size already
		// This avoids previous case of over-reporting a requirement for files we already have
		// which would only be skipped by PointerSmudgeObject later
		passFilter := lfs.FilenamePassesIncludeExcludeFilter(p.Name, include, exclude)

		lfs.LinkOrCopyFromReference(p.Oid, p.Size)

		if !lfs.ObjectExistsOfSize(p.Oid, p.Size) && passFilter {
			tracerx.Printf("fetch %v [%v]", p.Name, p.Oid)
			q.Add(lfs.NewDownloadable(p))
		} else {
			// Ensure progress matches
			q.Skip(p.Size)
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

	processQueue := time.Now()
	q.Wait()
	tracerx.PerformanceSince("process queue", processQueue)

	ok := true
	for _, err := range q.Errors() {
		ok = false
		ExitWithError(err)
	}
	return ok
}
