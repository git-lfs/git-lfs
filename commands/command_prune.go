package commands

import (
	"sync"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	pruneCmd = &cobra.Command{
		Use:   "prune",
		Short: "Deletes old LFS files from the local store",
		Run:   pruneCommand,
	}
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

	verify := !pruneDoNotVerifyArg &&
		(lfs.Config.FetchPruneConfig().PruneVerifyRemoteAlways || pruneVerifyArg)

	prune(verify, pruneDryRunArg, pruneVerboseArg)

}

func prune(verifyRemote, dryRun, verbose bool) {
	localObjects := make([]string, 0, 100)
	retainedObjects := lfs.NewStringSetWithCapacity(100)
	var reachableObjects lfs.StringSet
	var taskwait sync.WaitGroup

	// Data collection algorithm summary
	// Separate goroutines must find:
	//  1. Local objects
	//  2. Current checkout, 2 sub-goroutines
	//     a. LFS files at checkout + index
	//     b. LFS files at recent commits on HEAD
	//  3. List of recent refs (unique SHA, not HEAD), pass to 2 sub-goroutines >
	//     a. LFS files at ref
	//     b. LFS files at recent commits of ref
	//  4. Unpushed objects
	//  5. Other worktree checkouts (1 sub-goroutine per other worktree)
	//  6. [if --verify-remote] Reachable objects

	// This main routine will collate the outputs of chans, report progress
	// with spinner of # objects stored, # objects referenced

	// localObjects:array = 1
	// retainedObjects:set = 2..5
	// [if --verify-remote] reachableObjects:set = 6

	// Add all the base funcs to the waitgroup before starting them, in case
	// one completes really fast & hits 0 unexpectedly
	// each main process can Add() to the wg itself if it subdivides the task
	taskwait.Add(5) // 1..5: localObjects, current checkout, recent refs, unpushed, worktree
	if verifyRemote {
		taskwait.Add(1) // 6
	}

	// Populate the single list of local objects
	go pruneTaskGetLocalObjects(&localObjects, &taskwait)

	// Now find files to be retained from many sources
	retainChan := make(chan string, 100)

	go pruneTaskGetRetainedCurrentCheckout(retainChan, &taskwait)
	go pruneTaskGetRetainedRecentRefs(retainChan, &taskwait)
	go pruneTaskGetRetainedUnpushed(retainChan, &taskwait)
	go pruneTaskGetRetainedWorktree(retainChan, &taskwait)
	if verifyRemote {
		reachableObjects = lfs.NewStringSetWithCapacity(100)
		go pruneTaskGetReachableObjects(&reachableObjects, &taskwait)
	}

	// Now collect all the retained objects
	var retainwait sync.WaitGroup
	retainwait.Add(1)
	go func() {
		for oid := range retainChan {
			retainedObjects.Add(oid)
		}
	}()
	taskwait.Wait()   // wait for subtasks
	close(retainChan) // triggers retain collector to end now all tasks have
	retainwait.Wait() // make sure all retained objects added

	// prunableObjects:set = (localObjects - retainedObjects)
	// [if --verify-remote] verifyObjects:set = (prunableObjects n reachableObjects)
	//    (this is so we don't try to verify unreachable objects on remote)
	// [if --verify-remote] call remote, remove unverified objects from prunableObjects & report

	// Report number & size of objects to delete
	// [if !dry-run] delete prunableObjects
	// [if verbose] report oids & individual sizes

}

// Background task, must call waitg.Done() once at end
func pruneTaskGetLocalObjects(outLocalObjects *[]string, waitg *sync.WaitGroup) {
	defer waitg.Done()

	localObjectsChan := lfs.AllLocalObjectsChan()
	for oid := range localObjectsChan {
		*outLocalObjects = append(*outLocalObjects, oid)
	}
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedCurrentCheckout(retainChan chan string, waitg *sync.WaitGroup) {
	defer waitg.Done()

	// TODO
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedRecentRefs(retainChan chan string, waitg *sync.WaitGroup) {
	defer waitg.Done()

	// TODO
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedUnpushed(retainChan chan string, waitg *sync.WaitGroup) {
	defer waitg.Done()

	// TODO
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetRetainedWorktree(retainChan chan string, waitg *sync.WaitGroup) {
	defer waitg.Done()

	// TODO
}

// Background task, must call waitg.Done() once at end
func pruneTaskGetReachableObjects(outObjectSet *lfs.StringSet, waitg *sync.WaitGroup) {
	defer waitg.Done()

	// TODO
}

func init() {
	pruneCmd.Flags().BoolVarP(&pruneDryRunArg, "dry-run", "d", false, "Don't delete anything, just report")
	pruneCmd.Flags().BoolVarP(&pruneDryRunArg, "verbose", "v", false, "Print full details of what is/would be deleted")
	pruneCmd.Flags().BoolVarP(&pruneDryRunArg, "verify-remote", "c", false, "Verify that remote has LFS files before deleting")
	pruneCmd.Flags().BoolVar(&pruneDryRunArg, "no-verify-remote", false, "Override lfs.pruneverifyremotealways and don't verify")
	RootCmd.AddCommand(pruneCmd)
}
