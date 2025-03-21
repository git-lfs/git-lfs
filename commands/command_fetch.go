package commands

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/filepathfilter"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tasklog"
	"github.com/git-lfs/git-lfs/v3/tq"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

var (
	fetchRecentArg  bool
	fetchAllArg     bool
	fetchPruneArg   bool
	fetchRefetchArg bool
	fetchDryRunArg  bool
	fetchJsonArg    bool
)

type fetchWatcher struct {
	transfers []*tq.Transfer
	observed  map[string]bool
}

func (d *fetchWatcher) registerTransfer(t *tq.Transfer) {
	if fetchJsonArg {
		d.transfers = append(d.transfers, t)
	}
	if fetchDryRunArg || fetchRefetchArg {
		if d.observed == nil {
			d.observed = make(map[string]bool)
		}
		d.observed[t.Oid] = true
	}
	if fetchDryRunArg {
		printProgress("%s %s => %s", tr.Tr.Get("fetch"), t.Oid, t.Name)
	}
}

func (d *fetchWatcher) dumpJson() {
	data := struct {
		Transfers []*tq.Transfer `json:"transfers"`
	}{Transfers: d.transfers}
	if data.Transfers == nil {
		data.Transfers = []*tq.Transfer{}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(data); err != nil {
		ExitWithError(err)
	}
}

func (d *fetchWatcher) hasObserved(oid string) bool {
	return d.observed[oid]
}

func hasToPrintTransfers() bool {
	return fetchJsonArg || fetchDryRunArg
}

func printProgress(format string, args ...interface{}) {
	// TODO: Adjust all our commands to use a common function to output
	//       progress messages to stderr, following Git's model.
	Error(format, args...)
}

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
	setupRepository()

	var refs []*git.Ref

	if len(args) > 0 {
		// Remote is first arg
		if err := cfg.SetValidRemote(args[0]); err != nil {
			Exit(tr.Tr.Get("Invalid remote name %q: %s", args[0], err))
		}
	}

	if len(args) > 1 {
		resolvedrefs, err := git.ResolveRefs(args[1:])
		if err != nil {
			Panic(err, tr.Tr.Get("Invalid ref argument: %v", args[1:]))
		}
		refs = resolvedrefs
	} else if !fetchAllArg {
		ref, err := git.CurrentRef()
		if err != nil {
			Panic(err, tr.Tr.Get("Could not fetch"))
		}
		refs = []*git.Ref{ref}
	}

	if fetchJsonArg && fetchPruneArg {
		// git lfs prune has no `--json` flag, so let's not allow that here for the moment
		Exit(tr.Tr.Get("Cannot combine --json with --prune"))
	}

	success := true
	include, exclude := getIncludeExcludeArgs(cmd)
	fetchPruneCfg := lfs.NewFetchPruneConfig(cfg.Git)

	watcher := &fetchWatcher{}

	if fetchAllArg {
		if fetchRecentArg {
			Exit(tr.Tr.Get("Cannot combine --all with --recent"))
		}
		if include != nil || exclude != nil {
			Exit(tr.Tr.Get("Cannot combine --all with --include or --exclude"))
		}
		if len(cfg.FetchIncludePaths()) > 0 || len(cfg.FetchExcludePaths()) > 0 {
			printProgress(tr.Tr.Get("Ignoring global include / exclude paths to fulfil --all"))
		}

		if len(args) > 1 {
			refShas := make([]string, 0, len(refs))
			for _, ref := range refs {
				refShas = append(refShas, ref.Sha)
			}
			success = fetchRefs(refShas, watcher)
		} else {
			success = fetchAll(watcher)
		}

	} else { // !all
		filter := buildFilepathFilter(cfg, include, exclude, true)

		// Fetch refs sequentially per arg order; duplicates in later refs will be ignored
		for _, ref := range refs {
			printProgress(tr.Tr.Get("Fetching reference %s", ref.Refspec()))
			s := fetchRef(ref.Sha, filter, watcher)
			success = success && s
		}

		if fetchRecentArg || fetchPruneCfg.FetchRecentAlways {
			s := fetchRecent(fetchPruneCfg, refs, filter, watcher)
			success = success && s
		}
	}

	if fetchPruneArg {
		verify := fetchPruneCfg.PruneVerifyRemoteAlways
		verifyUnreachable := fetchPruneCfg.PruneVerifyUnreachableAlways

		// assume false for non available options in fetch
		prune(fetchPruneCfg, verify, verifyUnreachable, false, fetchDryRunArg, fetchDryRunArg)
	}

	if !success {
		c := getAPIClient()
		e := c.Endpoints.Endpoint("download", cfg.Remote())
		Exit(tr.Tr.Get("error: failed to fetch some objects from '%s'", e.Url))
	}
	if fetchJsonArg {
		watcher.dumpJson()
	}
}

func pointersToFetchForRef(ref string, filter *filepathfilter.Filter) ([]*lfs.WrappedPointer, error) {
	var pointers []*lfs.WrappedPointer
	var multiErr error
	tempgitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			multiErr = errors.Join(multiErr, err)
			return
		}

		pointers = append(pointers, p)
	})

	tempgitscanner.Filter = filter

	if err := tempgitscanner.ScanTree(ref, nil); err != nil {
		return nil, err
	}

	return pointers, multiErr
}

// Fetch all binaries for a given ref (that we don't have already)
func fetchRef(ref string, filter *filepathfilter.Filter, watcher *fetchWatcher) bool {
	pointers, err := pointersToFetchForRef(ref, filter)
	if err != nil {
		Panic(err, tr.Tr.Get("Could not scan for Git LFS files"))
	}
	return fetch(pointers, watcher)
}

func pointersToFetchForRefs(refs []string) ([]*lfs.WrappedPointer, error) {
	// This could be a long process so use the chan version & report progress
	logger := tasklog.NewLogger(OutputWriter,
		tasklog.ForceProgress(cfg.ForceProgress()),
	)
	task := logger.Simple()
	defer task.Complete()

	// use temp gitscanner to collect pointers
	var pointers []*lfs.WrappedPointer
	var multiErr error
	var numObjs int64
	tempgitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			multiErr = errors.Join(multiErr, err)
			return
		}

		numObjs++
		task.Log(tr.Tr.GetN("%d object found", "%d objects found", int(numObjs), numObjs))
		pointers = append(pointers, p)
	})

	if err := tempgitscanner.ScanRefs(refs, nil, nil); err != nil {
		return nil, err
	}

	return pointers, multiErr
}

func fetchRefs(refs []string, watcher *fetchWatcher) bool {
	pointers, err := pointersToFetchForRefs(refs)
	if err != nil {
		Panic(err, tr.Tr.Get("Could not scan for Git LFS files"))
	}
	return fetch(pointers, watcher)
}

// Fetch all previous versions of objects from since to ref (not including final state at ref)
// So this will fetch all the '-' sides of the diff from since to ref
func fetchPreviousVersions(ref string, since time.Time, filter *filepathfilter.Filter, watcher *fetchWatcher) bool {
	var pointers []*lfs.WrappedPointer

	tempgitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			Panic(err, tr.Tr.Get("Could not scan for Git LFS previous versions"))
			return
		}

		pointers = append(pointers, p)
	})

	tempgitscanner.Filter = filter

	if err := tempgitscanner.ScanPreviousVersions(ref, since, nil); err != nil {
		ExitWithError(err)
	}

	return fetch(pointers, watcher)
}

// Fetch recent objects based on config
func fetchRecent(fetchconf lfs.FetchPruneConfig, alreadyFetchedRefs []*git.Ref, filter *filepathfilter.Filter, watcher *fetchWatcher) bool {
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
		printProgress(tr.Tr.GetN(
			"Fetching recent branches within %v day",
			"Fetching recent branches within %v days",
			fetchconf.FetchRecentRefsDays,
			fetchconf.FetchRecentRefsDays,
		))
		refsSince := time.Now().AddDate(0, 0, -fetchconf.FetchRecentRefsDays)
		refs, err := git.RecentBranches(refsSince, fetchconf.FetchRecentRefsIncludeRemotes, cfg.Remote())
		if err != nil {
			Panic(err, tr.Tr.Get("Could not scan for recent refs"))
		}
		for _, ref := range refs {
			// Don't fetch for the same SHA twice
			if prevRefName, ok := uniqueRefShas[ref.Sha]; ok {
				if ref.Name != prevRefName {
					tracerx.Printf("Skipping fetch for %v, already fetched via %v", ref.Name, prevRefName)
				}
			} else {
				uniqueRefShas[ref.Sha] = ref.Name
				printProgress(tr.Tr.Get("Fetching reference %s", ref.Name))
				k := fetchRef(ref.Sha, filter, watcher)
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
				Error(tr.Tr.Get("Couldn't scan commits at %v: %v", refName, err))
				continue
			}
			printProgress(tr.Tr.GetN(
				"Fetching changes within %v day of %v",
				"Fetching changes within %v days of %v",
				fetchconf.FetchRecentCommitsDays,
				fetchconf.FetchRecentCommitsDays,
				refName,
			))
			commitsSince := summ.CommitDate.AddDate(0, 0, -fetchconf.FetchRecentCommitsDays)
			k := fetchPreviousVersions(commit, commitsSince, filter, watcher)
			ok = ok && k
		}

	}
	return ok
}

func fetchAll(watcher *fetchWatcher) bool {
	pointers := scanAll()
	printProgress(tr.Tr.Get("Fetching all references..."))
	return fetch(pointers, watcher)
}

func scanAll() []*lfs.WrappedPointer {
	// This could be a long process so use the chan version & report progress
	logger := tasklog.NewLogger(OutputWriter,
		tasklog.ForceProgress(cfg.ForceProgress()),
	)
	task := logger.Simple()
	defer task.Complete()

	// use temp gitscanner to collect pointers
	var pointers []*lfs.WrappedPointer
	var multiErr error
	var numObjs int64
	tempgitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			multiErr = errors.Join(multiErr, err)
			return
		}

		numObjs++
		task.Log(tr.Tr.GetN("%d object found", "%d objects found", int(numObjs), numObjs))
		pointers = append(pointers, p)
	})

	if err := tempgitscanner.ScanAll(nil); err != nil {
		Panic(err, tr.Tr.Get("Could not scan for Git LFS files"))
	}

	if multiErr != nil {
		Panic(multiErr, tr.Tr.Get("Could not scan for Git LFS files"))
	}

	return pointers
}

// Fetch
// Returns true if all completed with no errors, false if errors were written to stderr/log
func fetch(allPointers []*lfs.WrappedPointer, watcher *fetchWatcher) bool {
	pointersToFetch, meter := pointersToFetch(allPointers, watcher)
	q := newDownloadQueue(
		getTransferManifestOperationRemote("download", cfg.Remote()),
		cfg.Remote(), tq.WithProgress(meter), tq.DryRun(fetchDryRunArg),
	)
	var wg sync.WaitGroup

	if watcher != nil {
		transfers := q.Watch()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range transfers {
				watcher.registerTransfer(t)
			}
		}()
	}

	for _, p := range pointersToFetch {
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
	if watcher != nil {
		wg.Wait()
	}
	return ok
}

func pointersToFetch(allPointers []*lfs.WrappedPointer, watcher *fetchWatcher) ([]*lfs.WrappedPointer, *tq.Meter) {
	logger := tasklog.NewLogger(os.Stdout,
		tasklog.ForceProgress(cfg.ForceProgress()),
	)
	meter := buildProgressMeter(hasToPrintTransfers(), tq.Download)
	logger.Enqueue(meter)

	pointersToFetch := make([]*lfs.WrappedPointer, 0, len(allPointers))

	for _, p := range allPointers {
		// if running with --dry-run or --refetch, skip objects that have already been virtually or
		// already forcefully fetched
		if watcher != nil && watcher.hasObserved(p.Oid) {
			continue
		}

		// empty files are special, we always skip them in upload/download operations
		if p.Size == 0 {
			continue
		}

		// no need to download objects that exist locally already, unless `--refetch` was provided
		lfs.LinkOrCopyFromReference(cfg, p.Oid, p.Size)
		if !fetchRefetchArg && cfg.LFSObjectExists(p.Oid, p.Size) {
			continue
		}

		pointersToFetch = append(pointersToFetch, p)
		meter.Add(p.Size)
	}

	return pointersToFetch, meter
}

func init() {
	RegisterCommand("fetch", fetchCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
		cmd.Flags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")
		cmd.Flags().BoolVarP(&fetchRecentArg, "recent", "r", false, "Fetch recent refs & commits")
		cmd.Flags().BoolVarP(&fetchAllArg, "all", "a", false, "Fetch all LFS files ever referenced")
		cmd.Flags().BoolVarP(&fetchPruneArg, "prune", "p", false, "After fetching, prune old data")
		cmd.Flags().BoolVar(&fetchRefetchArg, "refetch", false, "Also fetch objects that are already present locally")
		cmd.Flags().BoolVarP(&fetchDryRunArg, "dry-run", "d", false, "Do not fetch, only show what would be fetched")
		cmd.Flags().BoolVarP(&fetchJsonArg, "json", "j", false, "Give the output in a stable JSON format for scripts")
	})
}
