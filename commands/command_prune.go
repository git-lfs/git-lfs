package commands

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/fs"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tasklog"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tools/humanize"
	"github.com/git-lfs/git-lfs/tq"
	"github.com/rubyist/tracerx"
	"github.com/spf13/cobra"
)

var (
	pruneDryRunArg      bool
	pruneVerboseArg     bool
	pruneVerifyArg      bool
	pruneDoNotVerifyArg bool
)

func pruneCommand(cmd *cobra.Command, args []string) {
	// Guts of this must be re-usable from fetch --prune so just parse & dispatch
	if pruneVerifyArg && pruneDoNotVerifyArg {
		Exit("Cannot specify both --verify-remote and --no-verify-remote")
	}

	fetchPruneConfig := lfs.NewFetchPruneConfig(cfg.Git)
	verify := !pruneDoNotVerifyArg &&
		(fetchPruneConfig.PruneVerifyRemoteAlways || pruneVerifyArg)
	prune(fetchPruneConfig, verify, pruneDryRunArg, pruneVerboseArg)
}

type PruneProgressType int

const (
	PruneProgressTypeLocal  = PruneProgressType(iota)
	PruneProgressTypeRetain = PruneProgressType(iota)
	PruneProgressTypeVerify = PruneProgressType(iota)
)

// Progress from a sub-task of prune
type PruneProgress struct {
	ProgressType PruneProgressType
	Count        int // Number of items done
}
type PruneProgressChan chan PruneProgress

func prune(fetchPruneConfig lfs.FetchPruneConfig, verifyRemote, dryRun, verbose bool) {
	localObjects := make([]fs.Object, 0, 100)
	retainedObjects := tools.NewStringSetWithCapacity(100)

	logger := tasklog.NewLogger(OutputWriter)
	defer logger.Close()

	var reachableObjects tools.StringSet
	var taskwait sync.WaitGroup

	// Add all the base funcs to the waitgroup before starting them, in case
	// one completes really fast & hits 0 unexpectedly
	// each main process can Add() to the wg itself if it subdivides the task
	taskwait.Add(4) // 1..4: localObjects, current & recent refs, unpushed, worktree
	if verifyRemote {
		taskwait.Add(1) // 5
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

	gitscanner := lfs.NewGitScanner(nil)
	gitscanner.Filter = filepathfilter.New(nil, cfg.FetchExcludePaths())

	go pruneTaskGetRetainedCurrentAndRecentRefs(gitscanner, fetchPruneConfig, retainChan, errorChan, &taskwait)
	go pruneTaskGetRetainedUnpushed(gitscanner, fetchPruneConfig, retainChan, errorChan, &taskwait)
	go pruneTaskGetRetainedWorktree(gitscanner, retainChan, errorChan, &taskwait)
	if verifyRemote {
		reachableObjects = tools.NewStringSetWithCapacity(100)
		go pruneTaskGetReachableObjects(gitscanner, &reachableObjects, errorChan, &taskwait)
	}

	// Now collect all the retained objects, on separate wait
	var retainwait sync.WaitGroup
	retainwait.Add(1)
	go pruneTaskCollectRetained(&retainedObjects, retainChan, progressChan, &retainwait)

	// Report progress
	var progresswait sync.WaitGroup
	progresswait.Add(1)
	go pruneTaskDisplayProgress(progressChan, &progresswait, logger)

	taskwait.Wait() // wait for subtasks
	gitscanner.Close()
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
				tracerx.Printf("VERIFYING: %v", file.Oid)

				verifyQueue.Add(downloadTransfer(&lfs.WrappedPointer{
					Pointer: lfs.NewPointer(file.Oid, file.Size, nil),
				}))
			}
		}
	}

	if verifyRemote {
		verifyQueue.Wait()
		verifywait.Wait()
		close(progressChan) // after verify but before check
		progresswait.Wait()
		pruneCheckVerified(prunableObjects, reachableObjects, verifiedObjects)
	} else {
		close(progressChan)
		progresswait.Wait()
	}

	if len(prunableObjects) == 0 {
		return
	}

	info := tasklog.NewSimpleTask()
	logger.Enqueue(info)
	if dryRun {
		info.Logf("prune: %d file(s) would be pruned (%s)", len(prunableObjects), humanize.FormatBytes(uint64(totalSize)))
		for _, item := range verboseOutput {
			info.Logf("\n * %s", item)
		}
		info.Complete()
	} else {
		for _, item := range verboseOutput {
			info.Logf("\n%s", item)
		}
		info.Complete()

		pruneDeleteFiles(prunableObjects, logger)
	}
}

func pruneCheckVerified(prunableObjects []string, reachableObjects, verifiedObjects tools.StringSet) {
	// There's no issue if an object is not reachable and missing, only if reachable & missing
	var problems bytes.Buffer
	for _, oid := range prunableObjects {
		// Test verified first as most likely reachable
		if !verifiedObjects.Contains(oid) {
			if reachableObjects.Contains(oid) {
				problems.WriteString(fmt.Sprintf(" * %v\n", oid))
			} else {
				// Just to indicate why it doesn't matter that we didn't verify
				tracerx.Printf("UNREACHABLE: %v", oid)
			}
		}
	}
	// technically we could still prune the other oids, but this indicates a
	// more serious issue because the local state implies that these can be
	// deleted but that's incorrect; bad state has occurred somehow, might need
	// push --all to resolve
	if problems.Len() > 0 {
		Exit("Abort: these objects to be pruned are missing on remote:\n%v", problems.String())
	}
}

func pruneCheckErrors(taskErrors []error) {
	if len(taskErrors) > 0 {
		for _, err := range taskErrors {
			LoggedError(err, "Prune error: %v", err)
		}
		Exit("Prune sub-tasks failed, cannot continue")
	}
}

func pruneTaskDisplayProgress(progressChan PruneProgressChan, waitg *sync.WaitGroup, logger *tasklog.Logger) {
	defer waitg.Done()

	task := tasklog.NewSimpleTask()
	defer task.Complete()

	logger.Enqueue(task)

	localCount := 0
	retainCount := 0
	verifyCount := 0
	var msg string
	for p := range progressChan {
		switch p.ProgressType {
		case PruneProgressTypeLocal:
			localCount++
		case PruneProgressTypeRetain:
			retainCount++
		case PruneProgressTypeVerify:
			verifyCount++
		}
		msg = fmt.Sprintf("prune: %d local object(s), %d retained", localCount, retainCount)
		if verifyCount > 0 {
			msg += fmt.Sprintf(", %d verified with remote", verifyCount)
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
	task := logger.Percentage("prune: Deleting objects", uint64(len(prunableObjects)))

	var problems bytes.Buffer
	// In case we fail to delete some
	var deletedFiles int
	for _, oid := range prunableObjects {
		mediaFile, err := cfg.Filesystem().ObjectPath(oid)
		if err != nil {
			problems.WriteString(fmt.Sprintf("Unable to find media path for %v: %v\n", oid, err))
			continue
		}
		err = os.Remove(mediaFile)
		if err != nil {
			problems.WriteString(fmt.Sprintf("Failed to remove file %v: %v\n", mediaFile, err))
			continue
		}
		deletedFiles++
		task.Count(1)
	}
	if problems.Len() > 0 {
		LoggedError(fmt.Errorf("Failed to delete some files"), problems.String())
		Exit("Prune failed, see errors above")
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
func pruneTaskGetRetainedAtRef(gitscanner *lfs.GitScanner, ref string, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup) {
	defer waitg.Done()

	err := gitscanner.ScanRef(ref, func(p *lfs.WrappedPointer, err error) {
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
func pruneTaskGetPreviousVersionsOfRef(gitscanner *lfs.GitScanner, ref string, since time.Time, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup) {
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
func pruneTaskGetRetainedCurrentAndRecentRefs(gitscanner *lfs.GitScanner, fetchconf lfs.FetchPruneConfig, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup) {
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
	waitg.Add(1)
	go pruneTaskGetRetainedAtRef(gitscanner, ref.Sha, retainChan, errorChan, waitg)

	// Now recent
	if fetchconf.FetchRecentRefsDays > 0 {
		pruneRefDays := fetchconf.FetchRecentRefsDays + fetchconf.PruneOffsetDays
		tracerx.Printf("PRUNE: Retaining non-HEAD refs within %d (%d+%d) days", pruneRefDays, fetchconf.FetchRecentRefsDays, fetchconf.PruneOffsetDays)
		refsSince := time.Now().AddDate(0, 0, -pruneRefDays)
		// Keep all recent refs including any recent remote branches
		refs, err := git.RecentBranches(refsSince, fetchconf.FetchRecentRefsIncludeRemotes, "")
		if err != nil {
			Panic(err, "Could not scan for recent refs")
		}
		for _, ref := range refs {
			if commits.Add(ref.Sha) {
				// A new commit
				waitg.Add(1)
				go pruneTaskGetRetainedAtRef(gitscanner, ref.Sha, retainChan, errorChan, waitg)
			}
		}
	}

	// For every unique commit we've fetched, check recent commits too
	// Only if we're fetching recent commits, otherwise only keep at refs
	if fetchconf.FetchRecentCommitsDays > 0 {
		pruneCommitDays := fetchconf.FetchRecentCommitsDays + fetchconf.PruneOffsetDays
		for commit := range commits.Iter() {
			// We measure from the last commit at the ref
			summ, err := git.GetCommitSummary(commit)
			if err != nil {
				errorChan <- fmt.Errorf("Couldn't scan commits at %v: %v", commit, err)
				continue
			}
			commitsSince := summ.CommitDate.AddDate(0, 0, -pruneCommitDays)
			waitg.Add(1)
			go pruneTaskGetPreviousVersionsOfRef(gitscanner, commit, commitsSince, retainChan, errorChan, waitg)
		}
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedUnpushed(gitscanner *lfs.GitScanner, fetchconf lfs.FetchPruneConfig, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup) {
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
func pruneTaskGetRetainedWorktree(gitscanner *lfs.GitScanner, retainChan chan string, errorChan chan error, waitg *sync.WaitGroup) {
	defer waitg.Done()

	// Retain other worktree HEADs too
	// Working copy, branch & maybe commit is different but repo is shared
	allWorktreeRefs, err := git.GetAllWorkTreeHEADs(cfg.LocalGitStorageDir())
	if err != nil {
		errorChan <- err
		return
	}
	// Don't repeat any commits, worktrees are always on their own branches but
	// may point to the same commit
	commits := tools.NewStringSet()
	// current HEAD is done elsewhere
	headref, err := git.CurrentRef()
	if err != nil {
		errorChan <- err
		return
	}
	commits.Add(headref.Sha)
	for _, ref := range allWorktreeRefs {
		if commits.Add(ref.Sha) {
			// Worktree is on a different commit
			waitg.Add(1)
			// Don't need to 'cd' to worktree since we share same repo
			go pruneTaskGetRetainedAtRef(gitscanner, ref.Sha, retainChan, errorChan, waitg)
		}
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetReachableObjects(gitscanner *lfs.GitScanner, outObjectSet *tools.StringSet, errorChan chan error, waitg *sync.WaitGroup) {
	defer waitg.Done()

	err := gitscanner.ScanAll(func(p *lfs.WrappedPointer, err error) {
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
		cmd.Flags().BoolVarP(&pruneVerifyArg, "verify-remote", "c", false, "Verify that remote has LFS files before deleting")
		cmd.Flags().BoolVar(&pruneDoNotVerifyArg, "no-verify-remote", false, "Override lfs.pruneverifyremotealways and don't verify")
	})
}
