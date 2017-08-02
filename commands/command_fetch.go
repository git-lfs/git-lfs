package commands

import (
	"fmt"
	"time"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/progress"
	"github.com/git-lfs/git-lfs/tq"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

var (
	fetchRecentArg bool
	fetchAllArg    bool
	fetchPruneArg  bool
)

func getIncludeExcludeArgs(cmd *cobra.Command) (include, exclude *string) {
	includeFlag := cmd.Flag("include")
	excludeFlag := cmd.Flag("exclude")
	if includeFlag.Changed {
		include = &includeArg
	}
	if excludeFlag.Changed {
		exclude = &excludeArg
	}

	return
}

func fetchCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	var refs []*git.Ref

	if len(args) > 0 {
		// Remote is first arg
		if err := git.ValidateRemote(args[0]); err != nil {
			Exit("Invalid remote name %q", args[0])
		}
		cfg.CurrentRemote = args[0]
	} else {
		cfg.CurrentRemote = ""
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
	gitscanner := lfs.NewGitScanner(nil)
	defer gitscanner.Close()

	include, exclude := getIncludeExcludeArgs(cmd)

	if fetchAllArg {
		if fetchRecentArg || len(args) > 1 {
			Exit("Cannot combine --all with ref arguments or --recent")
		}
		if include != nil || exclude != nil {
			Exit("Cannot combine --all with --include or --exclude")
		}
		if len(cfg.FetchIncludePaths()) > 0 || len(cfg.FetchExcludePaths()) > 0 {
			Print("Ignoring global include / exclude paths to fulfil --all")
		}
		success = fetchAll()

	} else { // !all
		filter := buildFilepathFilter(cfg, include, exclude)

		// Fetch refs sequentially per arg order; duplicates in later refs will be ignored
		for _, ref := range refs {
			Print("Fetching %v", ref.Name)
			s := fetchRef(ref.Sha, filter)
			success = success && s
		}

		if fetchRecentArg || cfg.FetchPruneConfig().FetchRecentAlways {
			s := fetchRecent(refs, filter)
			success = success && s
		}
	}

	if fetchPruneArg {
		fetchconf := cfg.FetchPruneConfig()
		verify := fetchconf.PruneVerifyRemoteAlways
		// no dry-run or verbose options in fetch, assume false
		prune(fetchconf, verify, false, false)
	}

	if !success {
		c := getAPIClient()
		e := c.Endpoints.Endpoint("download", cfg.CurrentRemote)
		Exit("error: failed to fetch some objects from '%s'", e.Url)
	}
}

func pointersToFetchForRef(ref string, filter *filepathfilter.Filter) ([]*lfs.WrappedPointer, error) {
	var pointers []*lfs.WrappedPointer
	var multiErr error
	tempgitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			if multiErr != nil {
				multiErr = fmt.Errorf("%v\n%v", multiErr, err)
			} else {
				multiErr = err
			}
			return
		}

		pointers = append(pointers, p)
	})

	tempgitscanner.Filter = filter

	if err := tempgitscanner.ScanTree(ref); err != nil {
		return nil, err
	}

	tempgitscanner.Close()
	return pointers, multiErr
}

// Fetch all binaries for a given ref (that we don't have already)
func fetchRef(ref string, filter *filepathfilter.Filter) bool {
	pointers, err := pointersToFetchForRef(ref, filter)
	if err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}
	return fetchAndReportToChan(pointers, filter, nil)
}

// Fetch all previous versions of objects from since to ref (not including final state at ref)
// So this will fetch all the '-' sides of the diff from since to ref
func fetchPreviousVersions(ref string, since time.Time, filter *filepathfilter.Filter) bool {
	var pointers []*lfs.WrappedPointer

	tempgitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			Panic(err, "Could not scan for Git LFS previous versions")
			return
		}

		pointers = append(pointers, p)
	})

	tempgitscanner.Filter = filter

	if err := tempgitscanner.ScanPreviousVersions(ref, since, nil); err != nil {
		ExitWithError(err)
	}

	tempgitscanner.Close()
	return fetchAndReportToChan(pointers, filter, nil)
}

// Fetch recent objects based on config
func fetchRecent(alreadyFetchedRefs []*git.Ref, filter *filepathfilter.Filter) bool {
	fetchconf := cfg.FetchPruneConfig()

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
		refs, err := git.RecentBranches(refsSince, fetchconf.FetchRecentRefsIncludeRemotes, cfg.CurrentRemote)
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
				k := fetchRef(ref.Sha, filter)
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
			k := fetchPreviousVersions(commit, commitsSince, filter)
			ok = ok && k
		}

	}
	return ok
}

func fetchAll() bool {
	pointers := scanAll()
	Print("Fetching objects...")
	return fetchAndReportToChan(pointers, nil, nil)
}

func scanAll() []*lfs.WrappedPointer {
	// This could be a long process so use the chan version & report progress
	Print("Scanning for all objects ever referenced...")
	spinner := progress.NewSpinner()
	var numObjs int64

	// use temp gitscanner to collect pointers
	var pointers []*lfs.WrappedPointer
	var multiErr error
	tempgitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			if multiErr != nil {
				multiErr = fmt.Errorf("%v\n%v", multiErr, err)
			} else {
				multiErr = err
			}
			return
		}

		numObjs++
		spinner.Print(OutputWriter, fmt.Sprintf("%d objects found", numObjs))
		pointers = append(pointers, p)
	})

	if err := tempgitscanner.ScanAll(nil); err != nil {
		Panic(err, "Could not scan for Git LFS files")
	}

	tempgitscanner.Close()

	if multiErr != nil {
		Panic(multiErr, "Could not scan for Git LFS files")
	}

	spinner.Finish(OutputWriter, fmt.Sprintf("%d objects found", numObjs))
	return pointers
}

// Fetch and report completion of each OID to a channel (optional, pass nil to skip)
// Returns true if all completed with no errors, false if errors were written to stderr/log
func fetchAndReportToChan(allpointers []*lfs.WrappedPointer, filter *filepathfilter.Filter, out chan<- *lfs.WrappedPointer) bool {
	// Lazily initialize the current remote.
	if len(cfg.CurrentRemote) == 0 {
		// Actively find the default remote, don't just assume origin
		defaultRemote, err := git.DefaultRemote()
		if err != nil {
			Exit("No default remote")
		}
		cfg.CurrentRemote = defaultRemote
	}

	ready, pointers, meter := readyAndMissingPointers(allpointers, filter)
	q := newDownloadQueue(getTransferManifest(), cfg.CurrentRemote, tq.WithProgress(meter))

	if out != nil {
		// If we already have it, or it won't be fetched
		// report it to chan immediately to support pull/checkout
		for _, p := range ready {
			out <- p
		}

		dlwatch := q.Watch()

		go func() {
			// fetch only reports single OID, but OID *might* be referenced by multiple
			// WrappedPointers if same content is at multiple paths, so map oid->slice
			oidToPointers := make(map[string][]*lfs.WrappedPointer, len(pointers))
			for _, pointer := range pointers {
				plist := oidToPointers[pointer.Oid]
				oidToPointers[pointer.Oid] = append(plist, pointer)
			}

			for t := range dlwatch {
				plist, ok := oidToPointers[t.Oid]
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
		tracerx.Printf("fetch %v [%v]", p.Name, p.Oid)

		q.Add(downloadTransfer(p))
	}

	processQueue := time.Now()
	q.Wait()
	tracerx.PerformanceSince("process queue", processQueue)

	ok := true
	for _, err := range q.Errors() {
		ok = false
		FullError(err)
	}
	return ok
}

func readyAndMissingPointers(allpointers []*lfs.WrappedPointer, filter *filepathfilter.Filter) ([]*lfs.WrappedPointer, []*lfs.WrappedPointer, *progress.ProgressMeter) {
	meter := buildProgressMeter(false)
	seen := make(map[string]bool, len(allpointers))
	missing := make([]*lfs.WrappedPointer, 0, len(allpointers))
	ready := make([]*lfs.WrappedPointer, 0, len(allpointers))

	for _, p := range allpointers {
		// no need to download the same object multiple times
		if seen[p.Oid] {
			continue
		}

		seen[p.Oid] = true

		// no need to download objects that exist locally already
		lfs.LinkOrCopyFromReference(p.Oid, p.Size)
		if lfs.ObjectExistsOfSize(p.Oid, p.Size) {
			ready = append(ready, p)
			continue
		}

		missing = append(missing, p)
		meter.Add(p.Size)
	}

	return ready, missing, meter
}

func init() {
	RegisterCommand("fetch", fetchCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
		cmd.Flags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")
		cmd.Flags().BoolVarP(&fetchRecentArg, "recent", "r", false, "Fetch recent refs & commits")
		cmd.Flags().BoolVarP(&fetchAllArg, "all", "a", false, "Fetch all LFS files ever referenced")
		cmd.Flags().BoolVarP(&fetchPruneArg, "prune", "p", false, "After fetching, prune old data")
	})
}
