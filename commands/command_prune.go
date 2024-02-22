package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/filepathfilter"
	"github.com/git-lfs/git-lfs/v3/fs"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tasklog"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tools/humanize"
	"github.com/git-lfs/git-lfs/v3/tq"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

var (
	pruneDryRunArg                 bool
	pruneVerboseArg                bool
	pruneVerifyArg                 bool
	pruneRecentArg                 bool
	pruneForceArg                  bool
	pruneDoNotVerifyArg            bool
	pruneVerifyUnreachableArg      bool
	pruneDoNotVerifyUnreachableArg bool
	pruneWhenUnverifiedArg         string
)

func pruneCommand(cmd *cobra.Command, args []string) {
	// Guts of this must be re-usable from fetch --prune so just parse & dispatch
	if pruneVerifyArg && pruneDoNotVerifyArg {
		Exit(tr.Tr.Get("Cannot specify both --verify-remote and --no-verify-remote"))
	}

	fetchPruneConfig := lfs.NewFetchPruneConfig(cfg.Git)
	verify := !pruneDoNotVerifyArg &&
		(fetchPruneConfig.PruneVerifyRemoteAlways || pruneVerifyArg)
	verifyUnreachable := !pruneDoNotVerifyUnreachableArg && (pruneVerifyUnreachableArg || fetchPruneConfig.PruneVerifyUnreachableAlways)

	continueWhenUnverified := false
	switch pruneWhenUnverifiedArg {
	case "halt":
		continueWhenUnverified = false
	case "continue":
		continueWhenUnverified = true
	default:
		Exit(tr.Tr.Get("Invalid value for --when-unverified: %s", pruneWhenUnverifiedArg))
	}

	fetchPruneConfig.PruneRecent = pruneRecentArg || pruneForceArg
	fetchPruneConfig.PruneForce = pruneForceArg
	prune(fetchPruneConfig, verify, verifyUnreachable, continueWhenUnverified, pruneDryRunArg, pruneVerboseArg)
}

type PruneProgressType int

const (
	PruneProgressTypeLocal      = PruneProgressType(iota)
	PruneProgressTypeRetain     = PruneProgressType(iota)
	PruneProgressTypeVerify     = PruneProgressType(iota)
	PruneProgressTypeUnverified = PruneProgressType(iota)
)

// Progress from a sub-task of prune
type PruneProgress struct {
	ProgressType PruneProgressType
	Count        int // Number of items done
}
type PruneProgressChan chan PruneProgress

func prune(fetchPruneConfig lfs.FetchPruneConfig, verifyRemote, verifyUnreachable, continueWhenUnverified, dryRun, verbose bool) {
	localObjects := make([]fs.Object, 0, 100)
	retainedObjects := tools.NewStringSetWithCapacity(100)

	logger := tasklog.NewLogger(OutputWriter,
		tasklog.ForceProgress(cfg.ForceProgress()),
	)
	defer logger.Close()

	var reachableObjects tools.StringSet
	var taskwait sync.WaitGroup

	// Add all the base funcs to the waitgroup before starting them, in case
	// one completes really fast & hits 0 unexpectedly
	// each main process can Add() to the wg itself if it subdivides the task
	taskwait.Add(5) // 1..5: localObjects, current & recent refs, unpushed, worktree, stashes
	if verifyRemote && !verifyUnreachable {
		taskwait.Add(1) // 6
	}

	progressChan := make(PruneProgressChan, 100)

	// Collect errors
	errorChan := make(chan error, 10)
	var errorwait sync.WaitGroup
	errorwait.Add(1)
	var taskErrors []error
	go pruneTaskCollectErrors(&taskErrors, errorChan, &errorwait)

	// Populate the single list of local objects
	go pruneTaskGetLocalObjects(&localObjects, progressChan, &taskwait)

	// Now find files to be retained from many sources
	retainChan := make(chan string, 100)

	gitscanner := lfs.NewGitScanner(cfg, nil)
	gitscanner.Filter = filepathfilter.New(nil, cfg.FetchExcludePaths(), filepathfilter.GitIgnore)

	sem := semaphore.NewWeighted(int64(runtime.NumCPU() * 2))

	go pruneTaskGetRetainedCurrentAndRecentRefs(gitscanner, fetchPruneConfig, retainChan, errorChan, &taskwait, sem)
	go pruneTaskGetRetainedUnpushed(gitscanner, fetchPruneConfig, retainChan, errorChan, &taskwait, sem)
	go pruneTaskGetRetainedWorktree(gitscanner, fetchPruneConfig, retainChan, errorChan, &taskwait, sem)
	go pruneTaskGetRetainedStashed(gitscanner, retainChan, errorChan, &taskwait, sem)
	if verifyRemote && !verifyUnreachable {
		reachableObjects = tools.NewStringSetWithCapacity(100)
		go pruneTaskGetReachableObjects(gitscanner, &reachableObjects, errorChan, &taskwait, sem)
	}

	// Now collect all the retained objects, on separate wait
	var retainwait sync.WaitGroup
	retainwait.Add(1)
	go pruneTaskCollectRetained(&retainedObjects, retainChan, progressChan, &retainwait)

	// Report progress
	var progresswait sync.WaitGroup
	progresswait.Add(1)
	go pruneTaskDisplayProgress(progressChan, &progresswait, logger)

	taskwait.Wait()   // wait for subtasks
	close(retainChan) // triggers retain collector to end now all tasks have
	retainwait.Wait() // make sure all retained objects added

	close(errorChan) // triggers error collector to end now all tasks have
	errorwait.Wait() // make sure all errors have been processed
	pruneCheckErrors(taskErrors)

	prunableObjects := make([]string, 0, len(localObjects)/2)

	// Build list of prunables (also queue for verify at same time if applicable)
	var verifyQueue *tq.TransferQueue
	var verifiedObjects tools.StringSet
	var totalSize int64
	var verboseOutput []string
	var verifyc chan *tq.Transfer
	var verifywait sync.WaitGroup

	if verifyRemote {
		verifyQueue = newDownloadCheckQueue(
			getTransferManifestOperationRemote("download", fetchPruneConfig.PruneRemoteName),
			fetchPruneConfig.PruneRemoteName,
		)
		verifiedObjects = tools.NewStringSetWithCapacity(len(localObjects) / 2)

		// this channel is filled with oids for which Check() succeeded & Transfer() was called
		verifyc = verifyQueue.Watch()
		verifywait.Add(1)
		go func() {
			for t := range verifyc {
				verifiedObjects.Add(t.Oid)
				tracerx.Printf("VERIFIED: %v", t.Oid)
				progressChan <- PruneProgress{PruneProgressTypeVerify, 1}
			}
			verifywait.Done()
		}()
	}

	for _, file := range localObjects {
		if !retainedObjects.Contains(file.Oid) {
			prunableObjects = append(prunableObjects, file.Oid)
			totalSize += file.Size
			if verbose {
				// Save up verbose output for the end.
				verboseOutput = append(verboseOutput,
					fmt.Sprintf("%s (%s)",
						file.Oid,
						humanize.FormatBytes(uint64(file.Size))))
			}

			if verifyRemote {
				verifyQueue.Add(downloadTransfer(&lfs.WrappedPointer{
					Pointer: lfs.NewPointer(file.Oid, file.Size, nil),
				}))
			}
		}
	}

	if verifyRemote {
		verifyQueue.Wait()
		verifywait.Wait()

		var problems bytes.Buffer
		prunableObjectsLen := len(prunableObjects)
		prunableObjects, problems = pruneGetVerifiedPrunableObjects(prunableObjects, reachableObjects, verifiedObjects, verifyUnreachable)
		if prunableObjectsLen != len(prunableObjects) {
			progressChan <- PruneProgress{PruneProgressTypeUnverified, prunableObjectsLen - len(prunableObjects)}
		}

		close(progressChan) // after verify but before check
		progresswait.Wait()

		if !continueWhenUnverified && problems.Len() > 0 {
			Exit("%s\n%v", tr.Tr.Get("These objects to be pruned are missing on remote:"), problems.String())
		}
	} else {
		close(progressChan)
		progresswait.Wait()
	}

	if len(prunableObjects) == 0 {
		return
	}

	logVerboseOutput(logger, verboseOutput, len(prunableObjects), totalSize, dryRun)

	if !dryRun {
		pruneDeleteFiles(prunableObjects, logger)
	}
}

func logVerboseOutput(logger *tasklog.Logger, verboseOutput []string, numPrunableObjects int, totalSize int64, dryRun bool) {
	info := logger.Simple()
	defer info.Complete()

	if dryRun {
		info.Logf("prune: %s", tr.Tr.GetN(
			"%d file would be pruned (%s)",
			"%d files would be pruned (%s)",
			numPrunableObjects,
			numPrunableObjects,
			humanize.FormatBytes(uint64(totalSize))))
		for _, item := range verboseOutput {
			info.Logf("\n * %s", item)
		}
	} else {
		for _, item := range verboseOutput {
			info.Logf("\n%s", item)
		}
	}
}

func pruneGetVerifiedPrunableObjects(prunableObjects []string, reachableObjects, verifiedObjects tools.StringSet, verifyUnreachable bool) ([]string, bytes.Buffer) {
	verifiedPrunableObjects := make([]string, 0, len(verifiedObjects))
	var unverified bytes.Buffer

	for _, oid := range prunableObjects {
		if verifiedObjects.Contains(oid) {
			verifiedPrunableObjects = append(verifiedPrunableObjects, oid)
		} else {
			if verifyUnreachable {
				tracerx.Printf("UNVERIFIED: %v", oid)
				unverified.WriteString(fmt.Sprintf(" * %v\n", oid))
			} else {
				// There's no issue if an object is not reachable and missing, only if reachable & missing
				if reachableObjects.Contains(oid) {
					unverified.WriteString(fmt.Sprintf(" * %v\n", oid))
				} else {
					// Just to indicate why it doesn't matter that we didn't verify
					tracerx.Printf("UNREACHABLE: %v", oid)
					verifiedPrunableObjects = append(verifiedPrunableObjects, oid)
				}
			}
		}
	}

	return verifiedPrunableObjects, unverified
}

func pruneCheckErrors(taskErrors []error) {
	if len(taskErrors) > 0 {
		for _, err := range taskErrors {
			LoggedError(err, tr.Tr.Get("Prune error: %v", err))
		}
		Exit(tr.Tr.Get("Prune sub-tasks failed, cannot continue"))
	}
}

func pruneTaskDisplayProgress(progressChan PruneProgressChan, waitg *sync.WaitGroup, logger *tasklog.Logger) {
	defer waitg.Done()

	task := logger.Simple()
	defer task.Complete()

	localCount := 0
	retainCount := 0
	verifyCount := 0
	notRemoteCount := 0
	var msg string
	for p := range progressChan {
		switch p.ProgressType {
		case PruneProgressTypeLocal:
			localCount++
		case PruneProgressTypeRetain:
			retainCount++
		case PruneProgressTypeVerify:
			verifyCount++
		case PruneProgressTypeUnverified:
			notRemoteCount += p.Count
		}
		msg = fmt.Sprintf("prune: %s, %s",
			tr.Tr.GetN("%d local object", "%d local objects", localCount, localCount),
			tr.Tr.GetN("%d retained", "%d retained", retainCount, retainCount))
		if verifyCount > 0 {
			msg += tr.Tr.GetN(", %d verified with remote", ", %d verified with remote", verifyCount, verifyCount)
		}
		if notRemoteCount > 0 {
			msg += tr.Tr.GetN(", %d not on remote", ", %d not on remote", notRemoteCount, notRemoteCount)
		}
		task.Log(msg)
	}
}

func pruneTaskCollectRetained(outRetainedObjects *tools.StringSet, retainChan chan string,
	progressChan PruneProgressChan, retainwait *sync.WaitGroup) {

	defer retainwait.Done()

	for oid := range retainChan {
		if outRetainedObjects.Add(oid) {
			progressChan <- PruneProgress{PruneProgressTypeRetain, 1}
		}
	}

}

func pruneTaskCollectErrors(outtaskErrors *[]error, errorChan chan error, errorwait *sync.WaitGroup) {
	defer errorwait.Done()

	for err := range errorChan {
		*outtaskErrors = append(*outtaskErrors, err)
	}
}

func pruneDeleteFiles(prunableObjects []string, logger *tasklog.Logger) {
	task := logger.Percentage(fmt.Sprintf("prune: %s", tr.Tr.Get("Deleting objects")), uint64(len(prunableObjects)))
	defer task.Complete()

	var problems bytes.Buffer
	// In case we fail to delete some
	var deletedFiles int
	for _, oid := range prunableObjects {
		mediaFile, err := cfg.Filesystem().ObjectPath(oid)
		if err != nil {
			problems.WriteString(tr.Tr.Get("Unable to find media path for %v: %v", oid, err))
			problems.WriteRune('\n')
			continue
		}
		if mediaFile == os.DevNull {
			continue
		}
		err = os.Remove(mediaFile)
		if err != nil {
			problems.WriteString(tr.Tr.Get("Failed to remove file %v: %v", mediaFile, err))
			problems.WriteRune('\n')
			continue
		}
		deletedFiles++
		task.Count(1)
	}
	if problems.Len() > 0 {
		LoggedError(errors.New(tr.Tr.Get("failed to delete some files")), problems.String())
		Exit(tr.Tr.Get("Prune failed, see errors above"))
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetLocalObjects(outLocalObjects *[]fs.Object, progChan PruneProgressChan, waitg *sync.WaitGroup) {
	defer waitg.Done()

	cfg.EachLFSObject(func(obj fs.Object) error {
		*outLocalObjects = append(*outLocalObjects, obj)
		progChan <- PruneProgress{PruneProgressTypeLocal, 1}
		return nil
	})
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedAtRef(gitscanner *lfs.GitScanner, ref string, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup, sem *semaphore.Weighted) {
	sem.Acquire(context.Background(), 1)
	defer sem.Release(1)
	defer waitg.Done()

	err := gitscanner.ScanTree(ref, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			errorChan <- err
			return
		}

		retainChan <- p.Oid
		tracerx.Printf("RETAIN: %v via ref %v", p.Oid, ref)
	})

	if err != nil {
		errorChan <- err
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetPreviousVersionsOfRef(gitscanner *lfs.GitScanner, ref string, since time.Time, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup, sem *semaphore.Weighted) {
	sem.Acquire(context.Background(), 1)
	defer sem.Release(1)
	defer waitg.Done()

	err := gitscanner.ScanPreviousVersions(ref, since, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			errorChan <- err
			return
		}

		retainChan <- p.Oid
		tracerx.Printf("RETAIN: %v via ref %v >= %v", p.Oid, ref, since)
	})

	if err != nil {
		errorChan <- err
		return
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedCurrentAndRecentRefs(gitscanner *lfs.GitScanner, fetchconf lfs.FetchPruneConfig, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup, sem *semaphore.Weighted) {
	defer waitg.Done()

	// We actually increment the waitg in this func since we kick off sub-goroutines
	// Make a list of what unique commits to keep, & search backward from
	commits := tools.NewStringSet()
	// Do current first
	ref, err := git.CurrentRef()
	if err != nil {
		errorChan <- err
		return
	}
	commits.Add(ref.Sha)
	if !fetchconf.PruneForce {
		waitg.Add(1)
		go pruneTaskGetRetainedAtRef(gitscanner, ref.Sha, retainChan, errorChan, waitg, sem)
	}

	// Now recent
	if !fetchconf.PruneRecent && fetchconf.FetchRecentRefsDays > 0 {
		pruneRefDays := fetchconf.FetchRecentRefsDays + fetchconf.PruneOffsetDays
		tracerx.Printf("PRUNE: Retaining non-HEAD refs within %d (%d+%d) days", pruneRefDays, fetchconf.FetchRecentRefsDays, fetchconf.PruneOffsetDays)
		refsSince := time.Now().AddDate(0, 0, -pruneRefDays)
		// Keep all recent refs including any recent remote branches
		refs, err := git.RecentBranches(refsSince, fetchconf.FetchRecentRefsIncludeRemotes, "")
		if err != nil {
			Panic(err, tr.Tr.Get("Could not scan for recent refs"))
		}
		for _, ref := range refs {
			if commits.Add(ref.Sha) {
				// A new commit
				waitg.Add(1)
				go pruneTaskGetRetainedAtRef(gitscanner, ref.Sha, retainChan, errorChan, waitg, sem)
			}
		}
	}

	// For every unique commit we've fetched, check recent commits too
	// Only if we're fetching recent commits, otherwise only keep at refs
	if !fetchconf.PruneRecent && fetchconf.FetchRecentCommitsDays > 0 {
		pruneCommitDays := fetchconf.FetchRecentCommitsDays + fetchconf.PruneOffsetDays
		for commit := range commits.Iter() {
			// We measure from the last commit at the ref
			summ, err := git.GetCommitSummary(commit)
			if err != nil {
				errorChan <- errors.New(tr.Tr.Get("couldn't scan commits at %v: %v", commit, err))
				continue
			}
			commitsSince := summ.CommitDate.AddDate(0, 0, -pruneCommitDays)
			waitg.Add(1)
			go pruneTaskGetPreviousVersionsOfRef(gitscanner, commit, commitsSince, retainChan, errorChan, waitg, sem)
		}
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedUnpushed(gitscanner *lfs.GitScanner, fetchconf lfs.FetchPruneConfig, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup, sem *semaphore.Weighted) {
	defer waitg.Done()

	err := gitscanner.ScanUnpushed(fetchconf.PruneRemoteName, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			errorChan <- err
		} else {
			retainChan <- p.Pointer.Oid
			tracerx.Printf("RETAIN: %v unpushed", p.Pointer.Oid)
		}
	})

	if err != nil {
		errorChan <- err
		return
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedWorktree(gitscanner *lfs.GitScanner, fetchconf lfs.FetchPruneConfig, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup, sem *semaphore.Weighted) {
	defer waitg.Done()

	// Retain other worktree HEADs too
	// Working copy, branch & maybe commit is different but repo is shared
	allWorktrees, err := git.GetAllWorkTrees(cfg.LocalGitStorageDir())
	if err != nil {
		errorChan <- err
		return
	}
	// Don't repeat any commits, worktrees are always on their own branches but
	// may point to the same commit
	commits := tools.NewStringSet()

	if !fetchconf.PruneForce {
		// current HEAD is done elsewhere
		headref, err := git.CurrentRef()
		if err != nil {
			errorChan <- err
			return
		}

		commits.Add(headref.Sha)
	}

	for _, worktree := range allWorktrees {
		if !fetchconf.PruneForce && commits.Add(worktree.Ref.Sha) {
			// Worktree is on a different commit
			waitg.Add(1)
			// Don't need to 'cd' to worktree since we share same repo
			go pruneTaskGetRetainedAtRef(gitscanner, worktree.Ref.Sha, retainChan, errorChan, waitg, sem)
		}

		// Always scan the index of the worktree
		waitg.Add(1)
		go pruneTaskGetRetainedIndex(gitscanner, worktree.Ref.Sha, worktree.Dir, retainChan, errorChan, waitg, sem)
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedStashed(gitscanner *lfs.GitScanner, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup, sem *semaphore.Weighted) {
	defer waitg.Done()

	err := gitscanner.ScanStashed(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			errorChan <- err
		} else {
			retainChan <- p.Pointer.Oid
			tracerx.Printf("RETAIN: %v stashed", p.Pointer.Oid)
		}
	})

	if err != nil {
		errorChan <- err
		return
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedIndex(gitscanner *lfs.GitScanner, ref string, workingDir string, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup, sem *semaphore.Weighted) {
	defer waitg.Done()

	err := gitscanner.ScanIndex(ref, workingDir, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			errorChan <- err
		} else {
			retainChan <- p.Pointer.Oid
			tracerx.Printf("RETAIN: %v index", p.Pointer.Oid)
		}
	})

	if err != nil {
		errorChan <- err
		return
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetReachableObjects(gitscanner *lfs.GitScanner, outObjectSet *tools.StringSet, errorChan chan error, waitg *sync.WaitGroup, sem *semaphore.Weighted) {
	defer waitg.Done()

	err := gitscanner.ScanAll(func(p *lfs.WrappedPointer, err error) {
		sem.Acquire(context.Background(), 1)
		defer sem.Release(1)

		if err != nil {
			errorChan <- err
			return
		}
		outObjectSet.Add(p.Oid)
	})

	if err != nil {
		errorChan <- err
	}
}

func init() {
	RegisterCommand("prune", pruneCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&pruneDryRunArg, "dry-run", "d", false, "Don't delete anything, just report")
		cmd.Flags().BoolVarP(&pruneVerboseArg, "verbose", "v", false, "Print full details of what is/would be deleted")
		cmd.Flags().BoolVarP(&pruneRecentArg, "recent", "", false, "Prune even recent objects")
		cmd.Flags().BoolVarP(&pruneForceArg, "force", "f", false, "Prune everything that has been pushed")
		cmd.Flags().BoolVarP(&pruneVerifyArg, "verify-remote", "c", false, "Verify that remote has reachable LFS files before deleting")
		cmd.Flags().BoolVar(&pruneDoNotVerifyArg, "no-verify-remote", false, "Override lfs.pruneverifyremotealways and don't verify")
		cmd.Flags().BoolVar(&pruneVerifyUnreachableArg, "verify-unreachable", false, "When using --verify-remote, additionally verify unreachable LFS files before deleting.")
		cmd.Flags().BoolVar(&pruneDoNotVerifyUnreachableArg, "no-verify-unreachable", false, "Override lfs.pruneverifyunreachablealways and don't verify unreachable objects")
		cmd.Flags().StringVar(&pruneWhenUnverifiedArg, "when-unverified", "halt", "halt|continue the execution when objects are not found on the remote")
	})
}
